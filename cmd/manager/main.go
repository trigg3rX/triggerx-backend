package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/execute/manager"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libnetwork "github.com/libp2p/go-libp2p/core/network"
)

var (
	db             *database.Connection
	logger         logging.Logger
	quorumState    bool
	validatorState bool
)

func handleJobEvent(event events.JobEvent) {
	logger.Infof("Received job event - Type: %s, JobID: %d, JobType: %d, ChainID: %d",
		event.Type, event.JobID, event.JobType, event.ChainID)

	jobScheduler := manager.NewJobScheduler(5, db)

	switch event.Type {
	case "job_created":
		logger.Infof("New job created: %d", event.JobID)
		jobID := strconv.FormatInt(event.JobID, 10)
		if err := jobScheduler.AddJob(jobID); err != nil {
			logger.Errorf("Failed to add job %s: %v", jobID, err)
			return
		}

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

	pubsub := eventBus.Redis().Subscribe(ctx, events.JobEventChannel)

	logger.Info("Subscribed to job events channel")

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

func shutdown(cancel context.CancelFunc, messaging *network.Messaging, wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	if err := messaging.BroadcastMessage("Manager Shutdown"); err != nil {
		logger.Errorf("Failed to broadcast shutdown message: %v", err)
	}

	time.Sleep(time.Second)

	cancel()

	if db != nil {
		db.Close()
	}

	logger.Info("Shutdown complete")
}

// Add connection state handler
func handleConnectionStateChange(peerID peer.ID, connected bool) {
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Errorf("Failed to load peer registry: %v", err)
		return
	}

	// Check if the peer is quorum or validator
	services := registry.GetAllServices()
	for serviceName, info := range services {
		if info.PeerID == peerID.String() {
			switch serviceName {
			case network.ServiceQuorum:
				quorumState = connected
				// logger.Infof("Quorum connection state changed to: %v", connected)
			case network.ServiceValidator:
				validatorState = connected
				// logger.Infof("Validator connection state changed to: %v", connected)
			}
		}
	}
}

func main() {
	quorumState = false
	validatorState = false

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

	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	host, err := network.SetupServiceWithRegistry(ctx, network.ServiceManager, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// discovery := network.NewDiscovery(ctx, host, network.ServiceManager)

	// Set up connection monitoring using NotifyBundle
	host.Network().Notify(&libnetwork.NotifyBundle{
		ConnectedF: func(n libnetwork.Network, conn libnetwork.Conn) {
			handleConnectionStateChange(conn.RemotePeer(), true)
		},
		DisconnectedF: func(n libnetwork.Network, conn libnetwork.Conn) {
			handleConnectionStateChange(conn.RemotePeer(), false)
		},
	})

	messaging := network.NewMessaging(host, network.ServiceManager)
	messaging.InitMessageHandling(func(msg network.Message) {})

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
			go shutdown(cancel, messaging, &wg)
		case <-ctx.Done():
			return
		}
	}()

	logger.Infof("Manager node is running. Node ID: %s", host.ID().String())

	wg.Wait()
}
