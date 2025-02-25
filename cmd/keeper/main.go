package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	if err := logging.InitLogger(logging.Development, logging.KeeperProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.KeeperProcess)
	logger.Info("Starting keeper node...")

	services.Init()

	keeperAddress := os.Getenv("OPERATOR_ADDRESS")
	
	ip, err := services.GetOutboundIP()
	if err != nil {
		logger.Error("Failed to get outbound IP", "error", err)
		ip = "localhost"
	}
	keeperIP := fmt.Sprintf("%s:%s", ip, os.Getenv("OPERATOR_RPC_PORT"))

	connected, err := services.ConnectToTaskManager(keeperAddress, keeperIP)
	if err != nil {
		logger.Error("Failed to connect to task manager", "error", err)
	}

	if connected {
		logger.Info("Connected to task manager")
	} else {
		logger.Error("Failed to connect to task manager", "error", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.POST("/task/execute", execution.ExecuteTask)
	router.POST("/task/validate", validation.ValidateTask)

	// Custom middleware for error handling
	errorHandler := func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			logger.Error("request failed", "errors", c.Errors)
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}
	}

	router.Use(errorHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("OPERATOR_RPC_PORT")),
		Handler: router,
	}

	// Channel to collect server errors from both goroutines
	serverErrors := make(chan error, 1)

	// Start both servers with automatic recovery
	go func() {
		for {
			logger.Info("Execution Service starting", "address", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrors <- err
				logger.Error("keeper server failed, restarting...", "error", err)
				time.Sleep(time.Second)
				continue
			}
			break
		}
	}()

	// Handle graceful shutdown on interrupt/termination signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("server error received", "error", err)

	case sig := <-shutdown:
		logger.Info("starting shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown keeper server failed",
				"timeout", 2*time.Second,
				"error", err)

			if err := srv.Close(); err != nil {
				logger.Fatal("could not stop keeper server gracefully", "error", err)
			}
		}
	}
	logger.Info("shutdown complete")
}
