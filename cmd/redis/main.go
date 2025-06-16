package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/api"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.RedisProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Redis service ...",
		"mode", config.IsDevMode(),
		"version", "1.0.0",
	)

	// Initialize metrics collector
	collector := metrics.NewCollector()
	logger.Info("[1/5] Dependency: Metrics collector Initialised")

	// Create Redis client and verify connection
	client, err := redis.NewRedisClient(logger)
	if err != nil {
		logger.Fatal("Failed to create Redis client", "error", err)
	}

	// Test Redis connection
	if err := client.Ping(); err != nil {
		logger.Fatal("Redis is not reachable", "error", err)
	}
	logger.Info("[2/5] Dependency: Redis client Initialised")

	// Initialize task stream manager
	taskStreamMgr, err := redis.NewTaskStreamManager(logger)
	if err != nil {
		logger.Fatal("Failed to initialize TaskStreamManager", "error", err)
	}
	logger.Info("[3/5] Dependency: Task stream manager Initialised")

	// Initialize job stream manager
	jobStreamMgr, err := redis.NewJobStreamManager(logger)
	if err != nil {
		logger.Fatal("Failed to initialize JobStreamManager", "error", err)
	}
	logger.Info("[4/5] Dependency: Job stream manager Initialised")

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetRedisAPIPort(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	deps := api.Dependencies{
		Logger:           logger,
		TaskStreamMgr:    taskStreamMgr,
		JobStreamMgr:     jobStreamMgr,
		MetricsCollector: collector,
	}

	server := api.NewServer(serverCfg, deps)
	logger.Info("[5/5] Dependency: API server Initialised")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background metrics collection
	api.StartBackgroundMetricsCollection()
	logger.Info("[1/3] Process: Background metrics collection Started")

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

	// Log Redis info
	redisInfo := redis.GetRedisInfo()
	logger.Info("Redis service is running", "config", redisInfo)

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, client, server, logger)
}

func performGracefulShutdown(ctx context.Context, client *redis.Client, server *api.Server, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Close Redis client connection
	logger.Info("[1/2] Process: Closing Redis client connection...")
	if err := client.Close(); err != nil {
		logger.Error("Error closing Redis client", "error", err)
	} else {
		logger.Info("Redis client closed successfully")
	}

	// Shutdown server gracefully
	logger.Info("[2/2] Process: Shutting down API server...")
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	} else {
		logger.Info("API server stopped successfully")
	}

	logger.Info("Redis service shutdown complete")

	// Ensure we exit cleanly
	select {
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
