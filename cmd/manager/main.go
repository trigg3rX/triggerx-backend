package main

/*
	TODO:
	1. Add P2P message receiver to know what's going with Aggregator
	2. Interlinking Jobs execution Logic

*/

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/internal/manager"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

var (
	db     *database.Connection
	logger logging.Logger
)

// KeeperConnection represents the connection payload from keepers
type KeeperConnection struct {
	KeeperID string `json:"keeper_id"`
	KeeperIP string `json:"keeper_ip"`
}

// handleJobEvent processes incoming job events and delegates to appropriate job scheduler
// based on the job type (time-based, event-based, or condition-based)
func handleJobEvent(event events.JobEvent) {
	logger.Infof("Received job event - Type: %s, JobID: %d",
		event.Type, event.JobID)

	jobScheduler, err := manager.NewJobScheduler(db, logger, network.GetP2PHost())
	if err != nil {
		logger.Errorf("Failed to initialize job scheduler: %v", err)
		return
	}

	switch event.Type {
	case "job_created":
		jobID := event.JobID
		switch event.TaskDefinitionID {
		case 1, 2:
			err := jobScheduler.StartTimeBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		case 3, 4:
			err := jobScheduler.StartEventBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		case 5, 6:
			err := jobScheduler.StartConditionBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		default:
			logger.Warnf("Unknown job type: %d for job: %d", event.TaskDefinitionID, event.JobID)
		}

	case "job_updated":
		logger.Infof("Job updated: %d", event.JobID)

	default:
		logger.Warnf("Unknown event type: %s", event.Type)
	}
}

// subscribeToEvents sets up Redis pub/sub subscription for job events
// and handles incoming messages in a separate goroutine
func subscribeToEvents(ctx context.Context) error {
	eventBus := events.GetEventBus()
	if eventBus == nil {
		return fmt.Errorf("event bus not initialized")
	}

	pubsub := eventBus.Redis().Subscribe(ctx, events.JobEventChannel)

	go func() {
		defer pubsub.Close()

		logger.Info("Starting event subscription...")

		_, err := pubsub.Receive(ctx)
		if err != nil {
			logger.Errorf("Failed to receive subscription confirmation: %v", err)
			return
		}

		logger.Info("Successfully subscribed to job events channel")
		ch := pubsub.Channel()

		for msg := range ch {
			var event events.JobEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				logger.Errorf("Failed to unmarshal event: %v", err)
				continue
			}

			handleJobEvent(event)
		}
	}()

	return nil
}

// handleKeeperConnection handles incoming keeper connection requests
func handleKeeperConnection(c *gin.Context) {
	var keeper KeeperConnection
	if err := c.BindJSON(&keeper); err != nil {
		logger.Error("Failed to parse keeper connection request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Convert keeper_id from string to int64
	keeperID, err := strconv.ParseInt(keeper.KeeperID, 10, 64)
	if err != nil {
		logger.Error("Invalid keeper ID format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper ID format"})
		return
	}

	// Update keeper's connection address in the database
	query := `UPDATE keeper_data SET connection_address = ? WHERE keeper_id = ?`
	if err := db.Session().Query(query, keeper.KeeperIP, keeperID).Exec(); err != nil {
		logger.Error("Failed to update keeper connection address",
			"keeper_id", keeperID,
			"keeper_ip", keeper.KeeperIP,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper data"})
		return
	}

	logger.Info("Keeper connected successfully",
		"keeper_id", keeperID,
		"keeper_ip", keeper.KeeperIP)

	c.JSON(http.StatusOK, gin.H{
		"status":    "connected",
		"keeper_id": keeper.KeeperID,
	})
}

// setupKeeperListener initializes the HTTP server for keeper connections
func setupKeeperListener(ctx context.Context, wg *sync.WaitGroup) error {
	// Get the RPC address from environment
	rpcAddress := os.Getenv("TASK_MANAGER_RPC_ADDRESS")
	if rpcAddress == "" {
		return fmt.Errorf("TASK_MANAGER_RPC_ADDRESS not set in .env file")
	}

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Add keeper connection endpoint
	router.POST("/connect", handleKeeperConnection)

	// Create HTTP server
	srv := &http.Server{
		Addr:    rpcAddress,
		Handler: router,
	}

	// Start server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting keeper connection listener", "address", rpcAddress)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Keeper connection listener failed", "error", err)
		}
	}()

	// Graceful shutdown handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown keeper connection listener gracefully", "error", err)
		}
	}()

	return nil
}

// shutdown performs graceful shutdown of all components:
// cancels context, closes DB connection and P2P host
func shutdown(cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	cancel()

	if db != nil {
		db.Close()
	}

	if err := network.CloseP2PHost(); err != nil {
		logger.Errorf("Failed to close P2P host: %v", err)
	}

	logger.Info("Shutdown complete")
}

// main initializes the manager node, sets up event handling,
// database connections, and handles graceful shutdown
func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	if err := logging.InitLogger(logging.Development, logging.ManagerProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	logger.Info("Starting manager node...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	if err := events.InitEventBus("localhost:6379"); err != nil {
		logger.Errorf("Failed to initialize event bus: %v", err)
	}

	dbConfig := &database.Config{
		Hosts:       []string{"localhost"},
		Timeout:     time.Second * 30,
		Retries:     3,
		ConnectWait: time.Second * 20,
	}

	var err error
	db, err = database.NewConnection(dbConfig)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := subscribeToEvents(ctx); err != nil {
		logger.Errorf("Failed to subscribe to events: %v", err)
	}

	if err := setupKeeperListener(ctx, &wg); err != nil {
		logger.Errorf("Failed to setup keeper listener: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-sigChan:
			logger.Info("Received shutdown signal")
			wg.Add(1)
			go shutdown(cancel, &wg)
		case <-ctx.Done():
			return
		}
	}()

	logger.Info("Connecting to Aggregator...")

	err = network.ConnectToAggregator()
	if err != nil {
		logger.Errorf("Failed to connect to aggregator: %v", err)
	}

	logger.Infof("Manager node is READY.")

	wg.Wait()
}
