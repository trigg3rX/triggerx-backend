package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/repository"
	conditionrpc "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	configPath := "config/services/condition-scheduler.yaml"
	if err := config.Init(configPath); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
		os.Exit(1)
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
		log.Fatalf("Failed to initialize observability: %v", err)
	}

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx := context.Background()
	logger.Info(ctx, "[1/6] Dependency: Observability Module Initialised")

	// Initialize database connection (using same defaults as time scheduler for now)
	// TODO: Add database config to condition scheduler config
	dbHost := os.Getenv("DATABASE_HOST_ADDRESS")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DATABASE_HOST_PORT")
	if dbPort == "" {
		dbPort = "9042"
	}
	dbConfig := database.NewConfig(dbHost, dbPort)
	dbConfig.Consistency = gocql.Quorum
	dbConfig.Timeout = 10 * time.Second
	dbConfig.Retries = 3
	dbConfig.ConnectWait = 5 * time.Second
	dbConfig.RetryConfig = retry.DefaultRetryConfig()

	// Configure authentication if provided
	if username := os.Getenv("DATABASE_USERNAME"); username != "" {
		if password := os.Getenv("DATABASE_PASSWORD"); password != "" {
			dbConfig.WithAuthentication(username, password)
		}
	}

	// Configure SSL/TLS if enabled
	if os.Getenv("DATABASE_SSL_ENABLED") == "true" {
		if certPath := os.Getenv("DATABASE_SSL_CERT_PATH"); certPath != "" {
			dbConfig.WithSSLCertificates(
				certPath,
				os.Getenv("DATABASE_SSL_KEY_PATH"),
				os.Getenv("DATABASE_SSL_CA_PATH"),
				os.Getenv("DATABASE_SSL_INSECURE_SKIP_VERIFY") == "true",
			)
		} else {
			dbConfig.WithSSL(&tls.Config{
				InsecureSkipVerify: os.Getenv("DATABASE_SSL_INSECURE_SKIP_VERIFY") == "true",
			})
		}
	}

	dbConn, err := database.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	logger.Info(ctx, "[2/6] Dependency: Database Connection Initialised")

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(dbConn)
	logger.Info(ctx, "[3/6] Dependency: Repositories Initialised")

	// Initialize condition-based scheduler
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(logger, tracer, obsMetrics, taskRepo)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize condition-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[4/6] Dependency: Condition Scheduler Initialised")

	// Setup API server (only /status endpoint and event webhook)
	apiPort := "9007" // Different port from RPC server
	apiSrv := api.NewServer(apiPort, logger, conditionScheduler)
	logger.Info(ctx, "[5/6] Dependency: API Server Initialised", observability.String("port", apiPort))

	// Setup RPC server (for schedule/unschedule operations)
	rpcSrv := conditionrpc.NewServer(logger, tracer, conditionScheduler)
	logger.Info(ctx, "[6/6] Dependency: RPC Server Initialised")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrics collection is started in NewConditionBasedScheduler
	logger.Info(ctx, "[1/4] Process: Metrics Collector Started")

	// Start scheduler in background
	go func() {
		conditionScheduler.Start(ctx)
	}()
	logger.Info(ctx, "[2/4] Process: Scheduler Background Job Started")

	// Start API server
	go func() {
		if err := apiSrv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/4] Process: API Server Started", observability.String("port", apiPort))

	// Start RPC server
	go func() {
		if err := rpcSrv.Start(ctx); err != nil {
			logger.Error(ctx, "RPC server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[4/4] Process: RPC Server Started", observability.String("port", config.GetSchedulerRPCPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(cancel, apiSrv, rpcSrv, conditionScheduler, dbConn, obs)
}

func performGracefulShutdown(
	cancel context.CancelFunc,
	apiSrv *api.Server,
	rpcSrv *conditionrpc.Server,
	conditionScheduler *scheduler.ConditionBasedScheduler,
	dbConn *database.Connection,
	obs *observability.Observability,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Cancel context to stop scheduler
	cancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop(shutdownCtx)
	log.Println("[1/5] Shutdown: Scheduler Stopped")

	// Stop RPC server gracefully
	if err := rpcSrv.Stop(shutdownCtx); err != nil {
		log.Fatalf("RPC server forced to shutdown: %v", err)
	}
	log.Println("[2/5] Shutdown: RPC Server Stopped")

	// Stop API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		log.Fatalf("API server forced to shutdown: %v", err)
	}
	log.Println("[3/5] Shutdown: API Server Stopped")

	// Close database connection
	dbConn.Close()
	log.Println("[4/5] Shutdown: Database Connection Closed")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error shutting down observability: %v", err)
	}
	log.Println("[5/5] Shutdown: Observability Shutdown Complete")

	log.Println("Service shutdown completed successfully")
}
