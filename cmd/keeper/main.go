package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net"
	"encoding/json"
	"bytes"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// ConnectToTaskManager sends the keeper's ID and IP to the task manager to establish a connection.
func ConnectToTaskManager(keeperID string, keeperIP string) error {
	// Load the TASK_MANAGER_RPC_ADDRESS from the .env file
	taskManagerRPCAddress := os.Getenv("TASK_MANAGER_RPC_ADDRESS")
	if taskManagerRPCAddress == "" {
		return fmt.Errorf("TASK_MANAGER_RPC_ADDRESS not set in .env file")
	}

	// Construct the payload to send to the task manager
	payload := map[string]string{
		"keeper_id": keeperID,
		"keeper_ip": keeperIP,
	}

	// Marshal the payload into JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request to the task manager
	req, err := http.NewRequest("POST", taskManagerRPCAddress, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content type to JSON
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("task manager returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

// getOutboundIP returns the preferred outbound IP of this machine
func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

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
