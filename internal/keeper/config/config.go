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
	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	PrivateKeyController = os.Getenv("OPERATOR_PRIVATE_KEY")
	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	KeeperRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	PublicIPV4Address = os.Getenv("PUBLIC_IPV4_ADDRESS")
	PeerID = os.Getenv("PEER_ID")

	if PrivateKeyConsensus == "" || KeeperAddress == "" || KeeperRPCPort == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	if  PeerID == "" || PublicIPV4Address == "" {
		logger.Info("Peer ID or Public IPV4 Address not set properly !!!")
		logger.Info("Please set the variables in the .env file")
		logger.Info("Without them the keeper will not be able to reconnect to the network")
		logger.Fatal("Get Peer ID from https://triggerx.gitbook.io/triggerx-docs/join-as-keeper#p2p-config")
	}

	gin.SetMode(gin.ReleaseMode)
}
