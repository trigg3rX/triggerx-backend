package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis/api"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/redis/streams/jobs"
	"github.com/trigg3rX/triggerx-backend/internal/redis/streams/tasks"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
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

	logger.Info("Starting Redis Orchestrator service ...")

	// Initialize metrics collector
	collector := metrics.NewCollector()
	logger.Info("Metrics collector Initialised")
	collector.Start()

	// Create Redis client and verify connection
	redisConfig := config.GetRedisClientConfig()
	client, err := redisClient.NewRedisClient(logger, redisConfig)
	if err != nil {
		logger.Fatal("Failed to create Redis client", "error", err)
	}

	// Set up monitoring hooks for metrics integration
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	client.SetMonitoringHooks(monitoringHooks)

	// Test Redis connection
	if err := client.Ping(); err != nil {
		logger.Fatal("Redis is not reachable", "error", err)
	}
	logger.Info("Redis client Initialised")

	// Initialize job stream manager for orchestration
	jobStreamMgr, err := jobs.NewJobStreamManager(logger, client)
	if err != nil {
		logger.Fatal("Failed to initialize JobStreamManager", "error", err)
	}
	logger.Info("Job stream manager Initialised")

	// Initialize task stream manager for orchestration
	taskStreamMgr, err := tasks.NewTaskStreamManager(logger, client)
	if err != nil {
		logger.Fatal("Failed to initialize TaskStreamManager", "error", err)
	}
	logger.Info("Task stream manager Initialised")

	// Initialize stream managers
	if err := jobStreamMgr.Initialize(); err != nil {
		logger.Fatal("Failed to initialize job streams", "error", err)
	}
	if err := taskStreamMgr.Initialize(); err != nil {
		logger.Fatal("Failed to initialize task streams", "error", err)
	}
	logger.Info("Redis streams initialized successfully")

	// Initialize API server
	serverCfg := api.Config{
		Port:           config.GetRedisRPCPort(),
		ReadTimeout:    config.GetReadTimeout(),
		WriteTimeout:   config.GetWriteTimeout(),
		MaxHeaderBytes: 1 << 20,
	}

	deps := api.Dependencies{
		Logger:           logger,
		TaskStreamMgr:    taskStreamMgr,
		JobStreamMgr:     jobStreamMgr,
		MetricsCollector: collector,
	}

	server := api.NewServer(serverCfg, deps)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start orchestration workers in background
	logger.Info("Starting Redis orchestration workers...")

	// Task Processor Workers - multiple instances for scaling
	consumerName := fmt.Sprintf("redis-worker-%d", time.Now().Unix())
	go taskStreamMgr.StartTaskProcessor(ctx, consumerName)
	logger.Info("Started task processor worker", "consumer_name", consumerName)

	// Timeout Worker - monitors processing timeouts
	go taskStreamMgr.StartTimeoutWorker(ctx)
	logger.Info("Started task timeout worker")

	// Retry Worker - processes retry streams
	go taskStreamMgr.StartRetryWorker(ctx)
	logger.Info("Started task retry worker")

	// Stream Health Monitor - monitors stream health
	go startStreamHealthMonitor(ctx, jobStreamMgr, taskStreamMgr, logger)

	// Start API server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()
	logger.Info("API server Started")

	// Start metrics collector in a goroutine
	go func() {
		collector.Start()
	}()

	// Log Redis info
	logger.Info("Redis Orchestrator service is running",
		"redis_available", config.IsRedisAvailable(),
		"redis_type", config.GetRedisType())

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-shutdown

	// Perform graceful shutdown
	performGracefulShutdown(ctx, client, server, logger)
}

// startStreamHealthMonitor monitors the health of Redis streams
func startStreamHealthMonitor(ctx context.Context, jobStreamMgr *jobs.JobStreamManager, taskStreamMgr *tasks.TaskStreamManager, logger logging.Logger) {
	logger.Info("Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second) // Check health every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			jobInfo := jobStreamMgr.GetJobStreamInfo()
			taskInfo := taskStreamMgr.GetStreamInfo()

			logger.Info("Stream health status",
				"job_streams", jobInfo,
				"task_streams", taskInfo)

			// Log warnings for high stream lengths
			if jobLengths, ok := jobInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range jobLengths {
					if length > 100 { // Warn if more than 100 pending jobs
						logger.Warn("High job stream length detected",
							"stream", stream,
							"length", length)
					}
				}
			}

			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 { // Warn if more than 50 tasks in any stream
						logger.Warn("High task stream length detected",
							"stream", stream,
							"length", length)
					}
				}
			}
		}
	}
}

func performGracefulShutdown(ctx context.Context, client redisClient.RedisClientInterface, server *api.Server, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Close Redis client connection
	logger.Info("Closing Redis client connection...")
	if err := client.Close(); err != nil {
		logger.Error("Error closing Redis client", "error", err)
	} else {
		logger.Info("Redis client closed successfully")
	}

	// Shutdown server gracefully
	logger.Info("Shutting down API server...")
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	} else {
		logger.Info("API server stopped successfully")
	}

	logger.Info("Redis Orchestrator service shutdown complete")

	// Ensure we exit cleanly
	select {
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
