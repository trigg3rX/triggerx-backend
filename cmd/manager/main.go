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

	"github.com/trigg3rX/triggerx-backend/internal/manager"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	kconfig "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

var logger logging.Logger

func main() {
	if err := logging.InitLogger(logging.Development, logging.ManagerProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	logger.Info("Starting manager node...")

	config.Init()
	kconfig.Init()

	var wg sync.WaitGroup

	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	wg.Add(1)

	manager.JobSchedulerInit()
	logger.Info("Job scheduler initialized successfully.")

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.POST("/job/create", manager.HandleCreateJobEvent)
	router.POST("/job/update", manager.HandleUpdateJobEvent)
	router.POST("/job/pause", manager.HandlePauseJobEvent)
	router.POST("/job/resume", manager.HandleResumeJobEvent)
	router.POST("/job/state/update", manager.HandleJobStateUpdate)

	router.POST("/p2p/message", services.ExecuteTask)
	router.POST("/task/validate", services.ValidateTask)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.ManagerRPCPort),
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
	logger.Infof("Manager node is READY on port %s...", config.ManagerRPCPort)

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
