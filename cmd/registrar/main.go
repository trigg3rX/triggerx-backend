package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.RegistrarProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting registrar service...",
		"port", config.GetRegistrarPort(),
		"avs_governance", config.GetAvsGovernanceAddress(),
		"attestation_center", config.GetAttestationCenterAddress(),
		"trigger_gas_registry", config.GetTriggerGasRegistryAddress(),
	)

	// Initialize and start registrar service
	registrarService, err := registrar.NewRegistrarService(logger)
	if err != nil {
		logger.Fatal("Failed to initialize registrar service", "error", err)
	}

	// Start services
	err = registrarService.Start()
	if err != nil {
		logger.Fatal("Failed to start registrar service", "error", err)
	}

	logger.Info("All services started successfully")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	logger.Infof("Received shutdown signal: %s", sig.String())

	// Perform graceful shutdown
	performGracefulShutdown(registrarService, logger)

	logger.Info("Shutdown complete")
}

func performGracefulShutdown(registrarService *registrar.RegistrarService, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := registrarService.Stop()
	if err != nil {
		logger.Fatal("Failed to stop registrar service", "error", err)
	}

	// Wait for context timeout or manual cancellation
	<-ctx.Done()

	logger.Info("Shutdown complete")
}
