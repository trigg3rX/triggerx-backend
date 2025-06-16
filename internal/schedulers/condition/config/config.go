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
	conditionSchedulerRPCPort string

	// Scheduler Private Key and Address
	conditionSchedulerSigningKey     string
	conditionSchedulerSigningAddress string

	// Maximum number of workers
	maxWorkers int

	// Database RPC URL
	dbServerURL string

	// Aggregator RPC URL
	aggregatorRPCURL string
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                 env.GetEnvBool("DEV_MODE", false),
		conditionSchedulerRPCPort:        env.GetEnv("CONDITION_SCHEDULER_RPC_PORT", "9006"),
		dbServerURL:             env.GetEnv("DBSERVER_RPC_URL", "http://localhost:9002"),
		conditionSchedulerSigningKey:     env.GetEnv("CONDITION_SCHEDULER_SIGNING_KEY", ""),
		conditionSchedulerSigningAddress: env.GetEnv("CONDITION_SCHEDULER_SIGNING_ADDRESS", ""),
		maxWorkers:              env.GetEnvInt("CONDITION_SCHEDULER_MAX_WORKERS", 100),
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
	if !env.IsValidPort(cfg.conditionSchedulerRPCPort) {
		return fmt.Errorf("invalid condition scheduler RPC port: %s", cfg.conditionSchedulerRPCPort)
	}
	if !env.IsValidURL(cfg.dbServerURL) {
		return fmt.Errorf("invalid database server URL: %s", cfg.dbServerURL)
	}
	if !env.IsValidURL(cfg.aggregatorRPCURL) {
		return fmt.Errorf("invalid aggregator RPC URL: %s", cfg.aggregatorRPCURL)
	}
	if !env.IsValidEthKeyPair(cfg.conditionSchedulerSigningKey, cfg.conditionSchedulerSigningAddress) {
		return fmt.Errorf("invalid condition scheduler signing key pair address: %s", cfg.conditionSchedulerSigningAddress)
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

// GetMaxWorkers returns the maximum number of concurrent workers allowed
func GetMaxWorkers() int {
	return cfg.maxWorkers
}

func GetSchedulerSigningKey() string {
	return cfg.conditionSchedulerSigningKey
}

func GetSchedulerSigningAddress() string {
	return cfg.conditionSchedulerSigningAddress
}
