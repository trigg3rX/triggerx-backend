package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/health"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/database"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/notification"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init("config/health.yaml"); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.HealthProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting health service...")

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	// Initialize database connection
	dbConfig := &connection.Config{
		Hosts:               []string{config.GetDatabaseHost() + ":" + config.GetDatabasePort()},
		Keyspace:            "triggerx",
		Consistency:         gocql.Quorum,
		Timeout:             time.Second * 30,
		Retries:             5,
		ConnectWait:         time.Second * 10,
		ProtoVersion:        4,
		HealthCheckInterval: 15 * time.Second,
		SocketKeepalive:     15 * time.Second,
		MaxPreparedStmts:    1000,
		DefaultIdempotence:  true,
	}
	dbConn, err := datastore.NewService(dbConfig, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database connection: %v", err))
	}

	// Create keeper repository
	keeperRepo := dbConn.Keeper()

	dbManager := database.InitDatabaseManager(logger, keeperRepo)
	logger.Info("Database manager initialized")

	// Initialize notification bot
	notificationBot, err := notification.NewBot(config.GetBotToken(), logger, dbManager)
	if err != nil {
		logger.Errorf("Failed to initialize notification bot: %v", err)
	}
	notificationBot.Start()
	logger.Info("Notification bot initialized")

	// Initialize state manager with notification bot for inactivity alerts
	stateManager := keeper.InitializeStateManager(logger, dbManager, notificationBot)
	logger.Info("Keeper state manager initialized with notification support")

	// Set global state manager for orchestrator access
	health.SetStateManager(stateManager)

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(context.Background()); err != nil {
		logger.Debug("Failed to load verified keepers from database", "error", err)
		// Continue anyway, as we can still operate with an empty state
	}

	// Create and start health service orchestrator (HTTP + gRPC)
	healthService := health.NewService(logger, &health.Config{
		HTTPPort: config.GetHTTPPort(),
		GRPCPort: config.GetGRPCPort(),
		GRPCHost: "0.0.0.0",
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := healthService.Start(context.Background()); err != nil {
			serverErrors <- fmt.Errorf("health service error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Health service is ready - HTTP: %s, gRPC: %s",
		config.GetHTTPPort(),
		config.GetGRPCPort())

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(healthService, &wg, logger)
}

func performGracefulShutdown(healthService *health.Service, wg *sync.WaitGroup, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Update all keepers to inactive in database
	stateManager := keeper.GetStateManager()
	if err := stateManager.DumpState(ctx); err != nil {
		logger.Error("Failed to dump keeper state", "error", err)
	}

	// Stop health service (HTTP + gRPC servers)
	if err := healthService.Stop(ctx); err != nil {
		logger.Error("Health service shutdown error", "error", err)
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
