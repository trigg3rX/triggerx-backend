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
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Start metrics collection
	metrics.StartMetricsCollection()

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.ConditionSchedulerProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting condition-based scheduler service...")

	// Initialize Redis connection
	logger.Info("Initializing Redis connection...")
	if err := redisx.Ping(); err != nil {
		logger.Warnf("Redis connection failed: %v", err)
		logger.Info("Scheduler will continue without Redis condition streaming")
	} else {
		logger.Info("Redis connection established successfully")

		// Add initial test job to verify Redis streams are working
		testJob := map[string]interface{}{
			"type":         "scheduler_startup",
			"scheduler_id": fmt.Sprintf("condition-scheduler-%d", time.Now().Unix()),
			"timestamp":    time.Now().Unix(),
			"message":      "Condition scheduler service started",
		}
		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, testJob); err != nil {
			logger.Warnf("Failed to add startup test job to Redis stream: %v", err)
		} else {
			logger.Info("Startup event added to Redis condition stream")
		}
	}

	// Initialize cache
	logger.Info("Initializing cache system...")
	if err := cache.Init(); err != nil {
		logger.Warnf("Cache initialization failed: %v", err)
		logger.Info("Scheduler will continue without caching features")
	} else {
		cacheInstance, err := cache.GetCache()
		if err != nil {
			logger.Warnf("Failed to get cache instance: %v", err)
		} else {
			logger.Info("Cache system initialized successfully")

			// Test cache functionality
			testKey := "condition_scheduler_startup_test"
			testValue := fmt.Sprintf("startup_%d", time.Now().Unix())
			if err := cacheInstance.Set(testKey, testValue, 1*time.Minute); err != nil {
				logger.Warnf("Cache test write failed: %v", err)
			} else {
				if retrieved, err := cacheInstance.Get(testKey); err == nil && retrieved == testValue {
					logger.Info("Cache functionality verified")
				} else {
					logger.Warnf("Cache test read failed: %v", err)
				}
				// Clean up test key
				if err := cacheInstance.Delete(testKey); err != nil {
					logger.Warnf("Failed to delete test key: %v", err)
				}
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

	// Initialize condition-based scheduler
	managerID := fmt.Sprintf("condition-scheduler-%d", time.Now().Unix())
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize condition-based scheduler", "error", err)
	}

	// Setup HTTP server with scheduler integration
	srv := api.NewServer(api.Config{
		Port: config.GetSchedulerRPCPort(),
	}, api.Dependencies{
		Logger:    logger,
		Scheduler: conditionScheduler,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler in background
	go func() {
		logger.Info("Starting condition-based scheduler worker management...")
		conditionScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server for condition job scheduling API...", "port", config.GetSchedulerRPCPort())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("Condition-based scheduler service ready",
		"manager_id", managerID,
		"api_port", config.GetSchedulerRPCPort(),
		"poll_interval", "1 second",
		"supported_conditions", []string{"greater_than", "less_than", "between", "equals", "not_equals", "greater_equal", "less_equal"},
		"supported_sources", []string{"api", "oracle", "static"},
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(cancel, srv, conditionScheduler, logger)
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

func performGracefulShutdown(cancel context.CancelFunc, srv *api.Server, conditionScheduler *scheduler.ConditionBasedScheduler, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	logger.Info("Condition-based scheduler shutdown complete")
	os.Exit(0)
}
