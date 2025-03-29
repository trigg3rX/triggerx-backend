package config

import (
	"fmt"
	"net"
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
	KeeperP2PPort       		string
	ConnectionAddress           string

	// Provided Information
	PinataApiKey        string
	PinataSecretApiKey  string
	IpfsHost            string
	AggregatorIPAddress string
	ManagerIPAddress    string
	DatabaseIPAddress   string
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
	ManagerIPAddress = os.Getenv("MANAGER_IP_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if PinataApiKey == "" || PinataSecretApiKey == "" || AggregatorIPAddress == "" || ManagerIPAddress == "" || DatabaseIPAddress == "" {
		logger.Fatal(".env FILE NOT PRESENT AT EXPEXTED PATH")
	}

	EthRPCUrl := os.Getenv("L1_RPC")
	AlchemyAPIKey = strings.TrimPrefix(EthRPCUrl, "https://eth-holesky.g.alchemy.com/v2/")
	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	PrivateKeyController = os.Getenv("OPERATOR_PRIVATE_KEY")
	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	KeeperRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	KeeperP2PPort = os.Getenv("OPERATOR_P2P_PORT")

	if PrivateKeyConsensus == "" || PrivateKeyController == "" || KeeperAddress == "" || KeeperRPCPort == "" || KeeperP2PPort == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	// Get the connection address (IP address of the machine running this code)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Error("Failed to determine local IP address", "error", err)
		// Fallback to localhost if we can't determine the IP
		ConnectionAddress = "127.0.0.1"
	} else {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		ConnectionAddress = fmt.Sprintf("http://%s:%s", localAddr.IP.String(), KeeperRPCPort)
	}
	
	// logger.Info("Keeper connection address set", "address", ConnectionAddress)

	gin.SetMode(gin.ReleaseMode)
}
