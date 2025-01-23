package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/execute/manager"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

// Add this as a package-level variable
var (
	db     *database.Connection
	logger logging.Logger
)

func handleJobEvent(event events.JobEvent) {
	logger.Infof("Received job event - Type: %s, JobID: %d, JobType: %d, ChainID: %d",
		event.Type, event.JobID, event.JobType, event.ChainID)

	// Initialize job scheduler with 5 workers
	jobScheduler := manager.NewJobScheduler(5, db)

	switch event.Type {
	case "job_created":
		logger.Infof("New job created: %d", event.JobID)
		// Convert job ID to string and add to scheduler
		jobID := strconv.FormatInt(event.JobID, 10)
		if err := jobScheduler.AddJob(jobID); err != nil {
			logger.Errorf("Failed to add job %s: %v", jobID, err)
			return
		}

		// Print metrics when new job is added
		queueStatus := jobScheduler.GetQueueStatus()
		systemMetrics := jobScheduler.GetSystemMetrics()

		logger.Infof("New job %s added. Current System Status:", jobID)
		logger.Infof("  Job Details: ID=%d, Type=%d, ChainID=%d",
			event.JobID, event.JobType, event.ChainID)
		logger.Infof("  Active Jobs: %d", queueStatus["active_jobs"])
		logger.Infof("  Waiting Jobs: %d", queueStatus["waiting_jobs"])
		logger.Infof("  CPU Usage: %.2f%%", systemMetrics.CPUUsage)
		logger.Infof("  Memory Usage: %.2f%%", systemMetrics.MemoryUsage)

	case "job_updated":
		logger.Infof("Job updated: %d", event.JobID)
		jobScheduler.UpdateJob(event.JobID)

	default:
		logger.Warnf("Unknown event type: %s", event.Type)
	}
}

func subscribeToEvents(ctx context.Context) error {
	eventBus := events.GetEventBus()
	if eventBus == nil {
		return fmt.Errorf("event bus not initialized")
	}

	// Subscribe to the job events channel
	pubsub := eventBus.Redis().Subscribe(ctx, events.JobEventChannel)

	logger.Info("Subscribed to job events channel")

	// Listen for messages in a separate goroutine
	go func() {
		defer pubsub.Close() // Move defer inside the goroutine

		logger.Info("Starting event subscription...")

		// Wait for confirmation of subscription
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

func main() {
	// Initialize logger
	if err := logging.InitLogger(logging.Development, "manager"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger()
	logger.Info("Starting manager node...")

	ctx := context.Background()

	// Initialize event bus
	if err := events.InitEventBus("localhost:6379"); err != nil {
		logger.Fatalf("Failed to initialize event bus: %v", err)
	}

	// Initialize database connection
	dbConfig := &database.Config{
		Hosts:       []string{"localhost"},
		Timeout:     time.Second * 30,
		Retries:     3,
		ConnectWait: time.Second * 20,
	}

	// Assign to the package-level variable
	var err error
	db, err = database.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize registry
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}
	// Setup P2P with registry
	config := network.P2PConfig{
		Name:    network.ServiceManager,
		Address: "/ip4/127.0.0.1/tcp/9000",
	}

	host, err := network.SetupP2PWithRegistry(ctx, config, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// Initialize discovery service
	// discovery := network.NewDiscovery(ctx, host, config.Name)

	// Initialize messaging
	messaging := network.NewMessaging(host, config.Name)
	messaging.InitMessageHandling(func(msg network.Message) {
		logger.Infof("Received message from %s: %+v", msg.From, msg.Content)
	})

	// Subscribe to events
	if err := subscribeToEvents(ctx); err != nil {
		logger.Fatalf("Failed to subscribe to events: %v", err)
	}

	logger.Infof("Manager node is running. Node ID: %s", host.ID().String())
	select {} // Keep the service running
}
