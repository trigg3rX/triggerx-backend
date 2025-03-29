package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	// "os/exec"
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

	// Start Docker containers
	// if err := startDockerContainers(); err != nil {
	// 	logger.Fatal("Failed to start Docker containers", "error", err)
	// }

	services.Init()

	routerValidation := gin.New()
	routerValidation.Use(gin.Recovery())
	routerValidation.Use(gin.Logger())
	
	routerValidation.POST("/p2p/message", execution.ExecuteTask)
	routerValidation.POST("/task/validate", validation.ValidateTask)
	
	routerValidation.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":         "healthy",
			"keeper_address": config.KeeperAddress,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
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

	routerValidation.Use(errorHandler)

	srvValidation := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.KeeperRPCPort),
		Handler: routerValidation,
	}

	// Channel to collect server errors from both goroutines
	serverErrors := make(chan error, 1)

	// Start both servers with automatic recovery
	go func() {
		for {
			logger.Info("Validation Service starting", "address", srvValidation.Addr)
			if err := srvValidation.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
		logger.Error("Server Error Received", "error", err)

	case sig := <-shutdown:
		logger.Info("Starting Shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Run docker-compose down command
		// if err := stopDockerContainers(); err != nil {
		// 	logger.Error("Failed to stop Docker containers", "error", err)
		// }

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Graceful Shutdown Keeper Server Failed",
				"timeout", 2*time.Second,
				"error", err)

			if err := srvValidation.Close(); err != nil {
				logger.Fatal("Could Not Stop Keeper Server Gracefully", "error", err)
			}
		}
	}
	logger.Info("Shutdown Complete")
}

// Function to start Docker containers
// func startDockerContainers() error {
// 	cmd := exec.Command("docker", "compose", "up", "-d")
// 	cmd.Dir = "./"                      // Set the directory where your docker-compose.yaml is located
// 	output, err := cmd.CombinedOutput() // Capture combined output (stdout and stderr)
// 	if err != nil {
// 		return fmt.Errorf("failed to start Docker containers: %v, output: %s", err, output)
// 	}
// 	return nil
// }

// Function to stop Docker containers
// func stopDockerContainers() error {
// 	cmd := exec.Command("docker", "compose", "down")
// 	cmd.Dir = "./"                      // Set the directory where your docker-compose.yaml is located
// 	output, err := cmd.CombinedOutput() // Capture combined output (stdout and stderr)
// 	if err != nil {
// 		return fmt.Errorf("failed to stop Docker containers: %v, output: %s", err, output)
// 	}
// 	return nil
// }
