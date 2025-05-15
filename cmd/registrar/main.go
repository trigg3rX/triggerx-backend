package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/rewards"
	dbpkg "github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:      logging.BaseDataDir,
		ProcessName: logging.RegistrarProcess,
		Environment: getEnvironment(),
		UseColors:   true,
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting registrar service...")

	// Initialize database connection
	dbConfig := dbpkg.NewConfig(config.GetDatabaseHost(), config.GetDatabaseHostPort())
	dbConn, err := dbpkg.NewConnection(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer dbConn.Close()

	// Initialize database manager with logger
	client.InitDatabaseManager(logger, dbConn)
	logger.Info("Database manager initialized")

	// Log contract addresses
	logger.Info("Contract addresses initialized",
		"avsGovernance", config.GetAvsGovernanceAddress(),
		"attestationCenter", config.GetAttestationCenterAddress(),
	)

	// Initialize and start registrar service
	registrarService, err := registrar.NewRegistrarService(logger)
	if err != nil {
		logger.Fatal("Failed to initialize registrar service", "error", err)
	}

	// Start services
	registrarService.Start()

	// Initialize and start rewards service
	rewardsService := rewards.NewRewardsService(logger)
	go rewardsService.StartDailyRewardsPoints()

	logger.Info("All services started successfully")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	logger.Info("Received shutdown signal", "signal", sig.String())

	// Cleanup
	registrarService.Stop()
	logger.Info("Shutdown complete")
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}
