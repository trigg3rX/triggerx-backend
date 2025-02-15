package main

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"syscall"
// 	"time"

// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// 	"github.com/trigg3rX/triggerx-backend/pkg/network"
// )

// var logger logging.Logger

// func shutdown(cancel context.CancelFunc, messaging *network.Messaging, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	logger.Info("Starting shutdown sequence...")

// 	if err := messaging.BroadcastMessage("Validator Shutdown"); err != nil {
// 		logger.Errorf("Failed to broadcast shutdown message: %v", err)
// 	}

// 	time.Sleep(time.Second)
// 	cancel()
// 	logger.Info("Shutdown complete")
// }

// func main() {
// 	if err := logging.InitLogger(logging.Development, "validator"); err != nil {
// 		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
// 	}
// 	logger = logging.GetLogger(logging.Development, logging.ValidatorProcess)

// 	logger.Info("Starting validator node...")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	var wg sync.WaitGroup
	
// 	registry, err := network.NewPeerRegistry()
// 	if err != nil {
// 		logger.Fatalf("Failed to initialize peer registry: %v", err)
// 	}

// 	host, err := network.SetupServiceWithRegistry(ctx, network.ServiceValidator, registry)
// 	if err != nil {
// 		logger.Fatalf("Failed to setup P2P: %v", err)
// 	}

// 	discovery := network.NewDiscovery(ctx, host, network.ServiceValidator)

// 	_, err = discovery.ConnectToPeer(network.ServiceManager)
// 	if err != nil {
// 		logger.Fatalf("Failed to connect to manager: %v", err)
// 	}
// 	_, err = discovery.ConnectToPeer(network.ServiceQuorum)
// 	if err != nil {
// 		logger.Fatalf("Failed to connect to quorum: %v", err)
// 	}
// 	logger.Infof("Successfully connected to Services")

// 	messaging := network.NewMessaging(host, network.ServiceValidator)
// 	messaging.InitMessageHandling(func(msg network.Message) {})

// 	if err := messaging.BroadcastMessage("Validator Set"); err != nil {
// 		logger.Errorf("Failed to broadcast initial message: %v", err)
// 	}

// 	logger.Infof("Validator node is running. Node ID: %s", host.ID().String())

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		select {
// 		case <-sigChan:
// 			logger.Info("Received shutdown signal")
// 			wg.Add(1)
// 			go shutdown(cancel, messaging, &wg)
// 		case <-ctx.Done():
// 			return
// 		}
// 	}()

// 	wg.Wait()

// }
