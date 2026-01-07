package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/repository"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

func main() {
	// Initialize configuration
	configPath := "config/services/time-scheduler.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.TimeSchedulerService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		config.GetServiceID(),
	)

	// Initialize observability (all three pillars)
	obs, err := observability.Initialize(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx := context.Background()
	logger.Info(ctx, "[1/5] Dependency: Observability Module Initialised")

	// Initialize database connection
	dbConfig := database.NewConfig(
		config.GetDatabaseHostAddress(),
		config.GetDatabaseHostPort(),
	)
	dbConfig.Consistency = gocql.Quorum
	dbConfig.Timeout = 10 * time.Second
	dbConfig.Retries = 3
	dbConfig.ConnectWait = 5 * time.Second
	dbConfig.RetryConfig = retry.DefaultRetryConfig()

	// Configure authentication if provided
	if config.GetDatabaseUsername() != "" && config.GetDatabasePassword() != "" {
		dbConfig.WithAuthentication(config.GetDatabaseUsername(), config.GetDatabasePassword())
	}

	// Configure SSL/TLS if enabled
	if config.GetDatabaseSSLEnabled() {
		if config.GetDatabaseSSLCertPath() != "" && config.GetDatabaseSSLKeyPath() != "" {
			dbConfig.WithSSLCertificates(
				config.GetDatabaseSSLCertPath(),
				config.GetDatabaseSSLKeyPath(),
				config.GetDatabaseSSLCAPath(),
				config.GetDatabaseSSLInsecureSkipVerify(),
			)
		} else {
			dbConfig.WithSSL(&tls.Config{
				InsecureSkipVerify: config.GetDatabaseSSLInsecureSkipVerify(),
			})
		}
	}

	dbConn, err := database.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	logger.Info(ctx, "[2/5] Dependency: Database Connection Initialised")

	// Initialize repositories
	timeJobRepo := repository.NewTimeJobRepository(dbConn)
	customJobRepo := repository.NewCustomJobRepository(dbConn)
	scriptStorageRepo := repository.NewScriptStorageRepository(dbConn)
	taskRepo := repository.NewTaskRepository(dbConn)
	logger.Info(ctx, "[3/5] Dependency: Repositories Initialised")

	// Initialize time-based scheduler
	timeScheduler, err := scheduler.NewTimeBasedScheduler(logger, tracer, obsMetrics, timeJobRepo, customJobRepo, scriptStorageRepo, taskRepo)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize time-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[4/5] Dependency: Time Scheduler Initialised")

	// Setup HTTP server with only status endpoint
	srv := api.NewServer(config.GetHTTPPort(), logger)
	logger.Info(ctx, "[5/5] Dependency: Status Server Initialised")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrics collection is started in NewTimeBasedScheduler
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start scheduler in background
	go func() {
		timeScheduler.Start(ctx)
	}()
	logger.Info(ctx, "[2/3] Process: Polling Jobs Started")

	// Start HTTP server
	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: HTTP Server Started", observability.String("port", config.GetHTTPPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, cancel, srv, timeScheduler, dbConn, obs, logger)
}

func performGracefulShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	apiSrv *api.Server,
	timeScheduler *scheduler.TimeBasedScheduler,
	dbConn *database.Connection,
	obs *observability.Observability,
	logger observability.Logger,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GetShutdownTimeout())
	defer shutdownCancel()

	// Cancel context to stop scheduler and HTTP server
	cancel()

	// Stop scheduler gracefully (this also closes the RPC client)
	timeScheduler.Stop(shutdownCtx)
	logger.Info(ctx, "[1/4] Shutdown: Scheduler Stopped")

	// Shutdown API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "[2/4] Shutdown: API server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(ctx, "[2/4] Shutdown: API Server Stopped")
	}

	// Close database connection
	dbConn.Close()
	logger.Info(ctx, "[3/4] Shutdown: Database Connection Closed")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		// Use fmt here since logger is being shut down
		fmt.Printf("Error shutting down observability: %v\n", err)
	}
	fmt.Println("[4/4] Shutdown: Observability Shutdown Complete")

	fmt.Println("Service shutdown completed successfully")
}
