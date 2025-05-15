package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/manager"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/internal/manager/ha"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	kconfig "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

var logger logging.Logger
var isLeaderManager bool

func main() {
	if err := logging.InitLogger(logging.Development, logging.ManagerProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	logger.Info("Starting manager node...")

	config.Init()
	kconfig.Init()

	var wg sync.WaitGroup

	// Channel to collect setup errors
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	// Check if HA mode is enabled
	haEnabled := os.Getenv("MANAGER_HA_ENABLED") == "true"

	if haEnabled {
		// Initialize high availability
		logger.Info("Initializing manager in high availability mode")

		redisHost := os.Getenv("REDIS_HOST")
		if redisHost == "" {
			redisHost = "redis" // Default Redis host
		}
		redisPort := os.Getenv("REDIS_PORT")
		if redisPort == "" {
			redisPort = "6379" // Default Redis port
		}
		redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

		otherManagers := []string{}
		if otherManagersStr := os.Getenv("OTHER_MANAGER_ADDRESSES"); otherManagersStr != "" {
			otherManagers = strings.Split(otherManagersStr, ",")
		}

		haConfig := &ha.HAConfig{
			RedisAddress:          redisAddress,
			RedisPassword:         os.Getenv("REDIS_PASSWORD"),
			OtherManagerAddresses: otherManagers,
			OnRoleChange:          handleRoleChange,
		}

		ha.Init(haConfig)

		// Wait a moment for initial role election
		time.Sleep(2 * time.Second)
		isLeaderManager = ha.IsLeader()
		logger.Infof("Initial manager role: %s", getRoleString())
	} else {
		// In non-HA mode, this manager is always the leader
		isLeaderManager = true
		logger.Info("Running in standalone mode (high availability disabled)")
	}

	// Initialize the job scheduler
	manager.JobSchedulerInit()
	logger.Info("Job scheduler initialized successfully.")

	wg.Add(1)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Add a middleware that only allows leader to process job management endpoints
	router.Use(leaderOnlyMiddleware())

	router.POST("/job/create", manager.HandleCreateJobEvent)
	router.POST("/job/update", manager.HandleUpdateJobEvent)
	router.POST("/job/pause", manager.HandlePauseJobEvent)
	router.POST("/job/resume", manager.HandleResumeJobEvent)
	router.POST("/job/state/update", manager.HandleJobStateUpdate)

	// P2P and task validation endpoints are available on all instances
	router.POST("/p2p/message", services.ExecuteTask)
	router.POST("/task/validate", services.ValidateTask)

	// Add health check endpoint
	router.GET("/health", handleHealthCheck)

	// Add manager status endpoint
	router.GET("/status", handleManagerStatus)

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
	logger.Infof("Manager node is READY on port %s with role %s...", config.ManagerRPCPort, getRoleString())

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

// Handle role changes between leader and follower
func handleRoleChange(newRole string) {
	isLeaderManager = (newRole == ha.RoleLeader)
	logger.Infof("Manager role changed to: %s", getRoleString())
}

// Get readable role string
func getRoleString() string {
	if isLeaderManager {
		return "LEADER"
	}
	return "FOLLOWER"
}

// Middleware to ensure only leader processes certain requests
func leaderOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for certain paths that all instances can handle
		path := c.Request.URL.Path
		if path == "/health" || path == "/status" || path == "/p2p/message" || path == "/task/validate" {
			c.Next()
			return
		}

		// Check if this instance is the leader
		if !isLeaderManager {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "This manager instance is not the leader",
				"message": "Request redirected to follower instance",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Health check endpoint
func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"instance":  ha.GetInstanceID(),
		"role":      getRoleString(),
		"timestamp": time.Now().UTC(),
	})
}

// Manager status endpoint
func handleManagerStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"instance":   ha.GetInstanceID(),
		"role":       getRoleString(),
		"timestamp":  time.Now().UTC(),
		"is_leader":  isLeaderManager,
		"ha_enabled": os.Getenv("MANAGER_HA_ENABLED") == "true",
	})
}
