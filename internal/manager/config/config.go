package config

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger             = logging.GetLogger(logging.Development, logging.ManagerProcess)
	FoundNextPerformer bool

	EtherscanApiKey string
	AlchemyApiKey   string
	IpfsHost        string

	DeployerPrivateKey string
	P2PPrivateKey      string

	ManagerRPCPort       string
	DatabaseIPAddress    string
	AggregatorRPCAddress string
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
	ManagerRPCPort = os.Getenv("MANAGER_RPC_PORT")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	AggregatorRPCAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	IpfsHost = os.Getenv("IPFS_HOST")

	if EtherscanApiKey == "" || AlchemyApiKey == "" || DeployerPrivateKey == "" || ManagerRPCPort == "" || DatabaseIPAddress == "" || AggregatorRPCAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}
