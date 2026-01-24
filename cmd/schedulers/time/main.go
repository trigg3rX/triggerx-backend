package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/core/scheduler"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/database"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/rpc/clients/taskdispatcher"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/time-scheduler.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TimeSchedulerService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		config.GetServiceID(),
	)

	// Initialize observability (all three pillars)
	obs, err := observability.Initialize(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx := context.Background()
	logger.Info(ctx, "[1/7] Dependency: Observability Module Initialised")

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	defer dbConn.Close()
	logger.Info(ctx, "[2/7] Dependency: Database Connection Initialised")

	// Initialize repositories
	timeJobRepo := repository.NewTimeJobRepository(dbConn)
	scriptStorageRepo := repository.NewScriptStorageRepository(dbConn)
	taskRepo := repository.NewTaskRepository(dbConn)
	logger.Info(ctx, "[3/7] Dependency: Repositories Initialised")

	// Initialize Redis client (optional, for polling state recovery)
	var redisClient *redis.Client
	if config.GetUpstashRedisUrl() != "" {
		redisClient, err = redis.NewClient(logger)
		if err != nil {
			logger.Warn(ctx, "Failed to initialize Redis client, continuing without it", observability.Error(err))
			redisClient = nil
		} else {
			logger.Info(ctx, "[4/7] Dependency: Redis Client Initialised")
		}
	} else {
		logger.Info(ctx, "[4/7] Dependency: Redis Client Skipped (no URL configured)")
	}

	// Initialize task dispatcher RPC client
	taskDispatcherClient, err := taskdispatcher.NewClient(
		config.GetTaskDispatcherRPCUrl(),
		logger,
		tracer,
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize task dispatcher client", observability.Error(err))
	}
	logger.Info(ctx, "[5/7] Dependency: Task Dispatcher Client Initialised")

	// Initialize time-based scheduler (includes both traditional TDI 1,2 and agent TDI 7 jobs)
	timeScheduler, err := scheduler.NewTimeBasedScheduler(logger, tracer, obsMetrics, timeJobRepo, scriptStorageRepo, taskRepo, taskDispatcherClient)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize time-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[6/7] Dependency: Time Scheduler Initialised")

	// Keep reference to redisClient for potential future use
	_ = redisClient

	// Setup HTTP server with only status endpoint
	srv := api.NewServer(config.GetHTTPPort(), logger)
	logger.Info(ctx, "[7/7] Dependency: API Server Initialised")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrics collection is started in NewTimeBasedScheduler
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start scheduler in background
	go func() {
		timeScheduler.Start(ctx)
	}()
	logger.Info(ctx, "[2/3] Process: Polling Jobs Started")

	// Start HTTP server
	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: API Server Started", observability.String("port", config.GetHTTPPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, cancel, srv, timeScheduler, obs, logger)
}

func performGracefulShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	apiSrv *api.Server,
	timeScheduler *scheduler.TimeBasedScheduler,
	obs *observability.Observability,
	logger observability.Logger,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GetShutdownTimeout())
	defer shutdownCancel()

	// Cancel context to stop scheduler and API server
	cancel()

	// Stop scheduler gracefully (this also closes the RPC client)
	timeScheduler.Stop(shutdownCtx)
	logger.Info(ctx, "[1/3] Shutdown: Scheduler Stopped")

	// Shutdown API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "API server shutdown error", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Shutdown: API Server Stopped")

	logger.Info(ctx, "Graceful shutdown completed successfully")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
	}
	logger.Info(ctx, "[3/3] Shutdown: Observability Shutdown Complete")
}
