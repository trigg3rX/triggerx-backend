package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
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
		ProcessName:   logging.ConditionSchedulerProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Condition-based Scheduler with Redis integration...")

	// Initialize database client
	dbClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerURL())
	if err != nil {
		logger.Fatal("Failed to initialize database client", "error", err)
	}

	// Perform initial health check
	logger.Info("Performing initial health check...")
	if err := dbClient.HealthCheck(); err != nil {
		logger.Warn("Database server health check failed", "error", err)
		logger.Info("Continuing startup - will retry connections during operation")
	} else {
		logger.Info("Database server health check passed")
	}

	// Initialize condition-based scheduler with Redis integration
	managerID := fmt.Sprintf("condition-scheduler-%d", time.Now().Unix())
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize condition-based scheduler", "error", err)
	}
	logger.Info("Condition-based scheduler initialized successfully")

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
		logger.Info("Starting condition monitoring and Redis job creation...")
		conditionScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server for condition job scheduling API...", "port", config.GetSchedulerRPCPort())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Log comprehensive service status
	serviceStatus := map[string]interface{}{
		"manager_id":           managerID,
		"api_port":             config.GetSchedulerRPCPort(),
		"max_workers":          config.GetMaxWorkers(),
		"poll_interval":        "1s",
		"supported_conditions": []string{"greater_than", "less_than", "between", "equals", "not_equals", "greater_equal", "less_equal"},
		"supported_sources":    []string{"api", "oracle", "static"},
		"request_timeout":      "10s",
		"value_cache_ttl":      "30s",
		"condition_state_ttl":  "5m",
		"redis_integration":    "enabled",
		"orchestration_mode":   "redis_job_streams",
		"trigger_mechanism":    "condition_monitoring",
		"task_creation":        "automatic_via_redis",
	}

	logger.Info("Condition-based scheduler service ready", "status", serviceStatus)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(cancel, srv, conditionScheduler, dbClient, logger)
}

func performGracefulShutdown(cancel context.CancelFunc, srv *api.Server, conditionScheduler *scheduler.ConditionBasedScheduler, dbClient *dbserver.DBServerClient, logger logging.Logger) {
	shutdownStart := time.Now()
	logger.Info("Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop()

	// Close database client
	dbClient.Close()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	shutdownDuration := time.Since(shutdownStart)

	logger.Info("Condition-based scheduler shutdown complete",
		"duration", shutdownDuration,
		"redis_integration", "disconnected")
	os.Exit(0)
}
