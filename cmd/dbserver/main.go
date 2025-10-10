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
	// "github.com/gin-gonic/gin"

	dbserver "github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	logConfig := logging.LoggerConfig{
		ProcessName:   logging.DatabaseProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting database server...",
		"mode", config.IsDevMode(),
		"port", config.GetDBServerRPCPort(),
		"host", config.GetDatabaseHostAddress(),
	)

	dbConfig := &connection.Config{
		Hosts:               []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:            "triggerx",
		Consistency:         gocql.Quorum,
		Timeout:             10 * time.Second,
		Retries:             3,
		ConnectWait:         5 * time.Second,
		RetryConfig:         retry.DefaultRetryConfig(),
		ProtoVersion:        4,
		HealthCheckInterval: 15 * time.Second,
		SocketKeepalive:     15 * time.Second,
		MaxPreparedStmts:    1000,
		DefaultIdempotence:  true,
	}

	datastore, err := datastore.NewService(dbConfig, logger)
	if err != nil || datastore == nil {
		logger.Fatalf("Failed to initialize main database connection: %v", err)
	}
	defer datastore.Close()

	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)
	ready := make(chan struct{})

	dockerExecutor, err := dockerexecutor.NewDockerExecutorFromFile("config/docker-executor.yaml", logger)
	if err != nil {
		logger.Errorf("Failed to create Docker manager: %v", err)
	} else {
		// Initialize Docker manager with language-specific pools
		if err := dockerExecutor.Initialize(context.Background()); err != nil {
			logger.Errorf("Failed to initialize Docker manager: %v", err)
		} else {
			logger.Infof("Docker manager initialized successfully")
		}
	}

	dbServer := dbserver.NewServer(datastore, logger)

	dbServer.RegisterRoutes(dbServer.GetRouter(), dockerExecutor)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetDBServerRPCPort()),
		Handler: dbServer.GetRouter(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Database Server initialized, starting on port %s...", config.GetDBServerRPCPort())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(srv, &wg, logger, dockerExecutor)
}

func performGracefulShutdown(srv *http.Server, wg *sync.WaitGroup, logger logging.Logger, dockerExecutor dockerexecutor.DockerExecutorAPI) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	if dockerExecutor != nil {
		if err := dockerExecutor.Close(ctx); err != nil {
			logger.Error("Failed to close Docker manager", "error", err)
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
