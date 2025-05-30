package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"

	// "github.com/trigg3rX/triggerx-backend/internal/registrar/rewards"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.RegistrarProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting registrar service...",
		"mode", getEnvironment(),
		"avs_governance", config.GetAvsGovernanceAddress(),
		"attestation_center", config.GetAttestationCenterAddress(),
	)

	// Initialize database connection
	dbConfig := database.NewConfig(config.GetDatabaseHost(), config.GetDatabaseHostPort())
	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer dbConn.Close()

	// Initialize database manager with logger
	client.InitDatabaseManager(logger, dbConn)
	logger.Info("Database manager initialized")

	// Initialize and start registrar service
	registrarService, err := registrar.NewRegistrarService(logger)
	if err != nil {
		logger.Fatal("Failed to initialize registrar service", "error", err)
	}

	// Start services
	registrarService.Start()

	// // Initialize and start rewards service
	// rewardsService := rewards.NewRewardsService(logger)
	// go rewardsService.StartDailyRewardsPoints()

	logger.Info("All services started successfully")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	logger.Infof("Received shutdown signal: %s", sig.String())

	// Cleanup
	registrarService.Stop()

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	logger.Info("Shutdown complete")
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

func getLogLevel() logging.Level {
	if config.IsDevMode() {
		return logging.DebugLevel
	}
	return logging.InfoLevel
}
