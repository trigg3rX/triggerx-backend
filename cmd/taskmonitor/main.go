package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.TaskMonitorProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Task Monitor service ...")

	// Initialize TaskManager (handles Redis, Database, IPFS, Event Listener, and Task Stream Manager)
	taskManager, err := taskmonitor.NewTaskManager(logger)
	if err != nil {
		logger.Fatal("Failed to create TaskManager", "error", err)
	}
	logger.Info("[1/5] TaskManager created successfully")

	// Initialize all components
	if err := taskManager.Initialize(); err != nil {
		logger.Fatal("Failed to initialize TaskManager components", "error", err)
	}
	logger.Info("[2/5] TaskManager components initialized successfully")

	// Create context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Log service status
	logger.Info("Task Monitor service is running")

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(taskManager, logger)
}

func performGracefulShutdown(taskManager *taskmonitor.TaskManager, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Close TaskManager (handles all components)
	logger.Info("Closing TaskManager...")
	if err := taskManager.Close(); err != nil {
		logger.Warn("Non-critical errors during TaskManager shutdown", "error", err)
	} else {
		logger.Info("TaskManager closed successfully")
	}
	logger.Info("Task Monitor service shutdown complete")
}
