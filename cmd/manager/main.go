package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

func handleJobEvent(event events.JobEvent) {
	logger.Infof("Received job event - Type: %s, JobID: %d, JobType: %d, ChainID: %d",
		event.Type, event.JobID, event.JobType, event.ChainID)

	jobScheduler, err := manager.NewJobScheduler(db, logger, network.GetP2PHost())
	if err != nil {
		logger.Errorf("Failed to initialize job scheduler: %v", err)
		return
	}

	switch event.Type {
	case "job_created":
		logger.Infof("New job created: %d", event.JobID)
		jobID := event.JobID
		switch event.JobType {
		case 1:
			logger.Infof("Processing Time-Based job: %d", event.JobID)
			err := jobScheduler.StartTimeBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		case 2:
			logger.Infof("Processing Event-Based job: %d", event.JobID)
			err := jobScheduler.StartEventBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		case 3:
			logger.Infof("Processing Condition-Based job: %d", event.JobID)
			err := jobScheduler.StartConditionBasedJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		case 4:
			logger.Infof("Processing Custom Script job: %d", event.JobID)
			err := jobScheduler.StartCustomScriptJob(jobID)
			if err != nil {
				logger.Errorf("Failed to add job %s: %v", jobID, err)
			}
		default:
			logger.Warnf("Unknown job type: %d for job: %d", event.JobType, event.JobID)
		}

	case "job_updated":
		logger.Infof("Job updated: %d", event.JobID)

	default:
		logger.Warnf("Unknown event type: %s", event.Type)
	}
}

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

func main() {
	if err := logging.InitLogger(logging.Development, logging.ManagerProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	logger.Info("Starting manager node...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	if err := events.InitEventBus("localhost:6379"); err != nil {
		logger.Fatalf("Failed to initialize event bus: %v", err)
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
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := subscribeToEvents(ctx); err != nil {
		logger.Fatalf("Failed to subscribe to events: %v", err)
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
		logger.Fatalf("Failed to connect to aggregator: %v", err)
	}

	logger.Infof("Manager node is READY.")

	wg.Wait()
}
