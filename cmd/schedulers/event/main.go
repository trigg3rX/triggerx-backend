package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler"
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
		ProcessName:   logging.EventSchedulerProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting event-based scheduler service...")

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

	// Initialize event-based scheduler
	managerID := fmt.Sprintf("event-scheduler-%d", time.Now().Unix())
	eventScheduler, err := scheduler.NewEventBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize event-based scheduler", "error", err)
	}

	// Setup HTTP server with scheduler integration
	srv := api.NewServer(api.Config{
		Port: config.GetSchedulerRPCPort(),
	}, api.Dependencies{
		Logger:    logger,
		Scheduler: eventScheduler,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler in background
	go func() {
		logger.Info("Starting event-based scheduler worker management...")
		eventScheduler.Start(ctx)
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
		"manager_id":       managerID,
		"api_port":         config.GetSchedulerRPCPort(),
		"supported_chains": []string{"OP Sepolia (11155420)", "Base Sepolia (84532)", "Ethereum Sepolia (11155111)"},
	}

	logger.Info("Event-based scheduler service ready", serviceStatus)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(cancel, srv, eventScheduler, logger)
}

func performGracefulShutdown(cancel context.CancelFunc, srv *api.Server, eventScheduler *scheduler.EventBasedScheduler, logger logging.Logger) {
	shutdownStart := time.Now()
	logger.Info("Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully (this will stop all job workers)
	eventScheduler.Stop()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	shutdownDuration := time.Since(shutdownStart)

	logger.Info("Event-based scheduler shutdown complete", "duration", shutdownDuration)
	os.Exit(0)
}
