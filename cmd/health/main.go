package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/health"
	"github.com/trigg3rX/triggerx-backend/internal/health/client"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Create observability configuration
	obsConfig := observability.NewConfig(
		observability.HealthService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)

	// Initialize observability (all three pillars: logger, tracer, metrics)
	obs, err := observability.Initialize(obsConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	// Extract individual components
	logger := obs.Logger()
	obsTracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Use observability logger for initial startup log
	ctx := context.Background()
	logger.Info(ctx, "Starting health service...")
	logger.Info(ctx, "[1/6] Dependency: Observability Module Initialised")

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)

	// Initialize database connection
	dbConfig := &database.Config{
		Hosts:        []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:     "triggerx",
		Consistency:  gocql.Quorum,
		Timeout:      time.Second * 30,
		Retries:      5,
		ConnectWait:  time.Second * 10,
		ProtoVersion: 4,
	}
	dbConn, err := database.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	logger.Info(ctx, "[2/6] Dependency: Database Connection Initialised")

	// Initialize Telegram bot
	telegramBot, err := telegram.NewBot(config.GetBotToken(), logger, dbConn)
	if err != nil {
		logger.Warn(ctx, "Failed to initialize Telegram bot", observability.Error(err))
	}
	logger.Info(ctx, "[3/6] Dependency: Telegram Bot Initialised")

	// Initialize database manager
	client.InitDatabaseManager(ctx, logger, obsTracer, dbConn, telegramBot)
	logger.Info(ctx, "[4/6] Dependency: Database Manager Initialised")

	// Initialize state manager
	stateManager := keeper.InitializeStateManager(ctx, logger, obsTracer)
	logger.Info(ctx, "[5/6] Dependency: Keeper State Manager Initialised")

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(ctx); err != nil {
		logger.Debug(ctx, "Failed to load verified keepers from database", observability.Error(err))
		// Continue anyway, as we can still operate with an empty state
	}

	// Setup HTTP server with tracing
	srv := setupHTTPServer(logger, obsTracer)
	logger.Info(ctx, "[6/6] Dependency: API Server Initialised")

	// Initialize metrics using observability metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/2] Process: Metrics Collector Started")

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()
	logger.Info(ctx, "[1/1] Process: HTTP Server Started")

	// TODO: When adding gRPC server, use the tracing interceptor:
	// import "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
	// grpcServer := grpc.NewServer(
	//     grpc.UnaryInterceptor(tracing.TraceInterceptor(obsTracer, "health")),
	// )

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error(ctx, "Error during HTTP server shutdown", observability.Error(err))
	case sig := <-shutdown:
		logger.Info(ctx, "Received shutdown signal",
			observability.String("signal", sig.String()),
		)
	}

	performGracefulShutdown(ctx, srv, &wg, obs, logger, stateManager)
}

func setupHTTPServer(logger observability.Logger, tracer observability.Tracer) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Add tracing middleware before logging middleware to ensure trace context is available
	router.Use(health.TraceMiddleware(tracer))
	router.Use(health.LoggerMiddleware(logger))

	// Register routes
	health.RegisterRoutes(router, logger)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetHealthRPCPort()),
		Handler: router,
	}
}

func performGracefulShutdown(
	ctx context.Context,
	srv *http.Server,
	wg *sync.WaitGroup,
	obs *observability.Observability,
	logger observability.Logger,
	stateManager *keeper.StateManager,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Update all keepers to inactive in database
		if stateManager != nil {
			if err := stateManager.DumpState(ctx); err != nil {
				logger.Error(shutdownCtx, "Failed to dump keeper state", observability.Error(err))
			}
		}

		// Shutdown HTTP server
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "HTTP server shutdown error", observability.Error(err))
			if err := srv.Close(); err != nil {
				logger.Error(shutdownCtx, "Forced HTTP server close error", observability.Error(err))
			}
		}

		// Wait for all goroutines to finish
		wg.Wait()

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
		}
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		// Shutdown completed successfully
	case <-shutdownCtx.Done():
		logger.Warn(ctx, "Shutdown timeout reached, forcing exit")
	}
	os.Exit(0)
}
