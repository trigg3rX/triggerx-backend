package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/database"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/redis"
	conditionrpc "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/rpc/clients/eventmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/rpc/clients/taskdispatcher"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/condition-scheduler.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.ConditionSchedulerService,
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
	logger.Info(ctx, "[1/9] Dependency: Observability Module Initialised")

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	defer dbConn.Close()
	logger.Info(ctx, "[2/9] Dependency: Database Connection Initialised")

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(dbConn)
	jobRepo := repository.NewJobRepository(dbConn)
	logger.Info(ctx, "[3/9] Dependency: Repositories Initialised")

	// Initialize Redis client (optional, for job state caching)
	var redisClient *redis.Client
	if config.GetUpstashRedisUrl() != "" {
		redisClient, err = redis.NewClient(logger)
		if err != nil {
			logger.Warn(ctx, "Failed to initialize Redis client, continuing without it", observability.Error(err))
			redisClient = nil
		} else {
			logger.Info(ctx, "[4/9] Dependency: Redis Client Initialised")
		}
	} else {
		logger.Info(ctx, "[4/9] Dependency: Redis Client Skipped (no URL configured)")
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
	logger.Info(ctx, "[5/9] Dependency: Task Dispatcher Client Initialised")

	// Initialize event monitor RPC client
	eventMonitorClient, err := eventmonitor.NewClient(
		config.GetEventMonitorRPCUrl(),
		logger,
		tracer,
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize event monitor client", observability.Error(err))
	}
	logger.Info(ctx, "[6/9] Dependency: Event Monitor Client Initialised")

	// Initialize condition-based scheduler
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(
		logger,
		tracer,
		obsMetrics,
		taskRepo,
		jobRepo,
		taskDispatcherClient,
		eventMonitorClient,
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize condition-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[7/9] Dependency: Condition Scheduler Initialised")

	// Keep reference to redisClient for potential future use
	_ = redisClient

	// Setup API server (only /status endpoint and event webhook)
	apiPort := config.GetHTTPPort()
	apiSrv := api.NewServer(apiPort, logger, conditionScheduler)
	logger.Info(ctx, "[8/9] Dependency: API Server Initialised", observability.String("port", apiPort))

	// Setup RPC server (for schedule/unschedule operations)
	rpcDeps := &conditionrpc.Dependencies{
		Scheduler: conditionScheduler,
	}
	rpcSrv, err := conditionrpc.NewServer(logger, tracer, rpcDeps)
	if err != nil {
		logger.Fatal(ctx, "Failed to create RPC server", observability.Error(err))
	}
	logger.Info(ctx, "[9/9] Dependency: RPC Server Initialised")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrics collection is started in NewConditionBasedScheduler
	logger.Info(ctx, "[1/4] Process: Metrics Collector Started")

	// Start scheduler in background
	go func() {
		conditionScheduler.Start(ctx)
	}()
	logger.Info(ctx, "[2/4] Process: Scheduler Background Job Started")

	// Start API server
	go func() {
		if err := apiSrv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/4] Process: API Server Started", observability.String("port", apiPort))

	// Start RPC server
	go func() {
		if err := rpcSrv.Start(ctx); err != nil {
			logger.Error(ctx, "RPC server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[4/4] Process: RPC Server Started", observability.String("port", config.GetGRPCPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, cancel, apiSrv, rpcSrv, conditionScheduler, obs, logger)
}

func performGracefulShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	apiSrv *api.Server,
	rpcSrv *conditionrpc.Server,
	conditionScheduler *scheduler.ConditionBasedScheduler,
	obs *observability.Observability,
	logger observability.Logger,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GetShutdownTimeout())
	defer shutdownCancel()

	// Cancel context to stop scheduler
	cancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop(shutdownCtx)
	logger.Info(ctx, "[1/4] Shutdown: Scheduler Stopped")

	// Stop RPC server gracefully
	if err := rpcSrv.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "RPC server shutdown error", observability.Error(err))
	}
	logger.Info(ctx, "[2/4] Shutdown: RPC Server Stopped")

	// Stop API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "API server shutdown error", observability.Error(err))
	}
	logger.Info(ctx, "[3/4] Shutdown: API Server Stopped")

	logger.Info(ctx, "Graceful shutdown completed successfully")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
	}
	logger.Info(ctx, "[4/4] Shutdown: Observability Shutdown Complete")
}
