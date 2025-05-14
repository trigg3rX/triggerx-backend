package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/utils"
)

var (
	AvsGovernanceAddress     string
	AttestationCenterAddress string

	EthRpcUrl  string
	BaseRpcUrl string

	PollingInterval time.Duration
	IpfsHost        string

	DatabaseHost     string
	DatabaseHostPort string

	LastRewardsUpdate string

	DevMode     bool
	configMutex sync.Mutex
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DevMode = os.Getenv("DEV_MODE") == "true"

	AvsGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	if !utils.IsValidAddress(AvsGovernanceAddress) {
		log.Fatal("Invalid AvsGovernanceAddress")
	}

	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	if !utils.IsValidAddress(AttestationCenterAddress) {
		log.Fatal("Invalid AttestationCenterAddress")
	}

	EthRpcUrl = os.Getenv("L1_RPC")
	if utils.IsEmpty(EthRpcUrl) {
		log.Fatal("Invalid EthRpcUrl")
	}

	BaseRpcUrl = os.Getenv("L2_RPC")
	if utils.IsEmpty(BaseRpcUrl) {
		log.Fatal("Invalid BaseRpcUrl")
	}

	pollingIntervalStr := os.Getenv("POLLING_INTERVAL")
	if utils.IsEmpty(pollingIntervalStr) {
		log.Fatal("Invalid PollingInterval")
	}
	var parseErr error
	PollingInterval, parseErr = time.ParseDuration(pollingIntervalStr)
	if parseErr != nil {
		log.Fatal("Invalid PollingInterval format: ", parseErr)
	}

	DatabaseHost = os.Getenv("DATABASE_HOST")
	if !utils.IsValidIPAddress(DatabaseHost) {
		log.Fatal("Invalid DatabaseHost")
	}

	DatabaseHostPort = os.Getenv("DATABASE_HOST_PORT")
	if !utils.IsValidPort(DatabaseHostPort) {
		log.Fatal("Invalid DatabaseHostPort")
	}

	IpfsHost = os.Getenv("IPFS_HOST")
	if utils.IsEmpty(IpfsHost) {
		log.Fatal("Invalid IpfsHost")
	}

	LastRewardsUpdate = os.Getenv("LAST_REWARDS_UPDATE")
	if utils.IsEmpty(LastRewardsUpdate) {
		LastRewardsUpdate = time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	}

	gin.SetMode(gin.ReleaseMode)
}

func UpdateLastRewardsTimestamp(timestamp string) {
	configMutex.Lock()
	defer configMutex.Unlock()

	LastRewardsUpdate = timestamp
	log.Printf("Updated LastRewardsUpdate to: %s", timestamp)
}
