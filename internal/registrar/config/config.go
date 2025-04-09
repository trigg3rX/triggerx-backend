package config

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger logging.Logger

	EthRpcUrl                string
	BaseRpcUrl               string
	AvsGovernanceAddress     string
	AttestationCenterAddress string

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
	BaseRpcUrl = os.Getenv("L2_RPC")
	AvsGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if EthRpcUrl == "" || BaseRpcUrl == "" || DatabaseIPAddress == "" || AvsGovernanceAddress == "" || AttestationCenterAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}