package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// ScyllaDB Host and Port
	databaseHost     string
	databaseHostPort string

	// Scheduler RPC Port
	schedulerRPCPort string

	// Database RPC URL
	dbServerURL string

	// Maximum number of workers
	maxWorkers int

	// API Keys for Alchemy
	alchemyAPIKey string

	// Scheduler Private Key and Address
	schedulerPrivateKey string
	schedulerAddress    string

	// Aggregator RPC URL for forwarding tasks to keeper
	aggregatorRPCURL string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:             env.GetEnvBool("DEV_MODE", false),
		databaseHost:        env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort:    env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort:    env.GetEnv("SCHEDULER_RPC_PORT", "9005"),
		dbServerURL:         env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		maxWorkers:          env.GetEnvInt("MAX_WORKERS", 100),
		alchemyAPIKey:       env.GetEnv("ALCHEMY_API_KEY", ""),
		schedulerPrivateKey: env.GetEnv("SCHEDULER_PRIVATE_KEY", ""),
		schedulerAddress:    env.GetEnv("SCHEDULER_ADDRESS", ""),
		aggregatorRPCURL:    env.GetEnv("AGGREGATOR_RPC_URL", "http://localhost:9001"),
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
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidPrivateKey(cfg.schedulerPrivateKey) {
		return fmt.Errorf("invalid scheduler private key: %s", cfg.schedulerPrivateKey)
	}
	if !env.IsValidEthAddress(cfg.schedulerAddress) {
		return fmt.Errorf("invalid scheduler address: %s", cfg.schedulerAddress)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetDatabaseHost() string {
	return cfg.databaseHost
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetSchedulerRPCPort() string {
	return cfg.schedulerRPCPort
}

func GetDBServerURL() string {
	return cfg.dbServerURL
}

func GetChainRPCUrls() map[string]string {
	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
	}
}

func GetMaxWorkers() int {
	return cfg.maxWorkers
}

func GetSchedulerPrivateKey() string {
	return cfg.schedulerPrivateKey
}

func GetSchedulerAddress() string {
	return cfg.schedulerAddress
}

func GetAggregatorRPCURL() string {
	return cfg.aggregatorRPCURL
}
