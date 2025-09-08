package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.TimeSchedulerProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Time-based Scheduler with Redis integration...")

	// Initialize database client
	dbClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerURL())
	if err != nil {
		logger.Fatal("Failed to initialize database client", "error", err)
	}
	logger.Info("Database client initialized successfully")

	// Initialize time-based scheduler with Redis integration via HTTP API
	managerID := fmt.Sprintf("time-scheduler-%d", time.Now().Unix())
	timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize time-based scheduler", "error", err)
	}
	logger.Info("Time-based scheduler initialized successfully")

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
		logger.Info("Starting time-based task polling and Redis submission...")
		timeScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server for scheduler management API...", "port", config.GetSchedulerRPCPort())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Log comprehensive service status
	serviceStatus := map[string]interface{}{
		"manager_id":            managerID,
		"api_port":              config.GetSchedulerRPCPort(),
		"poll_interval":         config.GetPollingInterval(),
		"look_ahead":            config.GetPollingLookAhead(),
		"batch_size":            config.GetTaskBatchSize(),
		"task_cache_ttl":        config.GetTaskCacheTTL(),
		"duplicate_task_window": config.GetDuplicateTaskWindow(),
		"redis_integration":     "enabled",
		"orchestration_mode":    "redis_streams",
		"performer_assignment":  "automatic_via_redis",
	}

	logger.Info("Time-based scheduler service ready", "status", serviceStatus)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(cancel, srv, timeScheduler, dbClient, logger)
}

func performGracefulShutdown(cancel context.CancelFunc, srv *api.Server, timeScheduler *scheduler.TimeBasedScheduler, dbClient *dbserver.DBServerClient, logger logging.Logger) {
	shutdownStart := time.Now()
	logger.Info("Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully
	timeScheduler.Stop()

	// Close database client
	dbClient.Close()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	shutdownDuration := time.Since(shutdownStart)

	logger.Info("Time-based scheduler shutdown complete",
		"duration", shutdownDuration,
		"redis_integration", "disconnected")
	os.Exit(0)
}
