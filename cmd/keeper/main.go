package main

// import (
//     "context"
//     "log"
//     "os"
//     "os/signal"
//     "syscall"
//     "gopkg.in/yaml.v3"
    
//     "github.com/trigg3rX/triggerx-backend/pkg/keeper"
//     "github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// func main() {
//     log.SetOutput(os.Stdout)
//     log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

//     ctx, cancel := context.WithCancel(context.Background())
//     defer cancel()

//     config := types.NodeConfig{}

//     // Load the YAML config
//     yamlFile, err := os.ReadFile("triggerx_operator.yaml")
//     if err != nil {
//         log.Fatalf("Error reading YAML config file: %v", err)
//     }

//     var yamlConfig struct {
//         Keeper struct {
//             Address          string `yaml:"address"`
//             EcdsaKeystore   string `yaml:"ecdsa_keystore_path"`
//             BlsKeystore     string `yaml:"bls_keystore_path"`
//         } `yaml:"keeper"`
//         Environment struct {
//             EthRpcUrl string `yaml:"ethrpcurl"`
//             EthWsUrl  string `yaml:"ethwsurl"`
//         } `yaml:"environment"`
//         Prometheus struct {
//             PortAddress string `yaml:"port_address"`
//         } `yaml:"prometheus"`
//         Addresses struct {
//             ServiceManagerAddress  string `yaml:"service_manager_address"`
//             OperatorStateRetriever string `yaml:"operator_state_retriever"`
//         } `yaml:"addresses"`
//     }

//     if err := yaml.Unmarshal(yamlFile, &yamlConfig); err != nil {
//         log.Fatalf("Error parsing YAML config: %v", err)
//     }

//     // Set config values from YAML
//     config.EthRpcUrl = yamlConfig.Environment.EthRpcUrl
//     config.EthWsUrl = yamlConfig.Environment.EthWsUrl
//     config.ServiceManagerAddress = yamlConfig.Addresses.ServiceManagerAddress
//     config.OperatorStateRetrieverAddress = yamlConfig.Addresses.OperatorStateRetriever
//     config.EigenMetricsIpPortAddress = yamlConfig.Prometheus.PortAddress
//     config.EcdsaKeystorePath = yamlConfig.Keeper.EcdsaKeystore
//     config.BlsKeystorePath = yamlConfig.Keeper.BlsKeystore

//     keeperNode, err := keeper.NewKeeperFromConfig(config)
//     if err != nil {
//         log.Fatalf("Failed to create keeper node: %v", err)
//     }

//     sigChan := make(chan os.Signal, 1)
//     signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

//     errChan := make(chan error, 1)
//     go func() {
//         if err := keeperNode.Start(ctx); err != nil {
//             errChan <- err
//         }
//     }()

//     select {
//     case <-sigChan:
//         log.Println("Received shutdown signal")
//         cancel()
//     case err := <-errChan:
//         log.Printf("Keeper node error: %v", err)
//         cancel()
//     }

//     log.Println("Shutting down keeper node...")
// }


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
