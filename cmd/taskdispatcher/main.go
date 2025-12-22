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
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	rpctracing "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TaskDispatcherService,
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
	logger.Info(ctx, "Starting Task Dispatcher service ...")

	// Initialize metrics
	obsMetrics, metricsShutdown, err := observability.NewMetrics(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize metrics: %v", err))
	}

	// Initialize metrics collector
	metrics.InitializeMetrics(obsMetrics)
	collector := metrics.NewCollector(obsMetrics)
	logger.Info(ctx, "[1/5] Metrics collector Initialised")
	collector.Start()

	// Create Redis client and verify connection
	redisConfig := config.GetRedisClientConfig()
	redisClient, err := redis.NewRedisClient(ctx, logger, redisConfig)
	if err != nil {
		logger.Fatal(ctx, "Failed to create Redis client", observability.Error(err))
	}
	if err := redisClient.Ping(ctx); err != nil {
		logger.Fatal(ctx, "Redis is not reachable", observability.Error(err))
	}
	logger.Info(ctx, "[2/5] Redis client Initialised")

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
		logger.Fatal(ctx, "Failed to create aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[3/5] Aggregator client Initialised")

	testAggCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetTestAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	}
	testAggClient, err := aggregator.NewAggregatorClient(logger, testAggCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to create aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[3/5] Test Aggregator client Initialised")

	healthClient := taskdispatcher.NewHealthClient(logger, config.GetHealthRPCUrl())
	logger.Info(ctx, "[4/5] Health client Initialised")

	// Initialize task stream manager for orchestration
	taskStreamMgr, err := tasks.NewTaskStreamManager(ctx, redisClient, aggClient, testAggClient, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskStreamManager", observability.Error(err))
	}
	logger.Info(ctx, "[5/5] Task stream manager Initialised")

	// TaskDispatcher is the main orchestrator. It needs all the other components.
	dispatcher, err := taskdispatcher.NewTaskDispatcher(
		logger,
		tracer,
		taskStreamMgr,
		healthClient,
		config.GetTaskDispatcherSigningKey(),
		config.GetTaskDispatcherSigningAddress(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskDispatcher", observability.Error(err))
	}
	logger.Info(ctx, "Task Dispatcher Initialised")

	// 5. Initialize the delivery mechanism (RPC Server) using the generic approach
	serverConfig := rpcserver.Config{
		Name:    "TaskDispatcher",
		Version: "1.0.0",
		Address: "0.0.0.0",
		Port:    config.GetTaskDispatcherRPCPort(),
	}
	srv := rpcserver.NewServer(serverConfig, logger)
	srv.AddInterceptor(rpcserver.LoggingInterceptor(logger))

	// Add trace interceptor for automatic trace context extraction
	tracingInterceptor := rpctracing.TraceInterceptor(tracer, "task-dispatcher")
	srv.AddInterceptor(tracingInterceptor)

	// Create and register the generic RPC handler
	handler := rpc.NewTaskDispatcherHandler(logger, dispatcher)
	srv.RegisterHandler("TaskDispatcher", handler)

	// 6. Start everything
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		logger.Fatal(ctx, "Failed to start RPC server", observability.Error(err))
	}

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, srv, dispatcher, logger, loggerShutdown, tracerShutdown, metricsShutdown)
}

// performGracefulShutdown handles graceful shutdown of the service
func performGracefulShutdown(ctx context.Context, server *rpcserver.Server, dispatcher *taskdispatcher.TaskDispatcher, logger observability.Logger, loggerShutdown func(context.Context) error, tracerShutdown func(context.Context) error, metricsShutdown func(context.Context) error) {
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	logger.Info(shutdownCtx, "Shutting down RPC server...")
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "RPC server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(shutdownCtx, "RPC server stopped successfully")
	}

	// Close the Dispatcher
	if err := dispatcher.Close(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Failed to close dispatcher", observability.Error(err))
	} else {
		logger.Info(shutdownCtx, "Dispatcher closed successfully")
	}

	// Shutdown tracer
	if tracerShutdown != nil {
		if err := tracerShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down tracer", observability.Error(err))
		}
	}

	// Shutdown metrics
	if metricsShutdown != nil {
		if err := metricsShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down metrics", observability.Error(err))
		}
	}

	// Shutdown logger
	if loggerShutdown != nil {
		if err := loggerShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
		}
	}

	logger.Info(shutdownCtx, "Task Dispatcher service shutdown complete")

	// Ensure we exit cleanly
	select {
	case <-shutdownCtx.Done():
		logger.Error(shutdownCtx, "Shutdown timeout exceeded")
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
