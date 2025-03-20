package registrar

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger logging.Logger

	EthRpcUrl   string
	EthWsRpcUrl string
	BaseRpcUrl  string
	BaseWsRpcUrl string
	AvsGovernanceAddress     string
	AttestationCenterAddress string

	DeployerPrivateKey string

	DatabaseIPAddress string
)

func Init() {
	logger = logging.GetLogger(logging.Development, logging.RegistrarProcess)

	// var err error
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	EthRpcUrl = os.Getenv("L1_RPC")
	EthWsRpcUrl = os.Getenv("L1_WS_RPC")
	BaseRpcUrl = os.Getenv("L2_RPC")
	BaseWsRpcUrl = os.Getenv("L2_WS_RPC")
	DeployerPrivateKey = os.Getenv("PRIVATE_KEY_DEPLOYER")
	AvsGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if EthRpcUrl == "" || EthWsRpcUrl == "" || BaseRpcUrl == "" || BaseWsRpcUrl == "" || DeployerPrivateKey == "" || DatabaseIPAddress == "" || AvsGovernanceAddress == "" || AttestationCenterAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}
