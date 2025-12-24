package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
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
		observability.ConditionSchedulerService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)

	// Initialize observability (all three pillars)
	obs, err := observability.Initialize(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx := context.Background()
	logger.Info(ctx, "[1/4] Dependency: Observability Module Initialised")

	// Initialize database client
	dbClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerURL())
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database client", observability.Error(err))
	}
	logger.Info(ctx, "[2/4] Dependency: Database Client Initialised")

	// Perform initial health check
	if err := dbClient.HealthCheck(); err != nil {
		logger.Warn(ctx, "Database server health check failed", observability.Error(err))
	}

	// Initialize condition-based scheduler with Redis integration
	managerID := fmt.Sprintf("condition-scheduler-%d", time.Now().Unix())
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(managerID, logger, tracer, obsMetrics, dbClient)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize condition-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[3/4] Dependency: Condition Scheduler Initialised")

	// Setup HTTP server with scheduler integration
	srv := api.NewServer(api.Config{
		Port: config.GetSchedulerRPCPort(),
	}, api.Dependencies{
		Logger:    logger,
		Metrics:   obsMetrics,
		Scheduler: conditionScheduler,
	})
	logger.Info(ctx, "[4/4] Dependency: API Server Initialised")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics.StartMetricsCollection()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start scheduler in background
	go func() {
		conditionScheduler.Start(ctx)
	}()
	logger.Info(ctx, "[2/3] Process: Scheduler Background Job Started")

	// Start HTTP server
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: HTTP Server Started")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, cancel, srv, conditionScheduler, dbClient, logger, obs)
}

func performGracefulShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	srv *api.Server,
	conditionScheduler *scheduler.ConditionBasedScheduler,
	dbClient *dbserver.DBServerClient,
	logger observability.Logger,
	obs *observability.Observability,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Cancel context to stop scheduler
		cancel()

		// Stop scheduler gracefully (this will stop all condition workers)
		conditionScheduler.Stop(ctx)

		// Close database client
		dbClient.Close()

		// Shutdown server gracefully
		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Server forced to shutdown", observability.Error(err))
		}

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability (handles logger, tracer, metrics)
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Error shutting down observability", observability.Error(err))
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
