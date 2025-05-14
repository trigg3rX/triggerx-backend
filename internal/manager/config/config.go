package config

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"log"

	"github.com/trigg3rX/triggerx-backend/pkg/utils"
)

var (
	FoundNextPerformer bool

	EtherscanApiKey string
	AlchemyApiKey   string
	IpfsHost        string

	DeployerPrivateKey string
	P2PPrivateKey      string

	ManagerRPCPort       string
	DatabaseRPCAddress   string
	AggregatorRPCAddress string

	DevMode bool
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DevMode = os.Getenv("DEV_MODE") == "true"

	FoundNextPerformer = false
	EtherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	if utils.IsEmpty(EtherscanApiKey) {
		log.Fatal("Invalid Etherscan API Key")
	}

	AlchemyApiKey = os.Getenv("ALCHEMY_API_KEY")
	if utils.IsEmpty(AlchemyApiKey) {
		log.Fatal("Invalid Alchemy API Key")
	}

	DatabaseRPCAddress = os.Getenv("DATABASE_RPC_ADDRESS")
	if !utils.IsValidRPCAddress(DatabaseRPCAddress) {
		log.Fatal("Invalid Database RPC Address")
	}
	ManagerRPCPort = os.Getenv("MANAGER_RPC_PORT")
	if !utils.IsValidPort(ManagerRPCPort) {
		log.Fatal("Invalid Manager RPC Port")
	}

	AggregatorRPCAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	if !utils.IsValidRPCAddress(AggregatorRPCAddress) {
		log.Fatal("Invalid Aggregator RPC Address")
	}
	
	IpfsHost = os.Getenv("IPFS_HOST")
	if utils.IsEmpty(IpfsHost) {
		log.Fatal("Invalid IPFS Host")
	}

	gin.SetMode(gin.ReleaseMode)
}
