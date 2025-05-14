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
	config.Init()

	if config.DevMode {
		if err := logging.InitLogger(logging.Development, logging.HealthProcess); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Development, logging.HealthProcess)
	} else {
		if err := logging.InitLogger(logging.Production, logging.HealthProcess); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Production, logging.HealthProcess)
	}
	logger.Info("Starting health node...")

	var wg sync.WaitGroup

	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	wg.Add(1)

	_ = health.GetKeeperStateManager()
	logger.Info("Keeper state manager initialized")

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.POST("/health", health.HandleCheckInEvent)
	router.GET("/status", health.GetKeeperStatus)
	router.GET("/operators", health.GetDetailedKeeperStatus)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "TriggerX Health Service",
			"status":    "running",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HealthRPCPort),
		Handler: router,
	}

	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Health node is READY on port %s...", config.HealthRPCPort)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case <-shutdown:
		logger.Info("Received shutdown signal")
	}

	logger.Info("Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
