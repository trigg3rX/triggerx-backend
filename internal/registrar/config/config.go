package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

type Config struct {
	AvsGovernanceAddress     string
	AttestationCenterAddress string
	EthRPCURL                string
	BaseRPCURL               string
	PollingInterval          time.Duration
	IPFSHost                 string
	DatabaseHost             string
	DatabaseHostPort         string
	LastRewardsUpdate        string
	DevMode                  bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode:                  os.Getenv("DEV_MODE") == "true",
		AvsGovernanceAddress:     os.Getenv("AVS_GOVERNANCE_ADDRESS"),
		AttestationCenterAddress: os.Getenv("ATTESTATION_CENTER_ADDRESS"),
		EthRPCURL:                os.Getenv("L1_RPC"),
		BaseRPCURL:               os.Getenv("L2_RPC"),
		PollingInterval:          setPollingInterval(),
		IPFSHost:                 os.Getenv("IPFS_HOST"),
		DatabaseHost:             os.Getenv("DATABASE_HOST"),
		DatabaseHostPort:         os.Getenv("DATABASE_HOST_PORT"),
		LastRewardsUpdate:        os.Getenv("LAST_REWARDS_UPDATE"),
	}

	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func validateConfig() error {
	if !validator.IsValidAddress(cfg.AvsGovernanceAddress) {
		log.Fatal("Invalid AvsGovernanceAddress")
	}

	if !validator.IsValidAddress(cfg.AttestationCenterAddress) {
		log.Fatal("Invalid AttestationCenterAddress")
	}

	if validator.IsEmpty(cfg.EthRPCURL) {
		log.Fatal("Invalid EthRpcUrl")
	}

	if validator.IsEmpty(cfg.BaseRPCURL) {
		log.Fatal("Invalid BaseRpcUrl")
	}

	pollingIntervalStr := os.Getenv("POLLING_INTERVAL")
	if validator.IsEmpty(pollingIntervalStr) {
		log.Fatal("Invalid PollingInterval")
	}
	var parseErr error
	cfg.PollingInterval, parseErr = time.ParseDuration(pollingIntervalStr)
	if parseErr != nil {
		log.Fatal("Invalid PollingInterval format: ", parseErr)
	}

	if !validator.IsValidIPAddress(cfg.DatabaseHost) {
		log.Fatal("Invalid DatabaseHost")
	}

	if !validator.IsValidPort(cfg.DatabaseHostPort) {
		log.Fatal("Invalid DatabaseHostPort")
	}

	if validator.IsEmpty(cfg.IPFSHost) {
		log.Fatal("Invalid IpfsHost")
	}

	if validator.IsEmpty(cfg.LastRewardsUpdate) {
		cfg.LastRewardsUpdate = time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	}

	return nil
}

func setPollingInterval() time.Duration {
	pollingIntervalStr := os.Getenv("POLLING_INTERVAL")
	if validator.IsEmpty(pollingIntervalStr) {
		log.Fatal("Invalid PollingInterval")
	}
	var parseErr error
	pollingInterval, parseErr := time.ParseDuration(pollingIntervalStr)
	if parseErr != nil {
		log.Fatal("Invalid PollingInterval format: ", parseErr)
	}
	return pollingInterval
}

func GetAvsGovernanceAddress() string {
	return cfg.AvsGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.AttestationCenterAddress
}

func GetEthRPCURL() string {
	return cfg.EthRPCURL
}

func GetBaseRPCURL() string {
	return cfg.BaseRPCURL
}

func GetIPFSHost() string {
	return cfg.IPFSHost
}

func GetDatabaseHost() string {
	return cfg.DatabaseHost
}

func GetDatabaseHostPort() string {
	return cfg.DatabaseHostPort
}

func GetLastRewardsUpdate() string {
	return cfg.LastRewardsUpdate
}

func GetPollingInterval() time.Duration {
	return cfg.PollingInterval
}

func IsDevMode() bool {
	return cfg.DevMode
}
