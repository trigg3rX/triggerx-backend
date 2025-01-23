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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/triggerx-backend/execute/quorum"
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

func handleKeeperEvent(event events.KeeperEvent) {
	logger.Infof("Received keeper event - Type: %s, KeeperID: %d",
		event.Type, event.KeeperID)

	switch event.Type {
	case "keeper_registered":
		logger.Infof("New keeper joined: %d", event.KeeperID)
		status, err := quorum.RegisterKeeper(event.KeeperID)
		if err != nil {
			logger.Errorf("Failed to register keeper: %v", err)
		}
		logger.Infof("Keeper registered: %d, Status: %d", event.KeeperID, status)

	case "keeper_deregistered":
		logger.Infof("Keeper deregistered: %d", event.KeeperID)
		status, err := quorum.DeregisterKeeper(event.KeeperID)
		if err != nil {
			logger.Errorf("Failed to deregister keeper: %v", err)
		}
		logger.Infof("Keeper deregistered: %d, Status: %d", event.KeeperID, status)

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
	pubsub := eventBus.Redis().Subscribe(ctx, events.KeeperEventChannel)

	logger.Info("Subscribed to keeper events channel")

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

		logger.Info("Successfully subscribed to keeper events channel")
		ch := pubsub.Channel()

		for msg := range ch {
			var event events.KeeperEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				logger.Errorf("Failed to unmarshal event: %v", err)
				continue
			}

			handleKeeperEvent(event)
		}
	}()

	return nil
}

func shutdown(cancel context.CancelFunc, messaging *network.Messaging, managerPeerID peer.ID, wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	// Send shutdown message to manager
	if err := messaging.SendMessage(network.ServiceManager, managerPeerID, "Quorum Shutdown"); err != nil {
		logger.Errorf("Failed to send shutdown message to manager: %v", err)
	} else {
		logger.Info("Sent shutdown message to manager")
	}

	// Give some time for the message to be sent
	time.Sleep(time.Second)

	// Cancel the context to signal all goroutines to stop
	cancel()

	// Close database connection
	if db != nil {
		db.Close()
	}

	logger.Info("Shutdown complete")
}

func main() {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Initialize logger
	if err := logging.InitLogger(logging.Development, "quorum"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.QuorumProcess)
	logger.Info("Starting quorum node...")

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

	var err error
	db, err = database.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize registry
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	// Setup P2P with registry
	config := network.P2PConfig{
		Name:    network.ServiceQuorum,
		Address: "/ip4/127.0.0.1/tcp/9001",
	}

	host, err := network.SetupP2PWithRegistry(ctx, config, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// Initialize discovery service
	discovery := network.NewDiscovery(ctx, host, config.Name)

	// Connect to manager service
	managerPeerID, err := discovery.ConnectToPeer(network.ServiceManager)
	if err != nil {
		logger.Fatalf("Failed to connect to manager: %v", err)
	}
	logger.Infof("Successfully connected to manager node: %s", managerPeerID.String())

	// Initialize messaging
	messaging := network.NewMessaging(host, config.Name)
	messaging.InitMessageHandling(func(msg network.Message) {
		logger.Infof("Received message from %s: %+v", msg.From, msg.Content)
	})

	// Send "Quorum Set" message to manager
	if err := messaging.SendMessage(network.ServiceManager, managerPeerID, "Quorum Set"); err != nil {
		logger.Errorf("Failed to send initial message to manager: %v", err)
	} else {
		logger.Info("Sent 'Quorum Set' message to manager")
	}

	// Subscribe to events
	if err := subscribeToEvents(ctx); err != nil {
		logger.Fatalf("Failed to subscribe to events: %v", err)
	}

	logger.Infof("Quorum node is running. Node ID: %s", host.ID().String())

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-sigChan:
			logger.Info("Received shutdown signal")
			wg.Add(1)
			go shutdown(cancel, messaging, managerPeerID, &wg)
		case <-ctx.Done():
			return
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}
