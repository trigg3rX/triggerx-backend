package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func main() {
	// Initialize configuration
	configPath := "config/services/condition-scheduler.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.ConditionSchedulerService,
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
	logger.Info(ctx, "[1/6] Dependency: Observability Module Initialised")

	// Initialize database connection (using same defaults as time scheduler for now)
	dbConfig := database.NewConfig(config.GetDatabaseHostAddress(), config.GetDatabaseHostPort())
	dbConfig.Consistency = gocql.Quorum
	dbConfig.Timeout = config.GetDatabaseTimeout()
	dbConfig.Retries = config.GetDatabaseRetries()
	dbConfig.ConnectWait = config.GetDatabaseConnectWait()
	dbConfig.RetryConfig = retry.DefaultRetryConfig()
	dbConfig.WithAuthentication(config.GetDatabaseUsername(), config.GetDatabasePassword())

	// dbConfig.WithSSLCertificates(
	// 	config.GetDatabaseSSLCertPath(),
	// 	config.GetDatabaseSSLKeyPath(),
	// 	config.GetDatabaseSSLCAPath(),
	// 	config.GetDatabaseSSLInsecureSkipVerify(),
	// )

	dbConn, err := database.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	logger.Info(ctx, "[2/6] Dependency: Database Connection Initialised")

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(dbConn)
	eventJobRepo := repository.NewEventJobRepository(dbConn)
	conditionJobRepo := repository.NewConditionJobRepository(dbConn)
	logger.Info(ctx, "[3/6] Dependency: Repositories Initialised")

	// Initialize condition-based scheduler
	conditionScheduler, err := scheduler.NewConditionBasedScheduler(logger, tracer, obsMetrics, taskRepo, eventJobRepo, conditionJobRepo)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize condition-based scheduler", observability.Error(err))
	}
	logger.Info(ctx, "[4/6] Dependency: Condition Scheduler Initialised")

	// Setup API server (only /status endpoint and event webhook)
	apiPort := config.GetHTTPPort()
	apiSrv := api.NewServer(apiPort, logger, conditionScheduler)
	logger.Info(ctx, "[5/6] Dependency: API Server Initialised", observability.String("port", apiPort))

	// Setup RPC server (for schedule/unschedule operations)
	rpcSrv, err := conditionrpc.NewServer(logger, tracer, conditionScheduler)
	if err != nil {
		logger.Fatal(ctx, "Failed to create RPC server", observability.Error(err))
	}
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
	logger.Info(ctx, "[4/4] Process: RPC Server Started", observability.String("port", config.GetGRPCPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, cancel, apiSrv, rpcSrv, conditionScheduler, dbConn, obs, logger)
}

func performGracefulShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	apiSrv *api.Server,
	rpcSrv *conditionrpc.Server,
	conditionScheduler *scheduler.ConditionBasedScheduler,
	dbConn *database.Connection,
	obs *observability.Observability,
	logger observability.Logger,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GetShutdownTimeout())
	defer shutdownCancel()

	// Cancel context to stop scheduler
	cancel()

	// Stop scheduler gracefully (this will stop all condition workers)
	conditionScheduler.Stop(shutdownCtx)
	logger.Info(ctx, "[1/5] Shutdown: Scheduler Stopped")

	// Stop RPC server gracefully
	if err := rpcSrv.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "[2/5] Shutdown: RPC server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(ctx, "[2/5] Shutdown: RPC Server Stopped")
	}

	// Stop API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "[3/5] Shutdown: API server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(ctx, "[3/5] Shutdown: API Server Stopped")
	}

	// Close database connection
	dbConn.Close()
	logger.Info(ctx, "[4/5] Shutdown: Database Connection Closed")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		// Use fmt here since logger is being shut down
		fmt.Printf("Error shutting down observability: %v\n", err)
	}
	fmt.Println("[5/5] Shutdown: Observability Shutdown Complete")

	fmt.Println("Service shutdown completed successfully")
}
