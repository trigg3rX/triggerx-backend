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

	// Start metrics collection
	metrics.StartMetricsCollection()

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.ConditionSchedulerService,
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
	logger.Info(ctx, "Starting Condition-based Scheduler with Redis integration...")

	// Initialize database client
	dbClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerURL())
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database client", observability.Error(err))
	}

	// Perform initial health check
	logger.Info(ctx, "Performing initial health check...")
	if err := dbClient.HealthCheck(); err != nil {
		logger.Warn(ctx, "Database server health check failed", observability.Error(err))
		logger.Info(ctx, "Continuing startup - will retry connections during operation")
	} else {
		logger.Info(ctx, "Database server health check passed")
	}

	// Initialize condition-based scheduler with Redis integration
	managerID := fmt.Sprintf("condition-scheduler-%d", time.Now().Unix())
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize condition-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "Condition-based scheduler initialized successfully")

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
		logger.Info(ctx, "Starting condition monitoring and Redis job creation...")
		conditionScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info(ctx, "Starting HTTP server for condition job scheduling API...", observability.String("port", config.GetSchedulerRPCPort()))
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()

	// Log comprehensive service status
	logger.Info(ctx, "Condition-based scheduler service ready",
		observability.String("manager_id", managerID),
		observability.String("api_port", config.GetSchedulerRPCPort()),
		observability.Int("max_workers", config.GetMaxWorkers()),
		observability.String("poll_interval", "1s"),
		observability.String("redis_integration", "enabled"),
		observability.String("orchestration_mode", "redis_job_streams"),
		observability.String("trigger_mechanism", "condition_monitoring"),
		observability.String("task_creation", "automatic_via_redis"),
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(ctx, cancel, srv, conditionScheduler, dbClient, logger, loggerShutdown)
}

func performGracefulShutdown(ctx context.Context, cancel context.CancelFunc, srv *api.Server, conditionScheduler *scheduler.ConditionBasedScheduler, dbClient *dbserver.DBServerClient, logger observability.Logger, loggerShutdown func(context.Context) error) {
	shutdownStart := time.Now()
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop(ctx)

	// Close database client
	dbClient.Close()

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

	logger.Info(shutdownCtx, "Condition-based scheduler shutdown complete",
		observability.Duration("duration", shutdownDuration),
		observability.String("redis_integration", "disconnected"))
	os.Exit(0)
}
