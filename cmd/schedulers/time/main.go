package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	redisConfig "github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize Redis configuration
	if err := redisConfig.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize Redis config: %v", err))
	}

	// Start metrics collection
	metrics.StartMetricsCollection()

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.TimeSchedulerProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting time-based scheduler service...")

	// Initialize enhanced Redis connection
	var redisClient *redisx.Client
	logger.Info("Initializing enhanced Redis connection...")

	if redisx.IsAvailable() {
		client, err := redisx.NewClient(logger)
		if err != nil {
			logger.Warnf("Failed to create Redis client: %v", err)
			logger.Info("Scheduler will continue without Redis streaming")
		} else {
			redisClient = client
			redisInfo := redisx.GetRedisInfo()
			logger.Infof("Redis client initialized successfully: type=%s, upstash=%v, local=%v",
				redisInfo["type"], redisInfo["upstash"], redisInfo["local"])

			// Add initial startup event to verify Redis streams are working
			startupEvent := map[string]interface{}{
				"event_type":   "service_startup",
				"service":      "time-scheduler",
				"scheduler_id": fmt.Sprintf("time-scheduler-%d", time.Now().Unix()),
				"redis_type":   redisInfo["type"],
				"redis_info":   redisInfo,
				"started_at":   time.Now().Unix(),
				"message":      "Time scheduler service started with enhanced Redis",
			}

			if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, startupEvent); err != nil {
				logger.Warnf("Failed to add startup event to Redis stream: %v", err)
			} else {
				logger.Info("Startup event added to Redis stream successfully")
			}
		}
	} else {
		logger.Warn("Redis not configured - job streaming disabled")
		logger.Infof("To enable Redis, set UPSTASH_REDIS_URL or REDIS_LOCAL_ENABLED=true")
	}

	// Initialize cache with enhanced Redis support
	logger.Info("Initializing enhanced cache system...")
	if err := cache.InitWithLogger(logger); err != nil {
		logger.Warnf("Cache initialization failed: %v", err)
		logger.Info("Scheduler will continue without caching features")
	} else {
		cacheInstance, err := cache.GetCache()
		if err != nil {
			logger.Warnf("Failed to get cache instance: %v", err)
		} else {
			cacheInfo := cache.GetCacheInfo()
			logger.Infof("Cache system initialized: type=%s, redis_available=%v",
				cacheInfo["type"], cacheInfo["redis_available"])

			// Test cache functionality
			testKey := "time_scheduler_startup_test"
			testValue := fmt.Sprintf("startup_%d", time.Now().Unix())
			if err := cacheInstance.Set(testKey, testValue, 1*time.Minute); err != nil {
				logger.Warnf("Cache test write failed: %v", err)
			} else {
				if retrieved, err := cacheInstance.Get(testKey); err == nil && retrieved == testValue {
					logger.Info("Cache functionality verified successfully")
				} else {
					logger.Warnf("Cache test read failed: %v", err)
				}
				// Clean up test key
				if err := cacheInstance.Delete(testKey); err != nil {
					logger.Warnf("Failed to delete test key: %v", err)
				}
			}

			// Test performer lock functionality
			testPerformerID := "test_time_startup_performer"
			if acquired, err := cacheInstance.AcquirePerformerLock(testPerformerID, 30*time.Second); err != nil {
				logger.Warnf("Performer lock test failed: %v", err)
			} else if acquired {
				logger.Info("Performer lock functionality verified")
				if err := cacheInstance.ReleasePerformerLock(testPerformerID); err != nil {
					logger.Warnf("Failed to release test performer lock: %v", err)
				}
			} else {
				logger.Warn("Performer lock test: lock not acquired (unexpected)")
			}
		}
	}

	// Initialize database client
	dbClientCfg := client.Config{
		DBServerURL:    config.GetDBServerURL(),
		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,
	}
	dbClient, err := client.NewDBServerClient(logger, dbClientCfg)
	if err != nil {
		logger.Fatal("Failed to initialize database client", "error", err)
	}
	defer dbClient.Close()

	// Perform initial health check
	logger.Info("Performing initial health check...")
	if err := dbClient.HealthCheck(); err != nil {
		logger.Warn("Database server health check failed", "error", err)
		logger.Info("Continuing startup - will retry connections during operation")
	} else {
		logger.Info("Database server health check passed")
	}

	// Initialize time-based scheduler
	managerID := fmt.Sprintf("time-scheduler-%d", time.Now().Unix())
	timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize time-based scheduler", "error", err)
	}

	// Setup HTTP server with scheduler integration
	srv := api.NewServer(api.Config{
		Port: config.GetSchedulerRPCPort(),
	}, api.Dependencies{
		Logger:    logger,
		Scheduler: timeScheduler,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler in background
	go func() {
		logger.Info("Starting time-based scheduler worker management...")
		timeScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server for job scheduling API...", "port", config.GetSchedulerRPCPort())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Log comprehensive service status
	serviceStatus := map[string]interface{}{
		"manager_id":      managerID,
		"api_port":        config.GetSchedulerRPCPort(),
		"redis_available": redisx.IsAvailable(),
		"cache_available": cache.IsRedisAvailable(),
		"max_workers":     config.GetMaxWorkers(),
		"poll_interval":   "30s",
	}

	if redisx.IsAvailable() {
		serviceStatus["redis_info"] = redisx.GetRedisInfo()
	}

	if cache.IsRedisAvailable() {
		serviceStatus["cache_info"] = cache.GetCacheInfo()
	}

	logger.Info("Time-based scheduler service ready", serviceStatus)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(cancel, srv, timeScheduler, redisClient, logger)
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

func performGracefulShutdown(cancel context.CancelFunc, srv *api.Server, timeScheduler *scheduler.TimeBasedScheduler, redisClient *redisx.Client, logger logging.Logger) {
	shutdownStart := time.Now()
	logger.Info("Initiating graceful shutdown...")

	// Add shutdown event to Redis stream
	if redisClient != nil {
		shutdownEvent := map[string]interface{}{
			"event_type":  "service_shutdown",
			"service":     "time-scheduler",
			"shutdown_at": shutdownStart.Unix(),
			"graceful":    true,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, shutdownEvent); err != nil {
			logger.Warnf("Failed to add shutdown event to Redis stream: %v", err)
		} else {
			logger.Info("Shutdown event added to Redis stream")
		}
	}

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully
	timeScheduler.Stop()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Close Redis client if available
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Warnf("Error closing Redis client: %v", err)
		} else {
			logger.Info("Redis client closed successfully")
		}
	}

	shutdownDuration := time.Since(shutdownStart)

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	logger.Info("Time-based scheduler shutdown complete", "duration", shutdownDuration)
	os.Exit(0)
}
