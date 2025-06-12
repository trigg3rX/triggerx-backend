package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Scheduler RPC Port
	eventSchedulerRPCPort string

	// Database RPC URL
	dbServerURL string

	// Maximum number of workers
	maxWorkers int

	// API Keys for Alchemy
	alchemyAPIKey string

	// Scheduler Private Key and Address
	eventSchedulerSigningKey     string
	eventSchedulerSigningAddress string

	// Aggregator RPC URL for forwarding tasks to keeper
	aggregatorRPCURL string
}

var cfg Config

// Helper to detect test environment
func isTestEnv() bool {
	return env.GetEnv("APP_ENV", "") == "test"
}

// Init initializes the configuration for production
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                 env.GetEnvBool("DEV_MODE", false),
		eventSchedulerRPCPort:   env.GetEnv("EVENT_SCHEDULER_RPC_PORT", "9005"),
		dbServerURL:             env.GetEnv("DBSERVER_RPC_URL", "http://localhost:9002"),
		maxWorkers:              env.GetEnvInt("EVENT_SCHEDULER_MAX_WORKERS", 100),
		alchemyAPIKey:           env.GetEnv("ALCHEMY_API_KEY", ""),
		eventSchedulerSigningKey:     env.GetEnv("EVENT_SCHEDULER_SIGNING_KEY", ""),
		eventSchedulerSigningAddress: env.GetEnv("EVENT_SCHEDULER_ADDRESS", ""),
		aggregatorRPCURL:        env.GetEnv("AGGREGATOR_RPC_URL", "http://localhost:9001"),
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
	if !env.IsValidPort(cfg.eventSchedulerRPCPort) {
		return fmt.Errorf("invalid event scheduler RPC port: %s", cfg.eventSchedulerRPCPort)
	}
	if !env.IsValidURL(cfg.dbServerURL) {
		return fmt.Errorf("invalid database server URL: %s", cfg.dbServerURL)
	}
	if !env.IsValidURL(cfg.aggregatorRPCURL) {
		return fmt.Errorf("invalid aggregator RPC URL: %s", cfg.aggregatorRPCURL)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidEthKeyPair(cfg.eventSchedulerSigningKey, cfg.eventSchedulerSigningAddress) {
		return fmt.Errorf("invalid event scheduler signing key pair address: %s", cfg.eventSchedulerSigningAddress)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetSchedulerRPCPort() string {
	return cfg.eventSchedulerRPCPort
}

func GetDBServerURL() string {
	return cfg.dbServerURL
}

// GetChainRPCUrlsTest returns local/test chain RPC URLs
func GetChainRPCUrlsTest() map[string]string {
	local := "http://127.0.0.1:8545"
	return map[string]string{
		"11155420": local,
		"84532":    local,
		"11155111": local,
	}
}

// GetChainRPCUrls returns chain RPC URLs for production or test
func GetChainRPCUrls() map[string]string {
	if isTestEnv() {
		return GetChainRPCUrlsTest()
	}
	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),  // OP Sepolia
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey), // Base Sepolia
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),  // Ethereum Sepolia
	}
}

func GetMaxWorkers() int {
	return cfg.maxWorkers
}

func GetSchedulerSigningKey() string {
	return cfg.eventSchedulerSigningKey
}

func GetSchedulerSigningAddress() string {
	return cfg.eventSchedulerSigningAddress
}

func GetAggregatorRPCURL() string {
	return cfg.aggregatorRPCURL
}

// SetMaxWorkersForTest sets maxWorkers for testing purposes only.
func SetMaxWorkersForTest(n int) {
	cfg.maxWorkers = n
}
