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

	// "github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	dockerconfig "github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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

	dbConfig := database.NewConfig(config.GetDatabaseHostAddress(), config.GetDatabaseHostPort())

	conn, err := database.NewConnection(dbConfig, logger)
	if err != nil || conn == nil {
		logger.Fatalf("Failed to initialize main database connection: %v", err)
	}
	defer conn.Close()

	mainSession := conn.Session()
	if mainSession == nil {
		logger.Fatalf("Database session cannot be nil")
	}

	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)
	ready := make(chan struct{})

	// Initialize Docker manager with language support
	dockerConfig := dockerconfig.DefaultConfig("go")
	supportedLanguages := []types.Language{
		types.LanguageGo,
		// types.LanguagePy,
		// types.LanguageJS,
		// types.LanguageTS,
		// types.LanguageNode,
	}

	dockerManager, err := docker.NewDockerManager(dockerConfig, logger)
	if err != nil {
		logger.Errorf("Failed to create Docker manager: %v", err)
	} else {
		// Initialize Docker manager with language-specific pools
		if err := dockerManager.Initialize(context.Background(), supportedLanguages); err != nil {
			logger.Errorf("Failed to initialize Docker manager: %v", err)
		} else {
			logger.Infof("Docker manager initialized successfully with %d language pools", len(supportedLanguages))
		}
	}

	dbServer := dbserver.NewServer(conn, logger)

	dbServer.RegisterRoutes(dbServer.GetRouter(), dockerManager)

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

	performGracefulShutdown(srv, &wg, logger, dockerManager)
}

func performGracefulShutdown(srv *http.Server, wg *sync.WaitGroup, logger logging.Logger, dockerManager *docker.DockerManager) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	if dockerManager != nil {
		if err := dockerManager.Close(); err != nil {
			logger.Error("Failed to close Docker manager", "error", err)
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
