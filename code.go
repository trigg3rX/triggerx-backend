package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   "code-executor-test",
		IsDevelopment: true,
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting CodeExecutor test...")

	// Initialize DockerManager from config file
	configFilePath := "config.yaml"
	dockerManager, err := dockerexecutor.NewDockerExecutorFromFile(configFilePath, logger)
	if err != nil {
		logger.Fatal("Failed to create DockerManager", "error", err)
	}

	// Initialize the DockerManager
	ctx := context.Background()
	if err := dockerManager.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize DockerManager", "error", err)
	}

	logger.Info("DockerManager initialized successfully")

	// Test execution parameters
	fileURL := "https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreia2figqi2fme3gw5qauzgmpkky4exmvmvpox6qlf45rjsmd6qqpum" // Replace with actual test URL
	fileLanguage := "go"
	noOfAttesters := 1

	// Execute the code
	result, err := dockerManager.Execute(ctx, fileURL, fileLanguage, noOfAttesters)
	if err != nil {
		logger.Fatal("Failed to execute code", "error", err)
	}
	logger.Info("Code executed successfully")

	logger.Infof("Code executed successfully: %v", result)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, logger, dockerManager)
}

func performGracefulShutdown(ctx context.Context, logger logging.Logger, dockerManager *dockerexecutor.DockerExecutor) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Close Docker manager with timeout
	if dockerManager != nil {
		logger.Info("Closing Docker manager...")
		if err := dockerManager.Close(shutdownCtx); err != nil {
			logger.Error("Failed to close Docker manager", "error", err)
		} else {
			logger.Info("Docker manager closed successfully")
		}
	}

	logger.Info("Shutdown complete")
	os.Exit(0)
}
