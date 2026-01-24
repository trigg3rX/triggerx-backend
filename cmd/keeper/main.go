package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/api"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/checkin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	taskmonitor "github.com/trigg3rX/triggerx-backend/internal/keeper/rpc/clients/taskmonitor"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/keeper.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	// Generate instance ID that includes keeper address for better uniqueness
	keeperInstanceID := observability.GenerateKeeperInstanceID(config.GetKeeperAddress())

	obsCfg := observability.NewConfig(
		observability.KeeperService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		keeperInstanceID,
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
	collector := metrics.NewCollector(obsMetrics, logger)

	logger.Info(ctx, "[1/7] Dependency: Observability Module Initialised")

	// Initialize health client first
	healthCfg := checkin.Config{
		HealthServiceURL: config.GetHealthRPCUrl(),
		PrivateKey:       config.GetPrivateKeyConsensus(),
		KeeperAddress:    config.GetKeeperAddress(),
		PeerID:           config.GetPeerID(),
		Version:          config.GetVersion(),
		Network:          config.GetNetwork(),
		RequestTimeout:   config.GetHealthRequestTimeout(),
	}
	healthClient, err := checkin.NewClient(logger, tracer, healthCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize health client", observability.Error(err))
	}
	logger.Info(ctx, "[2/7] Dependency: Health Client Initialised")

	// Initialize check-in manager and perform initial check-in
	var server *api.Server // Forward declaration for shutdown callback
	checkinManager := checkin.NewManager(healthClient, logger, checkin.ManagerConfig{
		Interval: config.GetHealthCheckInterval(),
		OnShutdown: func() {
			performGracefulShutdown(ctx, logger, obs,
				healthClient, nil, nil, nil, nil, server)
		},
	})

	// Perform initial health check-in to get configuration (API keys, etc.)
	if err := checkinManager.PerformInitialCheckIn(ctx); err != nil {
		logger.Fatal(ctx, "Initial health check-in failed", observability.Error(err))
	}
	logger.Info(ctx, "[3/7] Dependency: Initial Health Check-in Completed")

	// Initialize aggregator client (using local wrapper)
	aggregatorClient, err := aggregator.NewAggregatorClient(logger, aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetPrivateKeyController(),
		SenderAddress:    config.GetKeeperAddress(),
	})
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize aggregator client", observability.Error(err))
	}
	logger.Info(ctx, "[4/7] Dependency: Aggregator Client Initialised",
		observability.String("network", config.GetNetwork()),
		observability.String("aggregator_rpc_url", config.GetAggregatorRPCUrl()))

	dockerManager, err := dockerexecutor.NewDockerExecutorFromFile("config/services/docker-executor.yaml", logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize code executor", observability.Error(err))
	}

	// Initialize the Docker manager with language-specific pools
	if err := dockerManager.Initialize(ctx); err != nil {
		logger.Fatal(ctx, "Failed to initialize Docker manager", observability.Error(err))
	}
	logger.Info(ctx, "[5/7] Dependency: Code Executor Initialised")

	ipfsCfg := ipfs.NewConfig(config.GetIpfsHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize IPFS client", observability.Error(err))
	}
	logger.Info(ctx, "[6/7] Dependency: IPFS Client Initialised")

	// Initialize taskmonitor client (optional - may not be configured)
	var taskMonitorClient *taskmonitor.Client
	taskMonitorClient, err = taskmonitor.NewClient(logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize TaskMonitor client", observability.Error(err))
	}
	logger.Info(ctx, "[7/7] Dependency: TaskMonitor Client Initialised")

	// Initialize task executor and validator
	validator := validation.NewTaskValidator(config.GetAlchemyAPIKey(), config.GetEtherscanAPIKey(), dockerManager, aggregatorClient, logger, tracer, ipfsClient)
	executor := execution.NewTaskExecutor(config.GetAlchemyAPIKey(), validator, aggregatorClient, taskMonitorClient, logger, tracer)

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetOperatorRPCPort(),
		ReadTimeout:    config.GetAPIReadTimeout(),
		WriteTimeout:   config.GetAPIWriteTimeout(),
		MaxHeaderBytes: 1 << 20,
	}

	deps := &api.Dependencies{
		Logger:    logger,
		Metrics:   obsMetrics,
		Executor:  executor,
		Validator: validator,
	}

	server = api.NewServer(serverCfg, deps)
	logger.Info(ctx, "[1/3] Process: API server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics collector in a goroutine
	go func() {
		collector.Start()
	}()
	logger.Info(ctx, "[2/3] Process: Metrics Collector Started")

	// Start health check-in routine using checkin manager
	go checkinManager.Start(ctx)
	logger.Info(ctx, "[3/3] Process: Health Check Routine Started")

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal(ctx, "Failed to start server", observability.Error(err))
		}
	}()
	logger.Info(ctx, "Keeper node initialized and ready to serve requests", observability.String("port", config.GetOperatorRPCPort()))

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

func performGracefulShutdown(
	ctx context.Context,
	logger observability.Logger,
	obs *observability.Observability,
	healthClient *checkin.Client,
	aggregatorClient *aggregator.AggregatorClient,
	dockerManager dockerexecutor.DockerExecutorAPI,
	ipfsClient ipfs.IPFSClient,
	taskMonitorClient *taskmonitor.Client,
	server *api.Server,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, config.GetShutdownTimeout())
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Close health client
		if healthClient != nil {
			healthClient.Close()
		}
		logger.Info(ctx, "[1/6] Shutdown: Health Client Closed")

		// Close aggregator client
		if aggregatorClient != nil {
			aggregatorClient.Close()
		}
		logger.Info(ctx, "[2/6] Shutdown: Aggregator Client Closed")

		// Close code executor
		if dockerManager != nil {
			if err := dockerManager.Close(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error closing code executor", observability.Error(err))
			}
		}
		logger.Info(ctx, "[3/6] Shutdown: Docker Manager Closed")

		if ipfsClient != nil {
			ipfsClient.Close()
		}
		logger.Info(ctx, "[4/6] Shutdown: IPFS Client Closed")

		if taskMonitorClient != nil {
			if err := taskMonitorClient.Close(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error closing taskmonitor client", observability.Error(err))
			}
		}

		// Shutdown server gracefully
		if server != nil {
			if err := server.Stop(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
			}
		}
		logger.Info(ctx, "[5/6] Shutdown: API Server Stopped")

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (handles logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down observability", observability.Error(err))
		}
		logger.Info(ctx, "[6/6] Shutdown: Observability Shutdown Complete")
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
