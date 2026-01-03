package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/api"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
)

func main() {
	// Initialize configuration
	configPath := "config/services/task-monitor.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TaskMonitorService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
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
	logger.Info(ctx, "[1/3] Dependency: Observability Module Initialised")

	// Initialize TaskManager (handles Redis, Database, IPFS, Event Listener, and Task Stream Manager)
	taskManager, err := taskmonitor.NewTaskManager(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to create TaskManager", observability.Error(err))
	}

	// Initialize all components
	if err := taskManager.Initialize(); err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskManager components", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Dependency: TaskManager Initialised")

	// Setup API server with only /status endpoint
	apiSrv := api.NewServer(config.GetHTTPPort())
	logger.Info(ctx, "[3/4] Dependency: API Server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize and start gRPC server
	rpcServer, err := rpc.StartRPCServer(ctx, logger, taskManager, "0.0.0.0", config.GetGRPCPort())
	if err != nil {
		logger.Fatal(ctx, "Failed to start gRPC server", observability.Error(err))
	}
	logger.Info(ctx, "[4/4] Dependency: RPC Server Initialised")

	// Start metrics collector
	collector := metrics.NewCollector(obsMetrics)
	collector.Start()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start API server
	go func() {
		if err := apiSrv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[2/3] Process: API Server Started")
	logger.Info(ctx, "[3/3] Process: RPC Server Started")

	// Store RPC server in TaskManager for graceful shutdown
	taskManager.SetRPCServer(rpcServer)

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until signal is received
	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	// Perform graceful shutdown
	performGracefulShutdown(ctx, apiSrv, taskManager, rpcServer, logger, obs)
}

func performGracefulShutdown(
	ctx context.Context,
	apiSrv *api.Server,
	taskManager *taskmonitor.TaskManager,
	rpcServer *rpcserver.Server,
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

		// Shutdown API server gracefully
		if err := apiSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "API server forced to shutdown", observability.Error(err))
		}

		// Shutdown gRPC server gracefully
		if err := rpcServer.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "RPC server forced to shutdown", observability.Error(err))
		}

		// Close TaskManager (handles all components)
		if err := taskManager.Close(); err != nil {
			logger.Warn(shutdownCtx, "Error during TaskManager shutdown", observability.Error(err))
		}

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (handles logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down observability", observability.Error(err))
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
