package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
    "github.com/ethereum/go-ethereum/crypto"
	eigensdklogging "github.com/Layr-Labs/eigensdk-go/logging"
)

var (
	logger logging.Logger
	// managerState bool
	// quorumState  bool
	// validatorState bool
)

func shutdown(cancel context.CancelFunc, messaging *network.Messaging, managerPeerID peer.ID, wg *sync.WaitGroup, keeperName string) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	shutdownMsg := fmt.Sprintf("%s left the network", keeperName)
	if err := messaging.SendMessage(network.ServiceQuorum, quorumPeerID, shutdownMsg, false); err != nil {
		logger.Errorf("Failed to send shutdown message to quorum: %v", err)
	}

	time.Sleep(time.Second)

	if cancel != nil {
		cancel()
	}

	logger.Info("Shutdown complete")

}
func main() {
	if err := logging.InitLogger(logging.Development, "keeper"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.KeeperProcess)
	logger.Info("Starting keeper node...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

		// Load configuration
	yamlFile, err := os.ReadFile("config-files/triggerx_keeper.yaml")
	if err != nil {
		logger.Fatalf("Error reading YAML file: %v", err)
	}
	
	var config types.NodeConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		logger.Fatalf("Error parsing YAML: %v", err)
	}
	
		// Load private key from keystore
	ecdsaPrivateKey, err := metrics.LoadPrivateKeyFromKeystore(config.EcdsaPrivateKeyStorePath, config.EcdsaPassphrase)
	if err != nil {
		logger.Fatalf("Failed to load ECDSA private key: %v", err)
	}
	
	operatorAddr := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	logger.Info("Operator address", "address", operatorAddr.Hex())
	eigensdkLogger, err := eigensdklogging.NewZapLogger("development")
	// Initialize metrics service if enabled
	if config.EnableMetrics {
		metricsConfig := &metrics.MetricsConfig{
			AvsName:                    config.AvsName,
			EthRpcUrl:                 config.EthRpcUrl,
			EthWsUrl:                  config.EthWsUrl,
			RegistryCoordinatorAddress:    config.RegistryCoordinatorAddress,
			OperatorStateRetrieverAddress: config.OperatorStateRetrieverAddress,
		}
	
		metricsService, err := metrics.NewMetricsService(
			eigensdkLogger,
			ecdsaPrivateKey,
			operatorAddr,
			metricsConfig,
		)
		if err != nil {
			logger.Fatalf("Failed to initialize metrics service: %v", err)
		}
	
		// Start metrics service
		if err := metricsService.Start(ctx); err != nil {
			logger.Fatalf("Failed to start metrics service: %v", err)
		}
		logger.Info("Metrics service started successfully")
	}
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	host, err := network.SetupKeeperWithRegistry(ctx, config, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	messaging := network.NewMessaging(host, config.KeeperName)
	messaging.InitMessageHandling(func(msg network.Message) {})

	discovery := network.NewDiscovery(ctx, host, config.KeeperName)
	if err := discovery.SavePeerInfo(); err != nil {
		logger.Fatalf("Failed to save peer info: %v", err)
	}

	quorumPeerID, err := discovery.ConnectToPeer(network.ServiceQuorum)
	if err != nil {
		logger.Fatalf("Failed to connect to quorum: %v", err)
	}
	logger.Infof("Successfully connected to Quorum")

	services := []string{network.ServiceManager, network.ServiceValidator}
	for _, service := range services {
		_, err = discovery.ConnectToPeer(service)
		if err != nil {
			logger.Warnf("Failed to connect to %s: %v", service, err)
			continue
		}
		logger.Infof("Successfully connected to %s", service)
	}

	message := fmt.Sprintf("%s joined the network", config.KeeperName)
	if err := messaging.SendMessage(network.ServiceQuorum, quorumPeerID, message, false); err != nil {
		logger.Errorf("Failed to send initial message to quorum: %v", err)
	}
	
	logger.Infof("Keeper node is running. Node ID: %s", host.ID().String())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-sigChan:
			logger.Info("Received shutdown signal")
			wg.Add(1)
			go shutdown(cancel, messaging, quorumPeerID, &wg, config.KeeperName)
		case <-ctx.Done():
			return
		}
	}()

	wg.Wait()
}
