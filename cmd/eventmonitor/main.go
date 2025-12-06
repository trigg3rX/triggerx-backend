package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/api"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
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
		ProcessName:   logging.ProcessName("eventmonitor"),
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Event Monitor Service...")

	// Initialize service
	svc, err := service.NewService(logger)
	if err != nil {
		logger.Fatal("Failed to initialize service", "error", err)
	}

	// Start service
	if err := svc.Start(); err != nil {
		logger.Fatal("Failed to start service", "error", err)
	}

	// Setup HTTP server
	srv := api.NewServer(api.Config{
		Port: config.GetPort(),
	}, api.Dependencies{
		Logger:          logger,
		RegistryManager: svc.GetRegistryManager(),
		Service:         svc,
	})

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", "port", config.GetPort())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Log service status
	serviceStatus := map[string]interface{}{
		"port":                config.GetPort(),
		"host":                config.GetHost(),
		"poll_interval":       config.GetPollInterval(),
		"max_block_range":     config.GetMaxBlockRange(),
		"lookback_blocks":     config.GetLookbackBlocks(),
		"webhook_timeout":     config.GetWebhookTimeout(),
		"webhook_max_retries": config.GetWebhookMaxRetries(),
		"version":             "0.1.0-mvp",
	}

	logger.Info("Event Monitor Service ready", "status", serviceStatus)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(srv, svc, logger)
}

func performGracefulShutdown(
	srv *api.Server,
	svc *service.Service,
	logger logging.Logger,
) {
	shutdownStart := time.Now()
	logger.Info("Initiating graceful shutdown...")

	// Stop service
	svc.Stop()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	shutdownDuration := time.Since(shutdownStart)

	logger.Info("Event Monitor Service shutdown complete",
		"duration", shutdownDuration)
	os.Exit(0)
}
