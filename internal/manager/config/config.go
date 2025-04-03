package config

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	FoundNextPerformer bool

	EtherscanApiKey string
	AlchemyApiKey   string

	DeployerPrivateKey string
	P2PPrivateKey string

	ManagerRPCPort    string
	AggregatorRPCAddress string
	DatabaseIPAddress string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	FoundNextPerformer = false
	EtherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	AlchemyApiKey = os.Getenv("ALCHEMY_API_KEY")
	DeployerPrivateKey = os.Getenv("PRIVATE_KEY_DEPLOYER")
	P2PPrivateKey = os.Getenv("PRIVATE_KEY_MANAGER_P2P")
	ManagerRPCPort = os.Getenv("MANAGER_RPC_PORT")
	AggregatorRPCAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if EtherscanApiKey == "" || AlchemyApiKey == "" || DeployerPrivateKey == "" || ManagerRPCPort == "" || DatabaseIPAddress == "" || AggregatorRPCAddress == "" || P2PPrivateKey == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}