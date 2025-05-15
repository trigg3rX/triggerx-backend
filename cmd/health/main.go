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

	"github.com/trigg3rX/triggerx-backend/internal/health"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:      logging.BaseDataDir,
		ProcessName: logging.HealthProcess,
		Environment: getEnvironment(),
		UseColors:   true,
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting health service...")

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	// Initialize state manager
	_ = health.InitializeStateManager(logger)
	logger.Info("Keeper state manager initialized")

	// Setup HTTP server
	srv := setupHTTPServer(logger)

	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Health service is ready on port %s", config.GetHealthRPCPort())

	// Handle graceful shutdown
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

func setupHTTPServer(logger logging.Logger) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(health.LoggerMiddleware(logger))

	// Register routes
	health.RegisterRoutes(router)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetHealthRPCPort()),
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

	wg.Wait()
	logger.Info("Shutdown complete")
}
