package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	logger logging.Logger
)

func shutdown(cancel context.CancelFunc, messaging *network.Messaging, managerPeerID peer.ID, wg *sync.WaitGroup, keeperName string) {
	defer wg.Done()

	logger.Info("Starting shutdown sequence...")

	// Send shutdown message to manager
	shutdownMsg := fmt.Sprintf("%s Left the network", keeperName)
	if err := messaging.SendMessage(network.ServiceManager, managerPeerID, shutdownMsg); err != nil {
		logger.Errorf("Failed to send shutdown message to manager: %v", err)
	} else {
		logger.Info("Sent shutdown message to manager")
	}

	// Give some time for the message to be sent
	time.Sleep(time.Second)

	// Cancel the context to signal all goroutines to stop
	cancel()

	logger.Info("Shutdown complete")
}

func main() {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to track goroutines
	var wg sync.WaitGroup

	if err := logging.InitLogger(logging.Development, "keeper"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.KeeperProcess)

	yamlFile, err := os.ReadFile("config-files/triggerx_keeper.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		os.Exit(1)
	}

	var config types.NodeConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	// Update the addresses in the registry with the actual server IP
	for serviceType, info := range registry.GetAllServices() {
		if serviceType != network.ServiceKeeper {
			newAddrs := make([]string, len(info.Addresses))
			for i, addr := range info.Addresses {
				newAddrs[i] = strings.Replace(addr, "127.0.0.1", config.ServerIpAddress, 1)
			}
			info.Addresses = newAddrs
			if err := registry.UpdateService(serviceType, peer.ID(info.PeerID), newAddrs); err != nil {
				logger.Fatalf("Failed to update registry addresses: %v", err)
			}
		}
	}

	p2pconfig := network.P2PConfig{
		Name:    network.ServiceKeeper,
		Address: fmt.Sprintf("/ip4/%s/tcp/%s", config.ConnectionAddress, config.P2pPort),
	}

	host, err := network.SetupP2PWithRegistry(ctx, p2pconfig, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// Initialize messaging
	messaging := network.NewMessaging(host, p2pconfig.Name)
	messaging.InitMessageHandling(func(msg network.Message) {
		logger.Infof("Received message from %s: %+v", msg.From, msg.Content)
	})

	// Initialize discovery and attempt connections in order
	discovery := network.NewDiscovery(ctx, host, p2pconfig.Name)

	// Connect to manager service first
	managerPeerID, err := discovery.ConnectToPeer(network.ServiceManager)
	if err != nil {
		logger.Fatalf("Failed to connect to manager: %v", err)
	}
	logger.Infof("Successfully connected to manager node: %s", managerPeerID.String())

	// Send join message to manager
	joinMsg := fmt.Sprintf("%s joined the network", config.KeeperName)
	if err := messaging.SendMessage(network.ServiceManager, managerPeerID, joinMsg); err != nil {
		logger.Errorf("Failed to send join message to manager: %v", err)
	} else {
		logger.Info("Sent join message to manager")
	}

	// Try connecting to other services: quorum -> validator
	services := []string{network.ServiceQuorum, network.ServiceValidator}
	for _, service := range services {
		peerID, err := discovery.ConnectToPeer(service)
		if err != nil {
			logger.Warnf("Failed to connect to %s: %v", service, err)
			continue
		}
		logger.Infof("Successfully connected to %s (PeerID: %s)", service, peerID.String())
	}

	logger.Info("Starting keeper node...")
	logger.Infof("Keeper node is running. Node ID: %s", host.ID().String())
	logger.Infof("Listening on addresses: %v", host.Addrs())

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
			go shutdown(cancel, messaging, managerPeerID, &wg, config.KeeperName)
		case <-ctx.Done():
			return
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"strconv"
// 	"syscall"

// 	"github.com/joho/godotenv"
// 	"github.com/libp2p/go-libp2p"
// 	"github.com/libp2p/go-libp2p/core/peer"
// 	"github.com/multiformats/go-multiaddr"

// 	"github.com/trigg3rX/triggerx-backend/execute/keeper/handler"
// 	"github.com/trigg3rX/triggerx-backend/execute/manager"
// 	"github.com/trigg3rX/triggerx-backend/pkg/network"

// 	"github.com/ethereum/go-ethereum/ethclient"
// )

// func main() {
// 	// Setup logging
// 	log.SetOutput(os.Stdout)
// 	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

// 	// Create context
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	// Create libp2p host
// 	host, err := libp2p.New()
// 	if err != nil {
// 		log.Fatalf("Failed to create libp2p host: %v", err)
// 	}
// 	defer host.Close()

// 	// Create network messaging
// 	keeperName := "node1"
// 	messaging := network.NewMessaging(host, keeperName)
// 	err = godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatalf("Error loading .env file: %s", err)
// 	}
// 	alchemyAPIKey := os.Getenv("ALCHEMY_API_KEY")
// 	ethClient, err := ethclient.Dial(fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", alchemyAPIKey))
// 	if err != nil {
// 		log.Fatalf("Failed to create Ethereum client: %v", err)
// 	}
// 	defer ethClient.Close()

// 	etherscanAPIKey := os.Getenv("ETHERSCAN_API_KEY")
// 	if etherscanAPIKey == "" {
// 		log.Fatalf("ETHERSCAN_API_KEY is required")
// 	}
// 	jobHandler := handler.NewJobHandler(ethClient, etherscanAPIKey)

// 	// Setup message handling
// 	messaging.InitMessageHandling(func(msg network.Message) {
// 		// Check if it's a JOB_TRANSMISSION message
// 		if nestedContent, ok := msg.Content.(map[string]interface{})["content"]; ok {
// 			// Type assert the nested content
// 			jobMap, ok := nestedContent.(map[string]interface{})
// 			if !ok {
// 				log.Printf("Invalid job content type")
// 				return
// 			}

// 			// Convert map to Job struct
// 			jobData, err := convertMapToJob(jobMap)
// 			if err != nil {
// 				log.Printf("Job conversion error: %v", err)
// 				return
// 			}

// 			// Print the job in a formatted way
// 			log.Printf("Received Job:")
// 			log.Printf("Job ID: %s", jobData.JobID)
// 			log.Printf("Chain ID: %s", jobData.ChainID)
// 			log.Printf("Contract Address: %s", jobData.ContractAddress)
// 			log.Printf("Target Function: %s", jobData.TargetFunction)
// 			log.Printf("Status: %s", jobData.Status)
// 			log.Printf("Arguments: %+v", jobData.Arguments)
// 			log.Printf("Max Retries: %d", jobData.MaxRetries)
// 			log.Printf("Current Retries: %d", jobData.CurrentRetries)
// 			log.Printf("CodeURL: %s", jobData.CodeURL)

// 			// Handle job
// 			if err := jobHandler.HandleJob(jobData); err != nil {
// 				log.Printf("Job handling error: %v", err)
// 			}
// 		} else {
// 			log.Printf("Received non-job message: %+v", msg.Type)
// 		}
// 	})

// 	// Save peer info for discovery
// 	discovery := network.NewDiscovery(ctx, host, keeperName)
// 	if err := discovery.SavePeerInfo(); err != nil {
// 		log.Printf("Failed to save peer info: %v", err)
// 	}

// 	log.Println("Keeper addresses:", host.Addrs())

// 	peerInfos, err := network.LoadPeerInfo()
// 	if err != nil {
// 		log.Printf("Error loading peer info: %v", err)
// 	}

// 	for name, info := range peerInfos {
// 		if name != keeperName {
// 			// Attempt to connect to other known peers
// 			maddr, err := multiaddr.NewMultiaddr(info.Address)
// 			if err == nil {
// 				peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
// 				if err == nil {
// 					host.Connect(ctx, *peerInfo)
// 				}
// 			}
// 		}
// 	}

// 	// Wait for interrupt
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// 	log.Printf("🚀 Keeper node started. Listening on %s", host.Addrs())
// 	<-sigChan
// }

// // Helper function to convert map to Job struct
// func convertMapToJob(jobMap map[string]interface{}) (*manager.Job, error) {
// 	job := &manager.Job{
// 		JobID:           toString(jobMap["JobID"]),
// 		ArgType:         toString(jobMap["ArgType"]),
// 		ChainID:         toString(jobMap["ChainID"]),
// 		ContractAddress: toString(jobMap["ContractAddress"]),
// 		TargetFunction:  toString(jobMap["TargetFunction"]),
// 		Status:          toString(jobMap["Status"]),
// 		UserID:          toString(jobMap["UserID"]),
// 		MaxRetries:      toInt(jobMap["MaxRetries"]),
// 		CurrentRetries:  toInt(jobMap["CurrentRetries"]),
// 		CodeURL:         toString(jobMap["CodeURL"]),
// 	}

// 	// Convert arguments
// 	if args, ok := jobMap["Arguments"].(map[string]interface{}); ok {
// 		job.Arguments = args
// 	}

// 	return job, nil
// }

// // Helper functions for type conversion
// func toString(v interface{}) string {
// 	if s, ok := v.(string); ok {
// 		return s
// 	}
// 	return ""
// }

// func toInt(v interface{}) int {
// 	switch val := v.(type) {
// 	case int:
// 		return val
// 	case float64:
// 		return int(val)
// 	case string:
// 		intVal, _ := strconv.Atoi(val)
// 		return intVal
// 	default:
// 		return 0
// 	}
// }

// func toUint(v interface{}) uint {
// 	switch val := v.(type) {
// 	case uint:
// 		return val
// 	case int:
// 		return uint(val)
// 	case float64:
// 		return uint(val)
// 	default:
// 		return 0
// 	}
// }
