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

	// Initialize metrics
	obsMetrics, metricsShutdown, err := observability.NewMetrics(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize metrics: %v", err))
	}
	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/7] Dependency: Observability Module Initialised")

	// Create Redis client and verify connection
	redisConfig := config.GetRedisClientConfig()
	redisClient, err := redis.NewRedisClient(ctx, logger, redisConfig)
	if err != nil {
		logger.Fatal(ctx, "Failed to create Redis client", observability.Error(err))
	}
	if err := redisClient.Ping(ctx); err != nil {
		logger.Fatal(ctx, "Redis is not reachable", observability.Error(err))
	}
	logger.Info(ctx, "[2/7] Dependency: Redis Client Initialised")

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
	logger.Info(ctx, "[3/7] Dependency: Aggregator Client Initialised")

	testAggCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetTestAggregatorRPCUrl(),
		SenderPrivateKey: config.GetTaskDispatcherSigningKey(),
		SenderAddress:    config.GetTaskDispatcherSigningAddress(),
	}
	testAggClient, err := aggregator.NewAggregatorClient(logger, testAggCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to create aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[4/7] Dependency: Test Aggregator Client Initialised")

	healthClient := taskdispatcher.NewHealthClient(logger, config.GetHealthRPCUrl())
	logger.Info(ctx, "[5/7] Dependency: Health Client Initialised")

	// Initialize task stream manager for orchestration
	taskStreamMgr, err := tasks.NewTaskStreamManager(ctx, redisClient, aggClient, testAggClient, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskStreamManager", observability.Error(err))
	}
	logger.Info(ctx, "[6/7] Dependency: Task Stream Manager Initialised")

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
	logger.Info(ctx, "[7/7] Dependency: Task Dispatcher Initialised")

	// Initialize metrics collector
	collector := metrics.NewCollector(obsMetrics)
	collector.Start()
	logger.Info(ctx, "[1/2] Process: Metrics Collector Started")

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
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Fatal(ctx, "Failed to start RPC server", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[2/2] Process: RPC Server Started")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until signal is received
	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	// Perform graceful shutdown
	performGracefulShutdown(ctx, srv, dispatcher, logger, loggerShutdown, tracerShutdown, metricsShutdown)
}

// performGracefulShutdown handles graceful shutdown of the service
func performGracefulShutdown(ctx context.Context, server *rpcserver.Server, dispatcher *taskdispatcher.TaskDispatcher, logger observability.Logger, loggerShutdown func(context.Context) error, tracerShutdown func(context.Context) error, metricsShutdown func(context.Context) error) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		if tracerShutdown != nil {
			if err := tracerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down tracer", observability.Error(err))
			}
		}

		if metricsShutdown != nil {
			if err := metricsShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down metrics", observability.Error(err))
			}
		}

		// Close the Dispatcher
		if err := dispatcher.Close(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Failed to close dispatcher", observability.Error(err))
		}

		// Shutdown server gracefully
		if err := server.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "RPC server forced to shutdown", observability.Error(err))
		}

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown logger (optional here, often last)
		if loggerShutdown != nil {
			if err := loggerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
			}
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
