package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TaskMonitorService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)

	// Create resource for observability
	res, err := observability.NewResource(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create observability resource: %v", err))
	}

	// Initialize logger
	logger, loggerShutdown, err := observability.NewLogger(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Initialize tracer
	tracer, tracerShutdown, err := observability.NewTracer(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize tracer: %v", err))
	}

	ctx := context.Background()
	logger.Info(ctx, "Starting Task Monitor service ...")

	// Initialize TaskManager (handles Redis, Database, IPFS, Event Listener, and Task Stream Manager)
	taskManager, err := taskmonitor.NewTaskManager(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to create TaskManager", observability.Error(err))
	}
	logger.Info(ctx, "[1/6] TaskManager created successfully")

	// Initialize all components
	if err := taskManager.Initialize(); err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskManager components", observability.Error(err))
	}
	logger.Info(ctx, "[2/6] TaskManager components initialized successfully")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize and start gRPC server
	rpcServer, err := rpc.StartRPCServer(ctx, logger, taskManager, "0.0.0.0", config.GetTaskMonitorRPCPort())
	if err != nil {
		logger.Fatal(ctx, "Failed to start gRPC server", observability.Error(err))
	}
	logger.Info(ctx, "[3/6] gRPC server started successfully", observability.String("port", config.GetTaskMonitorRPCPort()))

	// Store RPC server in TaskManager for graceful shutdown
	taskManager.SetRPCServer(rpcServer)

	// Log service status
	logger.Info(ctx, "Task Monitor service is running")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, taskManager, rpcServer, logger, loggerShutdown, tracerShutdown)
}

func performGracefulShutdown(ctx context.Context, taskManager *taskmonitor.TaskManager, rpcServer *rpcserver.Server, logger observability.Logger, loggerShutdown func(context.Context) error, tracerShutdown func(context.Context) error) {
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Shutdown gRPC server gracefully
	logger.Info(shutdownCtx, "Shutting down gRPC server...")
	if err := rpcServer.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "RPC server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(shutdownCtx, "RPC server stopped successfully")
	}

	// Close TaskManager (handles all components)
	logger.Info(shutdownCtx, "Closing TaskManager...")
	if err := taskManager.Close(); err != nil {
		logger.Warn(shutdownCtx, "Non-critical errors during TaskManager shutdown", observability.Error(err))
	} else {
		logger.Info(shutdownCtx, "TaskManager closed successfully")
	}

	// Shutdown tracer
	if tracerShutdown != nil {
		if err := tracerShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down tracer", observability.Error(err))
		} else {
			logger.Info(shutdownCtx, "Tracer stopped successfully")
		}
	}

	// Shutdown logger
	if loggerShutdown != nil {
		if err := loggerShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
		} else {
			logger.Info(shutdownCtx, "Logger stopped successfully")
		}
	}

	logger.Info(shutdownCtx, "Task Monitor service shutdown complete")
}
