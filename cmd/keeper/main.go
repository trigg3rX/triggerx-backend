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
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	dockerconfig "github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// Initialize configuration
	err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize configuration: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.KeeperProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting keeper node ...",
		"keeper_address", config.GetKeeperAddress(),
		"consensus_address", config.GetConsensusAddress(),
		"version", config.GetVersion(),
	)

	collector := metrics.NewCollector()
	logger.Info("[1/6] Dependency: Metrics collector Initialised")

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
		logger.Fatal("Failed to initialize health client", "error", err)
	}
	logger.Info("[2/6] Dependency: Health client Initialised")

	// Perform initial health check-in to get configuration
	logger.Info("Performing initial health check-in to get configuration...")
	ctx := context.Background()
	response, err := healthClient.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, health.ErrKeeperNotVerified) {
			logger.Fatal("Keeper is not verified. Shutting down...", "error", err)
		}
		logger.Fatal("Failed initial health check-in", "error", response.Data)
	}
	logger.Info("Initial health check-in successful, configuration received")

	// Initialize clients: ECDSA
	aggregatorCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetPrivateKeyController(),
		SenderAddress:    config.GetKeeperAddress(),
	}
	aggregatorClient, err := aggregator.NewAggregatorClient(logger, aggregatorCfg)
	if err != nil {
		logger.Fatal("Failed to initialize aggregator client", "error", err)
	}
	logger.Info("[3/6] Dependency: Aggregator client Initialised")

	dockerCfg := dockerconfig.DefaultConfig("go")
	supportedLanguages := []types.Language{
		types.LanguageGo,
		// types.LanguagePy,
		// types.LanguageJS,
		// types.LanguageTS,
		// types.LanguageNode,
	}

	dockerManager, err := docker.NewDockerManager(dockerCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize code executor", "error", err)
	}

	// Initialize the Docker manager with language-specific pools
	if err := dockerManager.Initialize(ctx, supportedLanguages); err != nil {
		logger.Fatal("Failed to initialize Docker manager", "error", err)
	}
	logger.Infof("[4/6] Dependency: Code executor Initialised with %d language pools", len(supportedLanguages))

	ipfsCfg := ipfs.NewConfig(config.GetIpfsHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize IPFS client", "error", err)
	}
	logger.Info("[5/6] Dependency: IPFS client Initialised")

	// Initialize task executor and validator
	validator := validation.NewTaskValidator(config.GetAlchemyAPIKey(), config.GetEtherscanAPIKey(), dockerManager, aggregatorClient, logger, ipfsClient)
	executor := execution.NewTaskExecutor(config.GetAlchemyAPIKey(), validator, aggregatorClient, logger)

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetOperatorRPCPort(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	deps := &api.Dependencies{
		Logger:    logger,
		Executor:  executor,
		Validator: validator,
	}

	server := api.NewServer(serverCfg, deps)
	logger.Info("[6/6] Dependency: API server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check routine
	go startHealthCheckRoutine(ctx, healthClient, dockerManager, logger, server)
	logger.Debug("Note: Only first health-check will be logged, subsequent health-checks will not be logged.")
	logger.Info("[1/3] Process: Health check routine Started")

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()
	logger.Info("[2/3] Process: API server Started")

	// Start metrics collector in a goroutine
	go func() {
		collector.Start()
	}()
	logger.Info("[3/3] Process: Metrics collector Started")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, healthClient, dockerManager, server, logger)
}

// startHealthCheckRoutine starts a goroutine that sends periodic health check-ins
func startHealthCheckRoutine(ctx context.Context, healthClient *health.Client, dockerManager *docker.DockerManager, logger logging.Logger, server *api.Server) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Skip initial check-in since we already did it during startup
	logger.Debug("Starting periodic health check routine")

	for {
		select {
		case <-ticker.C:
			response, err := healthClient.CheckIn(ctx)
			if err != nil {
				if errors.Is(err, health.ErrKeeperNotVerified) {
					logger.Error("Keeper is not verified. Shutting down...", "error", err)
					performGracefulShutdown(ctx, healthClient, dockerManager, server, logger)
					return
				}
				logger.Error("Failed health check-in", "error", response.Data)
			}
		case <-ctx.Done():
			logger.Info("Stopping health check routine")
			return
		}
	}
}

func performGracefulShutdown(ctx context.Context, healthClient *health.Client, dockerManager *docker.DockerManager, server *api.Server, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Close health client
	healthClient.Close()
	logger.Info("[1/3] Process: Health client Closed")

	// Close code executor
	if err := dockerManager.Close(); err != nil {
		logger.Error("Error closing code executor", "error", err)
	}
	logger.Info("[2/3] Process: Code executor Closed")

	// Shutdown server gracefully
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	logger.Info("[3/3] Process: API server Stopped")

	logger.Info("Shutdown complete")
	os.Exit(0)
}
