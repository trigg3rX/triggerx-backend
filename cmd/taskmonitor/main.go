package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/api"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/core/manager"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/taskmonitor.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TaskMonitorService,
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

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx := context.Background()
	logger.Info(ctx, "[1/4] Dependency: Observability Module Initialised")

	// Initialize TaskManager (handles Redis, Database, IPFS, Event Listener, and Task Stream Manager)
	taskManager, err := manager.NewTaskManager(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to create TaskManager", observability.Error(err))
	}

	// Initialize all components
	if err := taskManager.Initialize(); err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskManager components", observability.Error(err))
	}
	logger.Info(ctx, "[2/4] Dependency: TaskManager Initialised")

	// Setup API server with only /status endpoint
	apiSrv := api.NewServer(config.GetHTTPPort())
	logger.Info(ctx, "[3/4] Dependency: API Server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get database client for handler
	taskRepo := taskManager.GetTaskRepository()

	// Initialize gRPC server
	rpcDeps := &rpc.Dependencies{
		Monitor:  taskManager,
		DBClient: taskRepo,
	}
	rpcSrv, err := rpc.NewServer(logger, tracer, rpcDeps)
	if err != nil {
		logger.Fatal(ctx, "Failed to create gRPC server", observability.Error(err))
	}
	logger.Info(ctx, "[4/4] Dependency: gRPC Server Initialised")

	// Start metrics collector
	collector := metrics.NewCollector(obsMetrics, logger)
	collector.Start()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

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
	performGracefulShutdown(ctx, apiSrv, rpcSrv, taskManager, logger, obs)
}

func performGracefulShutdown(
	ctx context.Context,
	apiSrv *api.Server,
	rpcSrv *rpc.Server,
	taskManager *manager.TaskManager,
	logger observability.Logger,
	obs *observability.Observability,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Shutdown gRPC server
		if err := rpcSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "gRPC server shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[1/4] Shutdown: gRPC Server Stopped")

		// Shutdown API server gracefully
		if err := apiSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "API server shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[2/4] Shutdown: API Server Stopped")

		// Close TaskManager (handles all components)
		if err := taskManager.Close(); err != nil {
			logger.Warn(shutdownCtx, "Error during TaskManager shutdown", observability.Error(err))
		}
		logger.Info(ctx, "[3/4] Shutdown: TaskManager Closed")

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
