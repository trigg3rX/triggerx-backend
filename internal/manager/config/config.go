package config

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)

	EtherscanApiKey string
	AlchemyApiKey   string

	DeployerPrivateKey string

	ManagerRPCPort    string
	DatabaseIPAddress string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	EtherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	AlchemyApiKey = os.Getenv("ALCHEMY_API_KEY")
	DeployerPrivateKey = os.Getenv("PRIVATE_KEY_DEPLOYER")
	ManagerRPCPort = os.Getenv("MANAGER_RPC_PORT")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if EtherscanApiKey == "" || AlchemyApiKey == "" || DeployerPrivateKey == "" || ManagerRPCPort == "" || DatabaseIPAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}