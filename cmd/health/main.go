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
	"github.com/trigg3rX/triggerx-backend/internal/health/cache"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/database"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/notification"
	"github.com/trigg3rX/triggerx-backend/internal/health/rewards"
	redisclient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
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

	// Initialize Redis client for rewards tracking
	redisConfig := redisclient.RedisConfig{
		UpstashConfig: redisclient.UpstashConfig{
			URL:   config.GetRedisURL(),
			Token: config.GetRedisPassword(),
		},
		ConnectionSettings: redisclient.ConnectionSettings{
			PoolSize:     10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
		},
	}

	redisClient, err := redisclient.NewRedisClient(logger, redisConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Redis client: %v", err))
	}
	logger.Info("Redis client initialized")

	// Create rewards cache wrapper
	rewardsCache := cache.NewRewardsCache(redisClient, logger)

	// Initialize notification bot
	notificationBot, err := notification.NewBot(config.GetBotToken(), logger, dbManager)
	if err != nil {
		logger.Errorf("Failed to initialize notification bot: %v", err)
	}
	notificationBot.Start()
	logger.Info("Notification bot initialized")

	// Initialize state manager with notification bot for inactivity alerts and cache for rewards
	stateManager := keeper.InitializeStateManager(logger, dbManager, notificationBot, rewardsCache)
	logger.Info("Keeper state manager initialized with notification support and rewards tracking")

	// Set global state manager for orchestrator access
	health.SetStateManager(stateManager)

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(context.Background()); err != nil {
		logger.Debug("Failed to load verified keepers from database", "error", err)
		// Continue anyway, as we can still operate with an empty state
	}

	// Initialize and start rewards service
	rewardsService := rewards.NewService(logger, rewardsCache, dbManager)
	if err := rewardsService.Start(); err != nil {
		logger.Error("Failed to start rewards service", "error", err)
		// Continue anyway, rewards will be unavailable but health checks will work
	} else {
		logger.Info("Rewards service started successfully")
	}

	// Set global rewards service for API access
	health.SetRewardsService(rewardsService)

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

	performGracefulShutdown(healthService, rewardsService, redisClient, &wg, logger)
}

func performGracefulShutdown(
	healthService *health.Service,
	rewardsService *rewards.Service,
	redisClient redisclient.RedisClientInterface,
	wg *sync.WaitGroup,
	logger logging.Logger,
) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop rewards service
	if rewardsService != nil {
		rewardsService.Stop()
		logger.Info("Rewards service stopped")
	}

	// Update all keepers to inactive in database
	stateManager := keeper.GetStateManager()
	if err := stateManager.DumpState(ctx); err != nil {
		logger.Error("Failed to dump keeper state", "error", err)
	}

	// Stop health service (HTTP + gRPC servers)
	if err := healthService.Stop(ctx); err != nil {
		logger.Error("Health service shutdown error", "error", err)
	}

	// Close Redis connection
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection", "error", err)
		} else {
			logger.Info("Redis connection closed")
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
