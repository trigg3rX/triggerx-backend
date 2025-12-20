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
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.EventMonitorService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)

	// Create resource for observability
	res, err := observability.NewResource(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create observability resource: %v", err))
	}

	// Initialize logger
	logger, loggerShutdown, err := observability.NewLogger(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	ctx := context.Background()
	logger.Info(ctx, "Starting Event Monitor Service...")

	// Initialize service
	svc, err := service.NewService(ctx, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize service", observability.Error(err))
	}

	// Start service
	if err := svc.Start(); err != nil {
		logger.Fatal(ctx, "Failed to start service", observability.Error(err))
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
		logger.Info(ctx, "Starting HTTP server", observability.String("port", config.GetPort()))
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()

	// Log service status
	logger.Info(ctx, "Event Monitor Service ready",
		observability.String("port", config.GetPort()),
		observability.String("host", config.GetHost()),
		observability.String("poll_interval", config.GetPollInterval().String()),
		observability.Int64("max_block_range", int64(config.GetMaxBlockRange())),
		observability.Int64("lookback_blocks", int64(config.GetLookbackBlocks())),
		observability.String("webhook_timeout", config.GetWebhookTimeout().String()),
		observability.Int("webhook_max_retries", config.GetWebhookMaxRetries()),
		observability.String("version", config.GetVersion()),
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(ctx, srv, svc, logger, loggerShutdown)
}

func performGracefulShutdown(
	ctx context.Context,
	srv *api.Server,
	svc *service.Service,
	logger observability.Logger,
	loggerShutdown func(context.Context) error,
) {
	shutdownStart := time.Now()
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Stop service
	svc.Stop()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Shutdown server gracefully
	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
	}

	// Shutdown logger
	if loggerShutdown != nil {
		if err := loggerShutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
		}
	}

	shutdownDuration := time.Since(shutdownStart)

	logger.Info(shutdownCtx, "Event Monitor Service shutdown complete",
		observability.Duration("duration", shutdownDuration))
	os.Exit(0)
}
