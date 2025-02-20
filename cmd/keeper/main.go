package main

/*
	DEV NOTES:
	1. The attestation code is not being used, as we yet have to figure out how to allow keepers to make contract calls. How to fund them, etc.
	2. So, for the time being, we run the keepers, and attesters validate their actions.
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"

	// "github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

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

	// Generate or load keeper ID
	keeperID := os.Getenv("KEEPER_ID")
	if keeperID == "" {
		keeperID = fmt.Sprintf("keeper_%d", time.Now().UnixNano())
	}

	// Get the local IP address using getOutboundIP
	ip, err := getOutboundIP()
	if err != nil {
		logger.Error("Failed to get outbound IP", "error", err)
		ip = "localhost" // fallback to localhost if IP detection fails
	}
	keeperIP := fmt.Sprintf("%s:4003", ip)

	// Connect to task manager
	if err := ConnectToTaskManager(keeperID, keeperIP); err != nil {
		logger.Error("Failed to connect to task manager", "error", err)
		// Continue execution as the keeper might want to retry connection later
	}

	// Set up performer server using Gin
	performerRouter := gin.New()
	performerRouter.Use(gin.Recovery())
	performerRouter.Use(gin.Logger())
	performerRouter.POST("/task/execute", execution.ExecuteTask)

	// // Set up attester server using Gin
	// attesterRouter := gin.New()
	// attesterRouter.Use(gin.Recovery())
	// attesterRouter.Use(gin.Logger())
	// attesterRouter.POST("/task/validate", validation.ValidateTask)

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

	performerRouter.Use(errorHandler)
	// attesterRouter.Use(errorHandler)

	performerSrv := &http.Server{
		Addr:    ":4003",
		Handler: performerRouter,
	}

	// attesterSrv := &http.Server{
	// 	Addr:    ":4002",
	// 	Handler: attesterRouter,
	// }

	// Channel to collect server errors from both goroutines
	serverErrors := make(chan error, 2)

	// Start both servers with automatic recovery
	go func() {
		for {
			logger.Info("Execution Service starting", "address", performerSrv.Addr)
			if err := performerSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrors <- err
				logger.Error("performer server failed, restarting...", "error", err)
				time.Sleep(time.Second)
				continue
			}
			break
		}
	}()

	// go func() {
	// 	for {
	// 		logger.Info("Validation Server starting", "address", attesterSrv.Addr)
	// 		if err := attesterSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 			serverErrors <- err
	// 			logger.Error("attester server failed, restarting...", "error", err)
	// 			time.Sleep(time.Second)
	// 			continue
	// 		}
	// 		break
	// 	}
	// }()

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

		if err := performerSrv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown performer server failed",
				"timeout", 2*time.Second,
				"error", err)

			if err := performerSrv.Close(); err != nil {
				logger.Fatal("could not stop performer server gracefully", "error", err)
			}
		}

		// if err := attesterSrv.Shutdown(ctx); err != nil {
		// 	logger.Error("graceful shutdown attester server failed",
		// 		"timeout", 2*time.Second,
		// 		"error", err)

		// 	if err := attesterSrv.Close(); err != nil {
		// 		logger.Fatal("could not stop attester server gracefully", "error", err)
		// 	}
		// }
	}
	logger.Info("shutdown complete")
}
