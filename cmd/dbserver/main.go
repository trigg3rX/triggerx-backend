package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gocql/gocql"

	dbserver "github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.ServerService,
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
	obsMetrics := obs.Metrics()

	ctx := context.Background()
	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/4] Dependency: Observability Module Initialised")

	dbConfig := &database.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
		RetryConfig: retry.DefaultRetryConfig(),
	}

	// Configure authentication if provided
	if config.GetDatabaseUsername() != "" && config.GetDatabasePassword() != "" {
		dbConfig.WithAuthentication(config.GetDatabaseUsername(), config.GetDatabasePassword())
	}

	// Configure SSL/TLS if enabled
	if config.GetDatabaseSSLEnabled() {
		dbConfig.WithSSLCertificates(
			config.GetDatabaseSSLCertPath(),
			config.GetDatabaseSSLKeyPath(),
			config.GetDatabaseSSLCAPath(),
			config.GetDatabaseSSLInsecureSkipVerify(),
		)
	}

	conn, err := database.NewConnection(dbConfig, logger)
	if err != nil || conn == nil {
		logger.Fatal(ctx, "Failed to initialize main database connection", observability.Error(err))
	}
	defer conn.Close()

	mainSession := conn.Session()
	if mainSession == nil {
		logger.Fatal(ctx, "Database session cannot be nil")
	}
	logger.Info(ctx, "[2/4] Dependency: Database Connection Initialised")

	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)
	ready := make(chan struct{})

	dockerExecutor, err := dockerexecutor.NewDockerExecutorFromFile("config/services/docker-executor.yaml", logger)
	if err != nil {
		logger.Error(ctx, "Failed to create Docker manager", observability.Error(err))
	} else {
		// Initialize Docker manager with language-specific pools
		if err := dockerExecutor.Initialize(context.Background()); err != nil {
			logger.Error(ctx, "Failed to initialize Docker manager", observability.Error(err))
		} else {
			logger.Info(ctx, "[3/4] Dependency: Docker Executor Initialised")
		}
	}

	tracer := obs.Tracer()
	dbServer := dbserver.NewServer(ctx, conn, logger, tracer, obsMetrics)

	dbServer.RegisterRoutes(ctx, dbServer.GetRouter(), dockerExecutor)
	logger.Info(ctx, "[4/4] Dependency: API server Initialised")

	// Start metrics collector
	collector := metrics.NewCollector(obsMetrics)
	collector.Start()
	logger.Info(ctx, "[1/2] Process: Metrics Collector Started")

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", config.GetHTTPPort()),
		Handler: dbServer.GetRouter(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()
	logger.Info(ctx, "[2/2] Process: HTTP Server Started")

	close(ready)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err := <-serverErrors:
		logger.Error(ctx, "Server error received", observability.Error(err))
	case sig := <-shutdown:
		logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))
	}

	performGracefulShutdown(ctx, srv, &wg, logger, obs, dockerExecutor)
}

func performGracefulShutdown(
	ctx context.Context,
	srv *http.Server,
	wg *sync.WaitGroup,
	logger observability.Logger,
	obs *observability.Observability,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Start shutdown in a goroutine to handle timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "HTTP server shutdown error", observability.Error(err))
			if err := srv.Close(); err != nil {
				logger.Error(shutdownCtx, "Forced HTTP server close error", observability.Error(err))
			}
		}

		if dockerExecutor != nil {
			if err := dockerExecutor.Close(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Failed to close Docker manager", observability.Error(err))
			}
		}

		wg.Wait()

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
