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
	defer obs.Shutdown(context.Background())

	// Extract individual components
	obsLogger := obs.Logger()
	// obsTracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Use observability logger for initial startup log
	ctx := context.Background()
	obsLogger.Info(ctx, "Starting health service...")

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
	dbConn, err := database.NewConnection(dbConfig, obsLogger)
	if err != nil {
		obsLogger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}

	// Initialize Telegram bot
	telegramBot, err := telegram.NewBot(config.GetBotToken(), obsLogger, dbConn)
	if err != nil {
		obsLogger.Warn(ctx, "Failed to initialize Telegram bot", observability.Error(err))
	}

	// Initialize database manager
	client.InitDatabaseManager(ctx, obsLogger, dbConn, telegramBot)
	obsLogger.Info(ctx, "Database manager initialized")

	// Initialize state manager
	stateManager := keeper.InitializeStateManager(ctx, obsLogger)
	obsLogger.Info(ctx, "Keeper state manager initialized")

	// Initialize metrics using observability metrics
	metrics.InitializeMetrics(obsMetrics)
	obsLogger.Info(ctx, "Metrics initialized")

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(ctx); err != nil {
		obsLogger.Debug(ctx, "Failed to load verified keepers from database", observability.Error(err))
		// Continue anyway, as we can still operate with an empty state
	}

	// Setup HTTP server
	srv := setupHTTPServer(obsLogger)

	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		obsLogger.Info(ctx, "Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	obsLogger.Info(ctx, "Health service is ready",
		observability.String("port", config.GetHealthRPCPort()),
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		obsLogger.Error(ctx, "Server error received", observability.Error(err))
	case sig := <-shutdown:
		obsLogger.Info(ctx, "Received shutdown signal",
			observability.String("signal", sig.String()),
		)
	}

	performGracefulShutdown(ctx, srv, &wg, obs, obsLogger, stateManager)
}

func setupHTTPServer(logger observability.Logger) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(health.LoggerMiddleware(logger))

	// Register routes
	health.RegisterRoutes(router, logger)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetHealthRPCPort()),
		Handler: router,
	}
}

// setupHTTPServerMinimal creates an HTTP server with observability components
// This is a temporary solution until health package is migrated to observability
func setupHTTPServerMinimal(logger observability.Logger, tracer observability.Tracer, metrics observability.Metrics) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	// TODO: Add observability middleware when health package is migrated

	// Register basic routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "TriggerX Health Service",
			"status":    "running",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

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
	logger.Info(ctx, "Initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

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

	// Shutdown observability (logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
	}

	logger.Info(shutdownCtx, "Shutdown complete")
}
