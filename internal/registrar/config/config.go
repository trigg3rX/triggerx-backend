package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/env"
)

type Config struct {
	devMode bool

	// Registrar Port
	registrarPort string

	// Contract Addresses to listen for events
	avsGovernanceAddress      string
	avsGovernanceLogicAddress string
	attestationCenterAddress  string
	oblsAddress               string
	triggerGasRegistryAddress string

	// RPC URLs for Ethereum and Base
	rpcProvider     string
	rpcAPIKey       string
	pollingInterval time.Duration

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// Upstash Redis URL and Rest Token
	upstashRedisUrl       string
	upstashRedisRestToken string

	// Sync Configs Update
	lastRewardsUpdate   string
	lastPolledEthBlock  uint64
	lastPolledBaseBlock uint64
	lastPolledOptBlock  uint64

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
		devMode:                   env.GetEnvBool("DEV_MODE", false),
		registrarPort:             env.GetEnvString("REGISTRAR_PORT", "9007"),
		avsGovernanceAddress:      env.GetEnvString("AVS_GOVERNANCE_ADDRESS", "0x12f45551f11Df20b3EcBDf329138Bdc65cc58Ec0"),
		avsGovernanceLogicAddress: env.GetEnvString("AVS_GOVERNANCE_LOGIC_ADDRESS", "0xACB667202C6F9b84D91dA1D66c82f30c66738299"),
		attestationCenterAddress:  env.GetEnvString("ATTESTATION_CENTER_ADDRESS", "0x9725fB95B5ec36c062A49ca2712b3B1ff66F04eD"),
		oblsAddress:               env.GetEnvString("OBLS_ADDRESS", "0x68853222A6Fc1DAE25Dd58FB184dc4470C98F73C"),
		triggerGasRegistryAddress: env.GetEnvString("TRIGGER_GAS_REGISTRY_ADDRESS", "0x85ea3eB894105bD7e7e2A8D34cf66C8E8163CD2a"),
		rpcProvider:               env.GetEnvString("RPC_PROVIDER", ""),
		rpcAPIKey:                 env.GetEnvString("RPC_API_KEY", ""),
		pollingInterval:           env.GetEnvDuration("REGISTRAR_POLLING_INTERVAL", 5*time.Minute),
		databaseHostAddress:       env.GetEnvString("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:          env.GetEnvString("DATABASE_HOST_PORT", "9042"),
		upstashRedisUrl:           env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken:     env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		pinataJWT:                 env.GetEnvString("PINATA_JWT", ""),
		pinataHost:                env.GetEnvString("PINATA_HOST", ""),
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
	if !env.IsValidPort(cfg.registrarPort) {
		return fmt.Errorf("invalid registrar port: %s", cfg.registrarPort)
	}
	if env.IsEmpty(cfg.rpcProvider) {
		return fmt.Errorf("empty RPC Provider")
	}
	if env.IsEmpty(cfg.rpcAPIKey) {
		return fmt.Errorf("empty RPC API Key")
	}
	if !env.IsValidEthAddress(cfg.avsGovernanceAddress) {
		return fmt.Errorf("invalid AVS Governance Address: %s", cfg.avsGovernanceAddress)
	}
	if !env.IsValidEthAddress(cfg.avsGovernanceLogicAddress) {
		return fmt.Errorf("invalid AVS Governance Logic Address: %s", cfg.avsGovernanceLogicAddress)
	}
	if !env.IsValidEthAddress(cfg.attestationCenterAddress) {
		return fmt.Errorf("invalid Attestation Address: %s", cfg.attestationCenterAddress)
	}
	if !env.IsValidEthAddress(cfg.oblsAddress) {
		return fmt.Errorf("invalid OBLS Address: %s", cfg.oblsAddress)
	}
	if !env.IsValidEthAddress(cfg.triggerGasRegistryAddress) {
		return fmt.Errorf("invalid Trigger Gas Registry Address: %s", cfg.triggerGasRegistryAddress)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if env.IsEmpty(cfg.upstashRedisUrl) {
		return fmt.Errorf("empty Upstash Redis URL field")
	}
	if env.IsEmpty(cfg.upstashRedisRestToken) {
		return fmt.Errorf("empty Upstash Redis Rest Token field")
	}
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("empty Pinata JWT field")
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("empty Pinata Host field")
	}
	return nil
}

// Setters
func SetLastRewardsUpdate(timestamp string) {
	cfg.lastRewardsUpdate = timestamp
}

func SetLastPolledEthBlock(blockNumber uint64) {
	cfg.lastPolledEthBlock = blockNumber
}

func SetLastPolledBaseBlock(blockNumber uint64) {
	cfg.lastPolledBaseBlock = blockNumber
}

func SetLastPolledOptBlock(blockNumber uint64) {
	cfg.lastPolledOptBlock = blockNumber
}

// Getters
func IsDevMode() bool {
	return cfg.devMode
}

func GetRegistrarPort() string {
	return cfg.registrarPort
}

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetAvsGovernanceLogicAddress() string {
	return cfg.avsGovernanceLogicAddress
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetOBLSAddress() string {
	return cfg.oblsAddress
}

func GetTriggerGasRegistryAddress() string {
	return cfg.triggerGasRegistryAddress
}

func GetPollingInterval() time.Duration {
	return cfg.pollingInterval
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetUpstashRedisUrl() string {
	return cfg.upstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.upstashRedisRestToken
}

func GetLastRewardsUpdate() string {
	return cfg.lastRewardsUpdate
}

func GetLastPolledEthBlock() uint64 {
	return cfg.lastPolledEthBlock
}

func GetLastPolledBaseBlock() uint64 {
	return cfg.lastPolledBaseBlock
}

func GetLastPolledOptBlock() uint64 {
	return cfg.lastPolledOptBlock
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetPinataHost() string {
	return cfg.pinataHost
}

// Get Chain Configs
func GetChainRPCUrl(isWebSocket bool, chainID string) string {
	var protocol string
	if isWebSocket {
		protocol = "wss://"
	} else {
		protocol = "https://"
	}
	var domain string
	if cfg.rpcProvider == "alchemy" {
		switch chainID {
		case "17000":
			domain = "eth-holesky.g.alchemy.com/v2/"
		case "11155111":
			domain = "eth-sepolia.g.alchemy.com/v2/"
		case "11155420":
			domain = "opt-sepolia.g.alchemy.com/v2/"
		case "84532":
			domain = "base-sepolia.g.alchemy.com/v2/"
		default:
			return ""
		}
	}
	if cfg.rpcProvider == "blast" {
		switch chainID {
		case "17000":
			domain = "eth-holesky.blastapi.io/"
		case "11155111":
			domain = "eth-sepolia.blastapi.io/"
		case "11155420":
			domain = "optimism-sepolia.blastapi.io/"
		case "84532":
			domain = "base-sepolia.blastapi.io/"
		default:
			return ""
		}
	}
	return fmt.Sprintf("%s%s%s", protocol, domain, cfg.rpcAPIKey)
}
