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
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/metrics"
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

	// Initialize tracer
	tracer, tracerShutdown, err := observability.NewTracer(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize tracer: %v", err))
	}

	// Initialize metrics
	obsMetrics, metricsShutdown, err := observability.NewMetrics(obsCfg, res)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize metrics: %v", err))
	}

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)
	
	ctx := context.Background()
	logger.Info(ctx, "[1/3] Dependency: Observability Module Initialised")

	// Initialize service
	svc, err := service.NewService(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize service", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Dependency: Service Initialised")

	// Setup HTTP server
	srv := api.NewServer(api.Config{
		Port: config.GetPort(),
	}, api.Dependencies{
		Logger:          logger,
		RegistryManager: svc.GetRegistryManager(),
		Service:         svc,
	})
	logger.Info(ctx, "[3/3] Dependency: API Server Initialised")

	metrics.StartMetricsCollection()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start service
	if err := svc.Start(); err != nil {
		logger.Fatal(ctx, "Failed to start service", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Process: Service Started")

	// Start HTTP server
	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: HTTP Server Started")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, srv, svc, logger, loggerShutdown, tracerShutdown, metricsShutdown)
}

func performGracefulShutdown(
	ctx context.Context,
	srv *api.Server,
	svc *service.Service,
	logger observability.Logger,
	loggerShutdown func(context.Context) error,
	tracerShutdown func(context.Context) error,
	metricsShutdown func(context.Context) error,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Shutdown tracer
		if tracerShutdown != nil {
			if err := tracerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down tracer", observability.Error(err))
			}
		}

		// Shutdown metrics
		if metricsShutdown != nil {
			if err := metricsShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down metrics", observability.Error(err))
			}
		}

		// Stop service
		svc.Stop()

		// Shutdown server gracefully
		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
		}

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown logger
		if loggerShutdown != nil {
			if err := loggerShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Error shutting down logger", observability.Error(err))
			}
		}
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		// Shutdown completed successfully
	case <-shutdownCtx.Done():
		logger.Warn(ctx, "Shutdown timeout reached, forcing exit")
	}
	os.Exit(0)
}
