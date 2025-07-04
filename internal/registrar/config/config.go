package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Contract Addresses to listen for events
	avsGovernanceAddress     string
	attestationCenterAddress string

	// RPC URLs for Ethereum and Base
	ethRPCURL       string
	baseRPCURL      string
	pollingInterval time.Duration

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	lastRewardsUpdate string

	// Pinata JWT and Host
	pinataJWT  string
	pinataHost string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                  env.GetEnvBool("DEV_MODE", false),
		avsGovernanceAddress:     env.GetEnv("AVS_GOVERNANCE_ADDRESS", "0x0C77B6273F4852200b17193837960b2f253518FC"),
		attestationCenterAddress: env.GetEnv("ATTESTATION_CENTER_ADDRESS", "0x710DAb96f318b16F0fC9962D3466C00275414Ff0"),
		ethRPCURL:                env.GetEnv("L1_RPC", ""),
		baseRPCURL:               env.GetEnv("L2_RPC", ""),
		pollingInterval:          env.GetEnvDuration("REGISTRAR_POLLING_INTERVAL", 15*time.Minute),
		databaseHostAddress:      env.GetEnv("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:         env.GetEnv("DATABASE_HOST_PORT", "9042"),
		lastRewardsUpdate:        env.GetEnv("LAST_REWARDS_UPDATE", ""),
		pinataJWT:                env.GetEnv("PINATA_JWT", ""),
		pinataHost:               env.GetEnv("PINATA_HOST", ""),
	}
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func validateConfig() error {
	if env.IsEmpty(cfg.ethRPCURL) {
		return fmt.Errorf("empty Ethereum RPC URL")
	}
	if env.IsEmpty(cfg.baseRPCURL) {
		return fmt.Errorf("empty Base RPC URL")
	}
	if !env.IsValidEthAddress(cfg.avsGovernanceAddress){
		return fmt.Errorf("invalid AVS Governance Address: %s", cfg.avsGovernanceAddress)
	}
	if !env.IsValidEthAddress(cfg.attestationCenterAddress) {
		return fmt.Errorf("invalid Attestation Address: %s", cfg.attestationCenterAddress)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if env.IsEmpty(cfg.lastRewardsUpdate) {
		cfg.lastRewardsUpdate = time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	}
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("empty Pinata JWT field")
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("empty Pinata Host field")
	}
	return nil
}

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetEthRPCURL() string {
	return cfg.ethRPCURL
}

func GetBaseRPCURL() string {
	return cfg.baseRPCURL
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetLastRewardsUpdate() string {
	return cfg.lastRewardsUpdate
}

func GetPollingInterval() time.Duration {
	return cfg.pollingInterval
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetPinataHost() string {
	return cfg.pinataHost
}
