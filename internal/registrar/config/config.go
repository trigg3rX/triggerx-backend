package config

import (
	"os"
	"sync"
	"time"

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

	IpfsHost string

	DatabaseDockerIPAddress string
	DatabaseDockerPort      string

	LastRewardsUpdate string

	// Mutex to protect config updates
	configMutex sync.Mutex
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
	DatabaseDockerIPAddress = os.Getenv("DATABASE_DOCKER_IP_ADDRESS")
	DatabaseDockerPort = os.Getenv("DATABASE_DOCKER_PORT")
	LastRewardsUpdate = os.Getenv("LAST_REWARDS_UPDATE")
	IpfsHost = os.Getenv("IPFS_HOST")
	// If LastRewardsUpdate is not set, initialize with yesterday's date
	if LastRewardsUpdate == "" {
		LastRewardsUpdate = time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		logger.Info("LastRewardsUpdate not set, initializing with yesterday's date")
	}

	if EthRpcUrl == "" || BaseRpcUrl == "" || DatabaseDockerIPAddress == "" || DatabaseDockerPort == "" || AvsGovernanceAddress == "" || AttestationCenterAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}

// UpdateLastRewardsTimestamp updates the LastRewardsUpdate timestamp in memory
// It doesn't persist the change to the .env file for simplicity, but logs the update
func UpdateLastRewardsTimestamp(timestamp string) {
	configMutex.Lock()
	defer configMutex.Unlock()

	LastRewardsUpdate = timestamp
	logger.Infof("Updated LastRewardsUpdate to: %s", timestamp)
}
