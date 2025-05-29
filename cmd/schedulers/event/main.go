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
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.EventSchedulerProcess,
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

	// Initialize scheduler
	managerID := fmt.Sprintf("scheduler-%d", time.Now().Unix())
	eventScheduler, err := scheduler.NewEventBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal("Failed to initialize time-based scheduler", "error", err)
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
		logger.Info("Starting event-based scheduler...")
		eventScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server...")
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(ctx, cancel, srv, eventScheduler, logger)
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

func performGracefulShutdown(ctx context.Context, cancel context.CancelFunc, srv *api.Server, timeScheduler *scheduler.EventBasedScheduler, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

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

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	logger.Info("Shutdown complete")
	os.Exit(0)
}
