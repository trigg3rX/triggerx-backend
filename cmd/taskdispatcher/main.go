package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/tasks"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.TaskDispatcherProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Task Dispatcher service ...")

	// Initialize metrics collector
	collector := metrics.NewCollector()
	logger.Info("[1/5] Metrics collector Initialised")
	collector.Start()

	// Create Redis client and verify connection
	redisConfig := config.GetRedisClientConfig()
	redisClient, err := redis.NewRedisClient(logger, redisConfig)
	if err != nil {
		logger.Fatal("Failed to create Redis client", "error", err)
	}
	if err := redisClient.Ping(context.Background()); err != nil {
		logger.Fatal("Redis is not reachable", "error", err)
	}
	logger.Info("[2/5] Redis client Initialised")

	// Set up monitoring hooks for metrics integration
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	redisClient.SetMonitoringHooks(monitoringHooks)

	aggCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	}
	aggClient, err := aggregator.NewAggregatorClient(logger, aggCfg)
	if err != nil {
		logger.Fatal("Failed to create aggregator client", "error", err)
	}
	logger.Info("[3/5] Aggregator client Initialised")

	testAggCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetTestAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	}
	testAggClient, err := aggregator.NewAggregatorClient(logger, testAggCfg)
	if err != nil {
		logger.Fatal("Failed to create aggregator client", "error", err)
	}
	logger.Info("[3/5] Test Aggregator client Initialised")

	// Initialize gRPC client for health service
	healthClient, err := rpc.NewHealthClient(config.GetHealthRPCUrl(), logger)
	if err != nil {
		logger.Fatal("Failed to create health gRPC client", "error", err)
	}
	logger.Info("[4/5] Health gRPC client Initialised")

	// Initialize task stream manager for orchestration
	taskStreamMgr, err := tasks.NewTaskStreamManager(redisClient, aggClient, testAggClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize TaskStreamManager", "error", err)
	}
	logger.Info("[5/5] Task stream manager Initialised")

	// TaskDispatcher is the main orchestrator. It needs all the other components.
	dispatcher, err := taskdispatcher.NewTaskDispatcher(
		logger,
		taskStreamMgr,
		healthClient,
		config.GetTaskDispatcherSigningKey(),
		config.GetTaskDispatcherSigningAddress(),
	)
	if err != nil {
		logger.Fatal("Failed to initialize TaskDispatcher", "error", err)
	}
	logger.Info("Task Dispatcher Initialised")

	// 5. Initialize the delivery mechanism (RPC Server) using the generic approach
	serverConfig := rpcserver.Config{
		Name:    "TaskDispatcher",
		Version: "1.0.0",
		Address: "0.0.0.0",
		Port:    config.GetTaskDispatcherRPCPort(),
	}
	srv := rpcserver.NewServer(serverConfig, logger)
	srv.AddInterceptor(rpcserver.LoggingInterceptor(logger))

	// Create and register the generic RPC handler
	handler := rpc.NewTaskDispatcherHandler(logger, dispatcher)
	srv.RegisterHandler("TaskDispatcher", handler)

	// 6. Start everything
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		logger.Fatal("Failed to start RPC server", "error", err)
	}

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, srv, dispatcher, logger)
}

// performGracefulShutdown handles graceful shutdown of the service
func performGracefulShutdown(ctx context.Context, server *rpcserver.Server, dispatcher *taskdispatcher.TaskDispatcher, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	logger.Info("Shutting down RPC server...")
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("RPC server forced to shutdown", "error", err)
	} else {
		logger.Info("RPC server stopped successfully")
	}

	// Close the Dispatcher
	if err := dispatcher.Close(); err != nil {
		logger.Error("Failed to close dispatcher", "error", err)
	} else {
		logger.Info("Dispatcher closed successfully")
	}

	logger.Info("Task Dispatcher service shutdown complete")

	// Ensure we exit cleanly
	select {
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
