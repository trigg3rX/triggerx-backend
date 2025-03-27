package main

/*
	TODO:
	1. Add P2P message receiver to know what's going with Aggregator
*/

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

	"github.com/trigg3rX/triggerx-backend/internal/manager"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

var logger logging.Logger

func main() {
	if err := logging.InitLogger(logging.Development, logging.ManagerProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	logger.Info("Starting manager node...")

	config.Init()

	var wg sync.WaitGroup

	// Channel to collect setup errors
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	wg.Add(1)

	err := network.ConnectToAggregator()
	if err != nil {
		logger.Fatalf("Failed to connect to aggregator: %v", err)
	} else {
		logger.Info("Connected to aggregator successfully.")
	}

	// Initialize the job scheduler
	manager.JobSchedulerInit()
	logger.Info("Job scheduler initialized successfully.")

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.POST("/job/create", manager.HandleCreateJobEvent)
	router.POST("/job/update", manager.HandleUpdateJobEvent)
	router.POST("/job/pause", manager.HandlePauseJobEvent)
	router.POST("/job/resume", manager.HandleResumeJobEvent)
	router.POST("/keeper/connect", manager.HandleKeeperConnectEvent)
	
	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.ManagerRPCPort),
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
	logger.Infof("Manager node is READY on port %s...", config.ManagerRPCPort)

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