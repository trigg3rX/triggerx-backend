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
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/internal/loadbalancer"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger logging.Logger

func main() {
	if err := logging.InitLogger(logging.Development, logging.LoadBalancerProcessType); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.LoadBalancerProcessType)
	logger.Info("Starting load balancer...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found, using environment variables")
	}

	// Get Redis configuration
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis" // Default Redis host
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379" // Default Redis port
	}
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// Initialize load balancer
	lb, err := loadbalancer.NewLoadBalancer(redisAddress)
	if err != nil {
		logger.Fatalf("Failed to initialize load balancer: %v", err)
	}

	// Start load balancer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := lb.Start(ctx); err != nil {
		logger.Fatalf("Failed to start load balancer: %v", err)
	}

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// Register task manager endpoint
	router.POST("/register", func(c *gin.Context) {
		var tm loadbalancer.TaskManager
		if err := c.BindJSON(&tm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		if err := lb.RegisterTaskManager(ctx, &tm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register task manager: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Task manager registered successfully"})
	})

	// Job creation endpoint
	router.POST("/job/create", func(c *gin.Context) {
		// Select a task manager
		tm, err := lb.SelectTaskManager()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available task managers"})
			return
		}

		// Forward the request to the selected task manager
		jobURL := fmt.Sprintf("http://%s/job/create", tm.Address)
		resp, err := http.Post(jobURL, "application/json", c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to forward request: %v", err)})
			return
		}
		defer resp.Body.Close()

		// Forward the response back to the client
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Received shutdown signal")

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

	logger.Info("Shutdown complete")
}
