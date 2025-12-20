package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/api"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/health"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/taskmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize configuration: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.KeeperService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)
	obsCfg.EnablePrometheusExport = config.GetEnablePrometheusExport()

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
	logger.Info(ctx, "Starting keeper node ...",
		observability.String("keeper_address", config.GetKeeperAddress()),
		observability.String("consensus_address", config.GetConsensusAddress()),
		observability.String("version", config.GetVersion()),
	)

	// Initialize observability Metrics
	obsMetrics, metricsShutdown, err := observability.NewMetrics(obsCfg, res)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize observability metrics", observability.Error(err))
	}
	logger.Info(ctx, "[1/7] Dependency: Observability Metrics Initialised")

	// Create metrics collector with observability Metrics
	collector := metrics.NewCollector(obsMetrics)
	logger.Info(ctx, "[2/7] Dependency: Metrics collector Initialised")

	// Initialize health client first
	healthCfg := health.Config{
		HealthServiceURL: config.GetHealthRPCUrl(),
		PrivateKey:       config.GetPrivateKeyConsensus(),
		KeeperAddress:    config.GetKeeperAddress(),
		PeerID:           config.GetPeerID(),
		Version:          config.GetVersion(),
		RequestTimeout:   10 * time.Second,
	}
	healthClient, err := health.NewClient(logger, healthCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize health client", observability.Error(err))
	}
	logger.Info(ctx, "[3/7] Dependency: Health client Initialised")

	// Perform initial health check-in to get configuration
	logger.Info(ctx, "Performing initial health check-in to get configuration...")
	response, err := healthClient.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, health.ErrKeeperNotVerified) {
			logger.Fatal(ctx, "Keeper is not verified. Shutting down...", observability.Error(err))
		}
		logger.Fatal(ctx, "Failed initial health check-in", observability.Any("error", response.Data))
	}
	logger.Info(ctx, "Initial health check-in successful, configuration received")

	// Initialize clients: ECDSA
	aggregatorCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetPrivateKeyController(),
		SenderAddress:    config.GetKeeperAddress(),
	}
	aggregatorClient, err := aggregator.NewAggregatorClient(logger, aggregatorCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[4/7] Dependency: Aggregator client Initialised")

	dockerManager, err := dockerexecutor.NewDockerExecutorFromFile("config/docker-executor.yaml", logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize code executor", observability.Error(err))
	}

	// Initialize the Docker manager with language-specific pools
	if err := dockerManager.Initialize(ctx); err != nil {
		logger.Fatal(ctx, "Failed to initialize Docker manager", observability.Error(err))
	}
	logger.Info(ctx, "[5/7] Dependency: Code executor Initialised")

	ipfsCfg := ipfs.NewConfig(config.GetIpfsHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize IPFS client", observability.Error(err))
	}
	logger.Info(ctx, "[6/7] Dependency: IPFS client Initialised")

	// Initialize taskmonitor client (optional - may not be configured)
	var taskMonitorClient *taskmonitor.Client
	if config.GetTaskMonitorRPCUrl() != "" {
		taskMonitorClient, err = taskmonitor.NewClient(logger)
		if err != nil {
			logger.Warn(ctx, "Failed to initialize taskmonitor client, error reporting will be disabled",
				observability.Error(err))
		} else {
			logger.Info(ctx, "[6/7] Dependency: TaskMonitor client Initialised")
		}
	} else {
		logger.Info(ctx, "[6/7] Dependency: TaskMonitor client skipped (not configured)")
	}

	// Initialize task executor and validator
	validator := validation.NewTaskValidator(config.GetAlchemyAPIKey(), config.GetEtherscanAPIKey(), dockerManager, aggregatorClient, logger, ipfsClient)
	executor := execution.NewTaskExecutor(config.GetAlchemyAPIKey(), validator, aggregatorClient, taskMonitorClient, logger, tracer)

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetOperatorRPCPort(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	deps := &api.Dependencies{
		Logger:    logger,
		Metrics:   obsMetrics,
		Executor:  executor,
		Validator: validator,
	}

	server := api.NewServer(serverCfg, deps)
	logger.Info(ctx, "[7/7] Dependency: API server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check routine
	go startHealthCheckRoutine(ctx, healthClient, dockerManager, taskMonitorClient, logger, server)
	logger.Debug(ctx, "Note: Only first health-check will be logged, subsequent health-checks will not be logged.")
	logger.Info(ctx, "[1/3] Process: Health check routine Started")

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal(ctx, "Failed to start server", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[2/3] Process: API server Started")

	// Start metrics collector in a goroutine
	go func() {
		collector.Start()
	}()
	logger.Info(ctx, "[3/3] Process: Metrics collector Started")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until signal is received
	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	// Perform graceful shutdown
	performGracefulShutdown(ctx, healthClient, dockerManager, server, taskMonitorClient, logger, loggerShutdown, metricsShutdown, tracerShutdown)
}

// startHealthCheckRoutine starts a goroutine that sends periodic health check-ins
func startHealthCheckRoutine(ctx context.Context, healthClient *health.Client, dockerManager dockerexecutor.DockerExecutorAPI, taskMonitorClient *taskmonitor.Client, logger observability.Logger, server *api.Server) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Skip initial check-in since we already did it during startup
	logger.Debug(ctx, "Starting periodic health check routine")

	for {
		select {
		case <-ticker.C:
			response, err := healthClient.CheckIn(ctx)
			if err != nil {
				if errors.Is(err, health.ErrKeeperNotVerified) {
					logger.Error(ctx, "Keeper is not verified. Shutting down...", observability.Error(err))
					// Note: shutdown functions are not available in this scope, but that's okay for emergency shutdown
					performGracefulShutdown(ctx, healthClient, dockerManager, server, taskMonitorClient, logger, nil, nil, nil)
					return
				}
				logger.Error(ctx, "Failed health check-in", observability.Any("error", response.Data))
			}
		case <-ctx.Done():
			logger.Info(ctx, "Stopping health check routine")
			return
		}
	}
}

func performGracefulShutdown(ctx context.Context, healthClient *health.Client, dockerManager dockerexecutor.DockerExecutorAPI, server *api.Server, taskMonitorClient *taskmonitor.Client, logger observability.Logger, loggerShutdown func(context.Context) error, metricsShutdown func(context.Context) error, tracerShutdown func(context.Context) error) {
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Shutdown metrics
		if metricsShutdown != nil {
			if err := metricsShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down metrics", observability.Error(err))
			} else {
				logger.Info(shutdownCtx, "[1/6] Process: Metrics Closed")
			}
		}

		// Shutdown tracer
		if tracerShutdown != nil {
			if err := tracerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down tracer", observability.Error(err))
			} else {
				logger.Info(shutdownCtx, "[2/6] Process: Tracer Closed")
			}
		}

		// Shutdown logger
		if loggerShutdown != nil {
			if err := loggerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
			} else {
				logger.Info(shutdownCtx, "[3/6] Process: Logger Closed")
			}
		}

		// Close health client
		healthClient.Close()
		logger.Info(shutdownCtx, "[4/6] Process: Health client Closed")

		// Close code executor
		if err := dockerManager.Close(ctx); err != nil {
			logger.Error(shutdownCtx, "Error closing code executor", observability.Error(err))
		}
		logger.Info(shutdownCtx, "[5/6] Process: Code executor Closed")

		// Close taskmonitor client
		if taskMonitorClient != nil {
			if err := taskMonitorClient.Close(ctx); err != nil {
				logger.Error(shutdownCtx, "Error closing taskmonitor client", observability.Error(err))
			} else {
				logger.Info(shutdownCtx, "[6/6] Process: TaskMonitor client Closed")
			}
		}

		// Shutdown server gracefully
		if err := server.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
		}
		logger.Info(shutdownCtx, "[7/6] Process: API server Stopped")
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		logger.Info(ctx, "Graceful shutdown completed successfully")
	case <-shutdownCtx.Done():
		logger.Warn(ctx, "Shutdown timeout reached, forcing exit")
	}

	logger.Info(ctx, "Shutdown complete")
	os.Exit(0)
}
