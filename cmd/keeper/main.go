package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/api"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/health"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize configuration: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:      logging.BaseDataDir,
		ProcessName: logging.KeeperProcess,
		Environment: getEnvironment(),
		UseColors:   true,
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting keeper node ...")
	logger.Info("Keeper address: ", "address", config.GetKeeperAddress())
	logger.Info("Consensus address: ", "address", config.GetConsensusAddress())

	// Initialize clients
	aggregatorCfg := aggregator.Config{
		RPCAddress:     config.GetAggregatorRPCUrl(),
		PrivateKey:     config.GetPrivateKeyController(),
		KeeperAddress:  config.GetKeeperAddress(),
		RetryAttempts:  3,
		RetryDelay:     2 * time.Second,
		RequestTimeout: 10 * time.Second,
	}
	aggregatorClient, err := aggregator.NewClient(logger, aggregatorCfg)
	if err != nil {
		logger.Fatal("Failed to initialize aggregator client", "error", err)
	}
	defer aggregatorClient.Close()

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
	defer healthClient.Close()

	ethClient, err := ethclient.Dial(config.GetAggregatorRPCUrl())
	if err != nil {
		logger.Fatal("Failed to initialize Ethereum client", "error", err)
	}
	defer ethClient.Close()

	// Initialize task executor and validator
	executor := execution.NewTaskExecutor(ethClient, config.GetEtherscanAPIKey(), logger)
	validator := validation.NewTaskValidator(logger, ethClient)

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetOperatorRPCPort(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	deps := api.Dependencies{
		Logger:    logger,
		Executor:  *executor,
		Validator: *validator,
	}

	server := api.NewServer(serverCfg, deps)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check routine
	go startHealthCheckRoutine(ctx, healthClient, logger, server)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, server, logger)
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

// startHealthCheckRoutine starts a goroutine that sends periodic health check-ins
func startHealthCheckRoutine(ctx context.Context, healthClient *health.Client, logger logging.Logger, server *api.Server) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial check-in
	response, err := healthClient.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, health.ErrKeeperNotVerified) {
			logger.Error("Keeper is not verified. Shutting down...", "error", err)
			performGracefulShutdown(ctx, server, logger)
			return
		}
		logger.Error("Failed initial health check-in", "error", response.Data)
	}

	for {
		select {
		case <-ticker.C:
			response, err := healthClient.CheckIn(ctx)
			if err != nil {
				logger.Error("Failed health check-in", "error", response.Data)
			}
		case <-ctx.Done():
			logger.Info("Stopping health check routine")
			return
		}
	}
}

func performGracefulShutdown(ctx context.Context, server *api.Server, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	logger.Info("Shutdown complete")
	os.Exit(0)
}
