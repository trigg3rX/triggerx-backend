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

	dbConfig := &database.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
		RetryConfig: retry.DefaultRetryConfig(),
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

	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)
	ready := make(chan struct{})

	dockerExecutor, err := dockerexecutor.NewDockerExecutorFromFile("config/docker-executor.yaml", logger)
	if err != nil {
		logger.Error(ctx, "Failed to create Docker manager", observability.Error(err))
	} else {
		// Initialize Docker manager with language-specific pools
		if err := dockerExecutor.Initialize(context.Background()); err != nil {
			logger.Error(ctx, "Failed to initialize Docker manager", observability.Error(err))
		} else {
			logger.Info(ctx, "Docker manager initialized successfully")
		}
	}

	dbServer := dbserver.NewServer(ctx, conn, logger)

	dbServer.RegisterRoutes(ctx, dbServer.GetRouter(), dockerExecutor)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetDBServerRPCPort()),
		Handler: dbServer.GetRouter(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info(ctx, "Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Info(ctx, "Database Server initialized, starting on port %s...", observability.String("port", config.GetDBServerRPCPort()))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error(ctx, "Server error received", observability.Error(err))
	case sig := <-shutdown:
		logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))
	}

	performGracefulShutdown(ctx, srv, &wg, logger, loggerShutdown, dockerExecutor)
}

func performGracefulShutdown(ctx context.Context, srv *http.Server, wg *sync.WaitGroup, logger observability.Logger, loggerShutdown func(context.Context) error, dockerExecutor dockerexecutor.DockerExecutorAPI) {
	logger.Info(ctx, "Initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP server shutdown error", observability.Error(err))
		if err := srv.Close(); err != nil {
			logger.Error(ctx, "Forced HTTP server close error", observability.Error(err))
		}
	}

	if dockerExecutor != nil {
		if err := dockerExecutor.Close(ctx); err != nil {
			logger.Error(ctx, "Failed to close Docker manager", observability.Error(err))
		}
	}

	wg.Wait()

	if loggerShutdown != nil {
		if err := loggerShutdown(ctx); err != nil {
			logger.Error(ctx, "Error shutting down logger", observability.Error(err))
		} else {
			logger.Info(ctx, "Logger closed successfully")
		}
	}

	logger.Info(ctx, "Shutdown complete")
}
