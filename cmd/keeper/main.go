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
	// Generate instance ID that includes keeper address for better uniqueness
	keeperInstanceID := observability.GenerateKeeperInstanceID(config.GetKeeperAddress())

	obsCfg := observability.NewConfigWithOptions(
		observability.KeeperService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		observability.WithInstanceID(keeperInstanceID),
		observability.WithPrometheusExport(config.GetEnablePrometheusExport()),
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
	logger.Info(ctx, "Starting keeper node ...",
		observability.String("version", config.GetVersion()),
		observability.String("keeper_address", config.GetKeeperAddress()),
		observability.String("consensus_address", config.GetConsensusAddress()),
	)

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	// Create metrics collector with observability Metrics
	collector := metrics.NewCollector(obsMetrics)

	logger.Info(ctx, "[1/7] Dependency: Observability Module Initialised")

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

	// Perform initial health check-in to get configuration
	response, err := healthClient.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, health.ErrKeeperNotVerified) {
			logger.Fatal(ctx, "Keeper is not verified. Shutting down...", observability.Error(err))
		}
		logger.Fatal(ctx, "Failed initial health check-in", observability.Any("error", response.Data))
	}
	logger.Info(ctx, "[2/7] Dependency: Health Client Initialised")

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
	logger.Info(ctx, "[3/7] Dependency: Aggregator Client Initialised")

	dockerManager, err := dockerexecutor.NewDockerExecutorFromFile("config/services/docker-executor.yaml", logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize code executor", observability.Error(err))
	}

	// Initialize the Docker manager with language-specific pools
	if err := dockerManager.Initialize(ctx); err != nil {
		logger.Fatal(ctx, "Failed to initialize Docker manager", observability.Error(err))
	}
	logger.Info(ctx, "[4/7] Dependency: Code Executor Initialised")

	ipfsCfg := ipfs.NewConfig(config.GetIpfsHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize IPFS client", observability.Error(err))
	}
	logger.Info(ctx, "[5/7] Dependency: IPFS Client Initialised")

	// Initialize taskmonitor client (optional - may not be configured)
	var taskMonitorClient *taskmonitor.Client
	taskMonitorClient, err = taskmonitor.NewClient(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskMonitor client", observability.Error(err))
	}
	logger.Info(ctx, "[6/7] Dependency: TaskMonitor Client Initialised")

	// Initialize task executor and validator
	validator := validation.NewTaskValidator(config.GetAlchemyAPIKey(), config.GetEtherscanAPIKey(), dockerManager, aggregatorClient, logger, tracer, ipfsClient)
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

	// Start metrics collector in a goroutine
	go func() {
		collector.Start()
	}()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start health check routine
	go startHealthCheckRoutine(ctx, logger, obs,
		healthClient, aggregatorClient, dockerManager, ipfsClient, taskMonitorClient, server)
	logger.Info(ctx, "[2/3] Process: Health Check Routine Started")

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal(ctx, "Failed to start server", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: API Server Started")
	logger.Info(ctx, "Keeper node initialized and ready to serve requests")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until signal is received
	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	// Perform graceful shutdown
	performGracefulShutdown(ctx, logger, obs,
		healthClient, aggregatorClient, dockerManager, ipfsClient, taskMonitorClient, server)
}

// startHealthCheckRoutine starts a goroutine that sends periodic health check-ins
func startHealthCheckRoutine(
	ctx context.Context,
	logger observability.Logger,
	obs *observability.Observability,
	healthClient *health.Client,
	aggregatorClient *aggregator.AggregatorClient,
	dockerManager dockerexecutor.DockerExecutorAPI,
	ipfsClient ipfs.IPFSClient,
	taskMonitorClient *taskmonitor.Client,
	server *api.Server,
) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Skip initial check-in since we already did it during startup
	// logger.Debug(ctx, "Starting periodic health check routine")

	for {
		select {
		case <-ticker.C:
			response, err := healthClient.CheckIn(ctx)
			if err != nil {
				if errors.Is(err, health.ErrKeeperNotVerified) {
					logger.Error(ctx, "Keeper is not verified. Shutting down...", observability.Error(err))
					performGracefulShutdown(ctx, logger, obs,
						healthClient, aggregatorClient, dockerManager, ipfsClient, taskMonitorClient, server)
					return
				}
				logger.Error(ctx, "Failed health check-in", observability.Any("error", response.Data))
			}
		case <-ctx.Done():
			return
		}
	}
}

func performGracefulShutdown(
	ctx context.Context,
	logger observability.Logger,
	obs *observability.Observability,
	healthClient *health.Client,
	aggregatorClient *aggregator.AggregatorClient,
	dockerManager dockerexecutor.DockerExecutorAPI,
	ipfsClient ipfs.IPFSClient,
	taskMonitorClient *taskmonitor.Client,
	server *api.Server,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Close health client
		healthClient.Close()

		// Close aggregator client
		aggregatorClient.Close()

		// Close code executor
		if err := dockerManager.Close(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error closing code executor", observability.Error(err))
		}

		ipfsClient.Close()

		if err := taskMonitorClient.Close(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error closing taskmonitor client", observability.Error(err))
		}

		// Shutdown server gracefully
		if err := server.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
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
