package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/api"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/core/dispatcher"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/redis"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/rpc/clients/health"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/taskdispatcher.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TaskDispatcherService,
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
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	ctx := context.Background()
	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/12] Dependency: Observability Module Initialised")

	// Create Redis client using the local wrapper
	redisClient, err := redis.NewClient(ctx, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to create Redis client", observability.Error(err))
	}
	logger.Info(ctx, "[2/12] Dependency: Redis Client Initialised")

	// Set up monitoring hooks for metrics integration
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	redisClient.SetMonitoringHooks(monitoringHooks)

	// Create aggregator clients using local wrappers
	aggClient, err := aggregator.NewAggregatorClient(logger, aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	})
	if err != nil {
		logger.Fatal(ctx, "Failed to create aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[3/12] Dependency: Aggregator Client Initialised")

	testAggClient, err := aggregator.NewAggregatorClient(logger, aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetTestAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	})
	if err != nil {
		logger.Fatal(ctx, "Failed to create test aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[4/12] Dependency: Test Aggregator Client Initialised")

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	defer dbConn.Close()
	logger.Info(ctx, "[5/12] Dependency: Database Connection Initialised")

	// Initialize task repository
	taskRepo := repository.NewTaskRepository(dbConn, logger)
	logger.Info(ctx, "[6/12] Dependency: Task Repository Initialised")

	// Create health client for performer selection
	healthClient, err := health.NewClient(config.GetHealthRPCUrl(), logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to create health client", observability.Error(err))
	}
	logger.Info(ctx, "[7/12] Dependency: Health Client Initialised")

	// Create performer fetcher
	performerFetcher := dispatcher.NewPerformerFetcher(healthClient, logger)
	logger.Info(ctx, "[8/12] Dependency: Performer Fetcher Initialised")

	// Initialize task stream manager for orchestration
	taskStreamMgr, err := dispatcher.NewTaskStreamManager(ctx, redisClient, taskRepo, aggClient, testAggClient, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskStreamManager", observability.Error(err))
	}
	logger.Info(ctx, "[9/12] Dependency: Task Stream Manager Initialised")

	// TaskDispatcher is the main orchestrator. It needs all the other components.
	taskDispatcher, err := dispatcher.NewTaskDispatcher(
		logger,
		tracer,
		taskStreamMgr,
		performerFetcher,
		config.GetTaskDispatcherSigningKey(),
		config.GetTaskDispatcherSigningAddress(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskDispatcher", observability.Error(err))
	}
	logger.Info(ctx, "[10/12] Dependency: Task Dispatcher Initialised")

	// Setup API server with only /status endpoint
	apiSrv := api.NewServer(config.GetHTTPPort())
	logger.Info(ctx, "[11/12] Dependency: API Server Initialised")

	// Initialize gRPC server
	rpcDeps := &rpc.Dependencies{
		Dispatcher: taskDispatcher,
	}
	rpcSrv, err := rpc.NewServer(logger, tracer, rpcDeps)
	if err != nil {
		logger.Fatal(ctx, "Failed to create gRPC server", observability.Error(err))
	}
	logger.Info(ctx, "[12/12] Dependency: gRPC Server Initialised")

	// Initialize metrics collector
	collector := metrics.NewCollector(obsMetrics, logger)
	collector.Start()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start API server
	go func() {
		if err := apiSrv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[2/3] Process: API Server Started", observability.String("port", config.GetHTTPPort()))

	// Start gRPC server
	go func() {
		if err := rpcSrv.Start(ctx); err != nil {
			logger.Error(ctx, "gRPC server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: gRPC Server Started", observability.String("port", config.GetGRPCPort()))

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until signal is received
	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	// Perform graceful shutdown
	performGracefulShutdown(ctx, apiSrv, rpcSrv, taskDispatcher, logger, obs)
}

// performGracefulShutdown handles graceful shutdown of the service
func performGracefulShutdown(ctx context.Context, apiSrv *api.Server, rpcSrv *rpc.Server, taskDispatcher *dispatcher.TaskDispatcher, logger observability.Logger, obs *observability.Observability) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, config.GetShutdownTimeout())
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Close the Dispatcher
		if err := taskDispatcher.Close(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Failed to close dispatcher", observability.Error(err))
		}
		logger.Info(ctx, "[1/4] Shutdown: Task Dispatcher Closed")

		// Shutdown gRPC server
		if err := rpcSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "gRPC server shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[2/4] Shutdown: gRPC Server Stopped")

		// Shutdown API server gracefully
		if err := apiSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "API server shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[3/4] Shutdown: API Server Stopped")

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (handles logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down observability", observability.Error(err))
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
