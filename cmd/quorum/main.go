package main

import (
	"context"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

func main() {
	// Initialize logger
	if err := logging.InitLogger(logging.Development, "quorum"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger()

	logger.Info("Starting quorum node...")

	ctx := context.Background()

	// Initialize registry
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	// Setup P2P with registry
	config := network.P2PConfig{
		Name:    network.ServiceQuorum,
		Address: "/ip4/0.0.0.0/tcp/9002",
	}

	host, err := network.SetupP2PWithRegistry(ctx, config, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// Initialize discovery service
	discovery := network.NewDiscovery(ctx, host, config.Name)

	// Initialize messaging
	messaging := network.NewMessaging(host, config.Name)
	messaging.InitMessageHandling(func(msg network.Message) {
		logger.Infof("Received message from %s: %+v", msg.From, msg.Content)
	})

	// Try to connect to manager
	if _, err := discovery.ConnectToPeer(network.ServiceManager); err != nil {
		logger.Warnf("Failed to connect to manager: %v", err)
	}

	logger.Infof("Quorum node is running. Node ID: %s", host.ID().String())
	select {}
}
