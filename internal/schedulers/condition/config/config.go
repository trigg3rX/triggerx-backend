package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/env"
)

type Config struct {
	devMode bool

	// Scheduler RPC Port
	conditionSchedulerRPCPort string

	// Database RPC URL
	dbServerURL string
	// Aggregator RPC URL
	aggregatorRPCURL string
	// Redis API URL
	redisRPCUrl string

	// Scheduler ID for consumer groups
	conditionSchedulerID int

	// Maximum number of workers
	maxWorkers int

	// API Keys for Alchemy
	alchemyAPIKey string
}

var cfg Config

// Helper to detect test environment
func isTestEnv() bool {
	return env.GetEnv("APP_ENV", "") == "test"
}

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                          env.GetEnvBool("DEV_MODE", false),
		conditionSchedulerRPCPort:        env.GetEnv("CONDITION_SCHEDULER_RPC_PORT", "9006"),
		dbServerURL:                      env.GetEnv("DBSERVER_RPC_URL", "http://localhost:9002"),
		aggregatorRPCURL:                 env.GetEnv("AGGREGATOR_RPC_URL", "http://localhost:9001"),
		redisRPCUrl:                      env.GetEnv("REDIS_RPC_URL", "http://localhost:9003"),
		conditionSchedulerID:             env.GetEnvInt("CONDITION_SCHEDULER_ID", 5678),
		maxWorkers:                       env.GetEnvInt("CONDITION_SCHEDULER_MAX_WORKERS", 100),
		alchemyAPIKey:                    env.GetEnv("ALCHEMY_API_KEY", ""),
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
	if !env.IsValidPort(cfg.conditionSchedulerRPCPort) {
		return fmt.Errorf("invalid condition scheduler RPC port: %s", cfg.conditionSchedulerRPCPort)
	}
	if !env.IsValidURL(cfg.dbServerURL) {
		return fmt.Errorf("invalid database server URL: %s", cfg.dbServerURL)
	}
	if !env.IsValidURL(cfg.aggregatorRPCURL) {
		return fmt.Errorf("invalid aggregator RPC URL: %s", cfg.aggregatorRPCURL)
	}
	if !env.IsValidURL(cfg.redisRPCUrl) {
		return fmt.Errorf("invalid Redis API URL: %s", cfg.redisRPCUrl)
	}
	return nil
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return cfg.devMode
}

// GetSchedulerRPCPort returns the scheduler RPC port
func GetSchedulerRPCPort() string {
	return cfg.conditionSchedulerRPCPort
}

// GetDBServerURL returns the database server URL
func GetDBServerURL() string {
	return cfg.dbServerURL
}

// GetAggregatorRPCURL returns the aggregator RPC URL
func GetAggregatorRPCURL() string {
	return cfg.aggregatorRPCURL
}

// GetRedisRPCUrl returns the Redis API URL
func GetRedisRPCUrl() string {
	return cfg.redisRPCUrl
}

// GetMaxWorkers returns the maximum number of concurrent workers allowed
func GetMaxWorkers() int {
	return cfg.maxWorkers
}

func GetSchedulerID() int {
	return cfg.conditionSchedulerID
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

	if cfg.alchemyAPIKey == "" {
		// Fallback to public endpoints if no Alchemy key
		return map[string]string{
			"11155420": "https://sepolia.optimism.io",
			"84532":    "https://sepolia.base.org",
			"11155111": "https://ethereum-sepolia.publicnode.com",
		}
	}

	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),  // OP Sepolia
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey), // Base Sepolia
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),  // Ethereum Sepolia
	}
}
