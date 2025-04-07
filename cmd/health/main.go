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

var logger logging.Logger

func main() {
	if err := logging.InitLogger(logging.Development, logging.HealthProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.HealthProcess)
	logger.Info("Starting health node...")

	config.Init()

	var wg sync.WaitGroup

	// Channel to collect setup errors
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	wg.Add(1)

	// Initialize the keeper state manager
	_ = health.GetKeeperStateManager()
	logger.Info("Keeper state manager initialized")

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.POST("/health", health.HandleCheckInEvent)
	router.GET("/status", health.GetKeeperStatus)

	// Add a simple health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "TriggerX Health Service",
			"status":    "running",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HealthRPCPort),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	// Signal server is ready
	close(ready)
	logger.Infof("Health node is READY on port %s...", config.HealthRPCPort)

	// Handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case <-shutdown:
		logger.Info("Received shutdown signal")
	}

	// Begin graceful shutdown
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	logger.Info("Shutdown complete")
}
