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

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
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

	// Setup a tunnel for the keeper service
	var connectionAddress string
	tunnelURL, err := services.SetupTunnel(config.KeeperRPCPort, config.KeeperAddress)
	if err != nil {
		logger.Error("Failed to setup tunnel", "error", err)
		// Continue with local connection if tunnel fails

		// Try to get public IP for better connectivity
		publicIP, ipErr := services.GetPublicIP()
		if ipErr != nil {
			logger.Error("Failed to get public IP", "error", ipErr)
			// Fall back to the configured IP if public IP retrieval fails
			connectionAddress = fmt.Sprintf("http://%s:%s", config.KeeperIP, config.KeeperRPCPort)
		} else {
			connectionAddress = fmt.Sprintf("http://%s:%s", publicIP, config.KeeperRPCPort)
		}

		logger.Info("Using direct connection address", "address", connectionAddress)
	} else {
		// Use the tunnel URL for connection
		connectionAddress = tunnelURL
		logger.Info("Using tunnel for connection", "url", tunnelURL)
	}

	// Connect to task manager with the tunnel URL or local address as fallback
	connected, err := services.ConnectToTaskManager(config.KeeperAddress, connectionAddress)
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

	// Add health endpoint for keeper verification
	// Note: This is now also handled by the tunnel server for tunnel connections
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":         "healthy",
			"keeper_address": config.KeeperAddress,
			"timestamp":      time.Now().Unix(),
		})
	})

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

		// Close the tunnel if it's active
		services.CloseTunnel()

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
