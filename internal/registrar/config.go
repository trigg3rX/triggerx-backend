package registrar

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger logging.Logger

	EthRpcUrl                string
	EthWsRpcUrl              string
	BaseRpcUrl               string
	BaseWsRpcUrl             string
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
	logger.Info(fmt.Sprintf("EthRpcUrl: %s", EthRpcUrl))
	// EthWsRpcUrl = os.Getenv("L1_WS_RPC")
	// logger.Info(fmt.Sprintf("EthWsRpcUrl: %s", EthWsRpcUrl))
	BaseRpcUrl = os.Getenv("L2_RPC")
	logger.Info(fmt.Sprintf("BaseRpcUrl: %s", BaseRpcUrl))
	// BaseWsRpcUrl = os.Getenv("L2_WS_RPC")
	// logger.Info(fmt.Sprintf("BaseWsRpcUrl: %s", BaseWsRpcUrl))
	DeployerPrivateKey = os.Getenv("PRIVATE_KEY_DEPLOYER")
	logger.Info(fmt.Sprintf("DeployerPrivateKey: %s", DeployerPrivateKey))
	AvsGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	logger.Info(fmt.Sprintf("AvsGovernanceAddress: %s", AvsGovernanceAddress))
	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	logger.Info(fmt.Sprintf("AttestationCenterAddress: %s", AttestationCenterAddress))
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")
	logger.Info(fmt.Sprintf("DatabaseIPAddress: %s", DatabaseIPAddress))

	if EthRpcUrl == "" || BaseRpcUrl == "" || DeployerPrivateKey == "" || DatabaseIPAddress == "" || AvsGovernanceAddress == "" || AttestationCenterAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}
