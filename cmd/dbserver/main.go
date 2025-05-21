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

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.DatabaseProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting database server...",
		"mode", getEnvironment(),
		"port", config.GetDatabaseRPCPort(),
		"host", config.GetDatabaseHost(),
	)

	dbConfig := database.NewConfig(config.GetDatabaseHost(), config.GetDatabaseHostPort())

	conn, err := database.NewConnection(dbConfig)
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

	dbServer := dbserver.NewServer(conn, logging.DatabaseProcess)

	srv := setupHTTPServer(logger, dbServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Database Server initialized, starting on port %s...", config.GetDatabaseRPCPort())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(srv, &wg, logger)
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

func getLogLevel() logging.Level {
	if config.IsDevMode() {
		return logging.DebugLevel
	}
	return logging.InfoLevel
}

func setupHTTPServer(logger logging.Logger, dbServer *dbserver.Server) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	dbServer.RegisterRoutes(router)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetDatabaseRPCPort()),
		Handler: router,
	}
}

func performGracefulShutdown(srv *http.Server, wg *sync.WaitGroup, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
