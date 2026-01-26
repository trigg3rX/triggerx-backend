package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/api"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/telegram"
	"github.com/trigg3rX/triggerx-backend/internal/health/database"
	"github.com/trigg3rX/triggerx-backend/internal/health/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/health/redis"
	"github.com/trigg3rX/triggerx-backend/internal/health/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/health.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Create observability configuration
	obsConfig := observability.NewConfig(
		observability.HealthService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		config.GetServiceID(),
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
	logger.Info(ctx, "[1/9] Dependency: Observability Module Initialised")

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	defer dbConn.Close()
	logger.Info(ctx, "[2/9] Dependency: Database Connection Initialised")

	// Initialize Telegram bot
	telegramBot, err := telegram.NewBot(config.GetBotToken(), logger, dbConn)
	if err != nil {
		logger.Warn(ctx, "Failed to initialize Telegram bot", observability.Error(err))
	}
	logger.Info(ctx, "[3/9] Dependency: Telegram Bot Initialised")

	// Initialize keeper repository
	keeperRepo := repository.NewKeeperRepository(dbConn, logger, obsTracer, telegramBot)
	logger.Info(ctx, "[4/9] Dependency: Keeper Repository Initialised")

	// Initialize Redis client
	redisClient, err := redis.NewClient(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize Redis client", observability.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error(ctx, "Failed to close Redis client", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[5/9] Dependency: Redis Client Initialised")

	// Initialize state manager
	stateManager := keeper.InitializeStateManager(ctx, logger, obsTracer, keeperRepo)
	logger.Info(ctx, "[6/9] Dependency: Keeper State Manager Initialised")

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(ctx); err != nil {
		logger.Debug(ctx, "Failed to load verified keepers from database", observability.Error(err))
		// Continue anyway, as we can still operate with an empty state
	}

	// Initialize performer selector
	performerSelector := keeper.NewPerformerSelector(stateManager, redisClient, logger)
	logger.Info(ctx, "[7/9] Dependency: Performer Selector Initialised")

	// Initialize API server
	apiCfg := api.Config{
		Port:           config.GetHTTPPort(),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	apiDeps := &api.Dependencies{
		Logger:       logger,
		Tracer:       obsTracer,
		Metrics:      obsMetrics,
		StateManager: stateManager,
	}

	apiSrv := api.NewServer(apiCfg, apiDeps)
	logger.Info(ctx, "[8/9] Dependency: HTTP API Server Initialised")

	// Setup gRPC server
	rpcSrv, err := rpc.NewServer(logger, obsTracer, stateManager, performerSelector)
	if err != nil {
		logger.Fatal(ctx, "Failed to create gRPC server", observability.Error(err))
	}
	logger.Info(ctx, "[9/9] Dependency: gRPC Server Initialised")

	// Initialize metrics using observability metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiSrv.Start(); err != nil {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()
	logger.Info(ctx, "[2/3] Process: HTTP Server Started", observability.String("port", config.GetHTTPPort()))

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rpcSrv.Start(ctx); err != nil {
			serverErrors <- fmt.Errorf("gRPC server error: %v", err)
		}
	}()
	logger.Info(ctx, "[3/3] Process: gRPC Server Started", observability.String("port", config.GetGRPCPort()))

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

	performGracefulShutdown(ctx, apiSrv, rpcSrv, &wg, obs, logger, stateManager)
}

func performGracefulShutdown(
	ctx context.Context,
	apiSrv *api.Server,
	rpcSrv *rpc.Server,
	wg *sync.WaitGroup,
	obs *observability.Observability,
	logger observability.Logger,
	stateManager *keeper.StateManager,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, config.GetShutdownTimeout())
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
		logger.Info(ctx, "[1/4] Shutdown: Keeper State Dumped")

		// Shutdown gRPC server
		if rpcSrv != nil {
			if err := rpcSrv.Stop(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "gRPC server shutdown error", observability.Error(err))
			}
		}
		logger.Info(ctx, "[2/4] Shutdown: gRPC Server Stopped")

		// Shutdown HTTP server
		if apiSrv != nil {
			if err := apiSrv.Stop(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "HTTP server shutdown error", observability.Error(err))
			}
		}
		logger.Info(ctx, "[3/4] Shutdown: HTTP Server Stopped")

		// Wait for all goroutines to finish
		wg.Wait()

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[4/4] Shutdown: Observability Shutdown Complete")
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
