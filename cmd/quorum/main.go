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

	"github.com/trigg3rX/triggerx-backend/execute/quorum"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

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

	pubsub := eventBus.Redis().Subscribe(ctx, events.KeeperEventChannel)

	logger.Info("Subscribed to keeper events channel")

	go func() {
		defer pubsub.Close()

		logger.Info("Starting event subscription...")

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

func shutdown(cancel context.CancelFunc, messaging *network.Messaging, wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	if err := messaging.BroadcastMessage("Quorum Shutdown"); err != nil {
		logger.Errorf("Failed to broadcast shutdown message: %v", err)
	}

	time.Sleep(time.Second)

	cancel()

	if db != nil {
		db.Close()
	}

	logger.Info("Shutdown complete")
}

func main() {
	if err := logging.InitLogger(logging.Development, "quorum"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.QuorumProcess)
	logger.Info("Starting quorum node...")

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

	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	host, err := network.SetupServiceWithRegistry(ctx, network.ServiceQuorum, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	discovery := network.NewDiscovery(ctx, host, network.ServiceQuorum)

	_, err = discovery.ConnectToPeer(network.ServiceManager)
	if err != nil {
		logger.Fatalf("Failed to connect to manager: %v", err)
	}
	logger.Infof("Successfully connected to Manager")

	messaging := network.NewMessaging(host, network.ServiceQuorum)
	messaging.InitMessageHandling(func(msg network.Message) {})

	if err := messaging.BroadcastMessage("Quorum Set"); err != nil {
		logger.Errorf("Failed to broadcast initial message: %v", err)
	}

	if err := subscribeToEvents(ctx); err != nil {
		logger.Fatalf("Failed to subscribe to events: %v", err)
	}

	logger.Infof("Quorum node is running. Node ID: %s", host.ID().String())

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

	wg.Wait()
}
