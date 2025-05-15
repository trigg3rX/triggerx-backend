package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/api"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/health"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"

	// "github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize configuration
	config.Init()

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

	logger.Info("Starting keeper node...")

	// Initialize clients
	aggregatorCfg := aggregator.Config{
		RPCAddress:     config.GetAggregatorRPCAddress(),
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
		HealthServiceURL: config.GetHealthRPCAddress(),
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

	// Initialize core services
	// executor := execution.NewExecutor(logger, aggregatorClient)
	// validator := validation.NewValidator(logger)

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetOperatorRPCPort(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	deps := api.Dependencies{
		Logger: logger,
		// Executor:  executor,
		// Validator: validator,
		HealthSvc: healthClient,
	}

	server := api.NewServer(serverCfg, deps)

	// Start health check routine
	healthCheckCtx, healthCheckCancel := context.WithCancel(context.Background())
	defer healthCheckCancel()
	go startHealthCheckRoutine(healthCheckCtx, healthClient, logger)

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

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Stop(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server shutdown complete")
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

// startHealthCheckRoutine starts a goroutine that sends periodic health check-ins
func startHealthCheckRoutine(ctx context.Context, healthClient *health.Client, logger logging.Logger) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial check-in
	if err := healthClient.CheckIn(ctx); err != nil {
		logger.Error("Failed initial health check-in", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := healthClient.CheckIn(ctx); err != nil {
				logger.Error("Failed health check-in", "error", err)
			}
		case <-ctx.Done():
			logger.Info("Stopping health check routine")
			return
		}
	}
}
