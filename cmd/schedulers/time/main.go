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
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TimeSchedulerService,
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
	logger.Info(ctx, "Starting Time-based Scheduler with Redis integration...")

	// Initialize database client
	dbClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerURL())
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database client", observability.Error(err))
	}
	logger.Info(ctx, "Database client initialized successfully")

	// Initialize time-based scheduler with Redis integration via HTTP API
	managerID := fmt.Sprintf("time-scheduler-%d", time.Now().Unix())
	timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize time-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "Time-based scheduler initialized successfully")

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
		logger.Info(ctx, "Starting time-based task polling and Redis submission...")
		timeScheduler.Start(ctx)
	}()

	// Start HTTP server
	go func() {
		logger.Info(ctx, "Starting HTTP server for scheduler management API...", observability.String("port", config.GetSchedulerRPCPort()))
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()

	// Log comprehensive service status
	logger.Info(ctx, "Time-based scheduler service ready",
		observability.String("manager_id", managerID),
		observability.String("api_port", config.GetSchedulerRPCPort()),
		observability.String("poll_interval", config.GetPollingInterval().String()),
		observability.String("look_ahead", config.GetPollingLookAhead().String()),
		observability.Int("batch_size", config.GetTaskBatchSize()),
		observability.String("task_cache_ttl", config.GetTaskCacheTTL().String()),
		observability.String("duplicate_task_window", config.GetDuplicateTaskWindow().String()),
		observability.String("redis_integration", "enabled"),
		observability.String("orchestration_mode", "redis_streams"),
		observability.String("performer_assignment", "automatic_via_redis"),
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-shutdown

	performGracefulShutdown(ctx, cancel, srv, timeScheduler, dbClient, logger, loggerShutdown)
}

func performGracefulShutdown(ctx context.Context, cancel context.CancelFunc, srv *api.Server, timeScheduler *scheduler.TimeBasedScheduler, dbClient *dbserver.DBServerClient, logger observability.Logger, loggerShutdown func(context.Context) error) {
	shutdownStart := time.Now()
	logger.Info(ctx, "Initiating graceful shutdown...")

	// Cancel context to stop scheduler
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Stop scheduler gracefully
	timeScheduler.Stop(ctx)

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

	logger.Info(shutdownCtx, "Time-based scheduler shutdown complete",
		observability.Duration("duration", shutdownDuration),
		observability.String("redis_integration", "disconnected"))
	os.Exit(0)
}
