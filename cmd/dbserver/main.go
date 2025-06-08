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
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	logConfig := logging.LoggerConfig{
		ProcessName:     logging.DatabaseProcess,
		IsDevelopment:   config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting database server...",
		"mode", config.IsDevMode(),
		"port", config.GetDatabaseRPCPort(),
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

	dbServer := dbserver.NewServer(conn, logger)

	dbServer.RegisterRoutes(dbServer.GetRouter())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetDatabaseRPCPort()),
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

	wg.Wait()
	logger.Info("Shutdown complete")
}
