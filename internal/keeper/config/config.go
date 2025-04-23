package config

import (
	"os"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

	// User Entered Information
	EthRPCUrl                   string
	AlchemyAPIKey               string
	PrivateKeyConsensus          string
	PrivateKeyController          string
	KeeperAddress             	string
	KeeperRPCPort               string
	PublicIPV4Address           string
	PeerID                       string

	// Provided Information
	PinataApiKey        string
	PinataSecretApiKey  string
	IpfsHost            string
	AggregatorIPAddress string
	HealthIPAddress     string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}

	PinataApiKey = os.Getenv("PINATA_API_KEY")
	PinataSecretApiKey = os.Getenv("PINATA_SECRET_API_KEY")
	IpfsHost = os.Getenv("IPFS_HOST")
	AggregatorIPAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	HealthIPAddress = os.Getenv("HEALTH_IP_ADDRESS")

	if PinataApiKey == "" || PinataSecretApiKey == "" || AggregatorIPAddress == "" || HealthIPAddress == "" {
		logger.Fatal(".env FILE NOT PRESENT AT EXPEXTED PATH")
	}

	EthRPCUrl := os.Getenv("L1_RPC")
	AlchemyAPIKey = strings.TrimPrefix(EthRPCUrl, "https://eth-holesky.g.alchemy.com/v2/")

	PrivateKeyController = os.Getenv("PRIVATE_KEY_CONTROLLER")
	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	if PrivateKeyConsensus == "" { logger.Fatal(">>> PRIVATE_KEY not set in ENV !!!")}

	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	if KeeperAddress == "" { logger.Fatal(">>> OPERATOR_ADDRESS not set in ENV !!!")}

	KeeperRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	if KeeperRPCPort == "" { logger.Fatal(">>> OPERATOR_RPC_PORT not set in ENV !!!")}

	PublicIPV4Address = os.Getenv("PUBLIC_IPV4_ADDRESS")
	PeerID = os.Getenv("PEER_ID")

	if  PeerID == "" || PublicIPV4Address == "" {
		logger.Info(">>> PEER_ID or PUBLIC_IPV4_ADDRESS not set properly in .env !!!")
		logger.Info("Please set the variables in the .env file")
		logger.Info("Without them the keeper will not be able to reconnect to the network")
		logger.Fatal("Get Peer ID from https://triggerx.gitbook.io/triggerx-docs/join-as-keeper#p2p-config")
	}

	gin.SetMode(gin.ReleaseMode)
}
