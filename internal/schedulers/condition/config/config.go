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
	schedulerRPCPort string

	// Scheduler Private Key and Address
	schedulerSigningKey     string
	schedulerSigningAddress string

	// Maximum number of workers
	maxWorkers int

	// Database RPC URL
	dbServerURL string
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                 env.GetEnvBool("DEV_MODE", false),
		schedulerRPCPort:        env.GetEnv("CONDITION_SCHEDULER_RPC_PORT", "9006"),
		dbServerURL:             env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		schedulerSigningKey:     env.GetEnv("CONDITION_SCHEDULER_SIGNING_KEY", ""),
		schedulerSigningAddress: env.GetEnv("CONDITION_SCHEDULER_ADDRESS", ""),
		maxWorkers:              env.GetEnvInt("CONDITION_SCHEDULER_MAX_WORKERS", 100),
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
	if !env.IsValidPrivateKey(cfg.schedulerSigningKey) {
		return fmt.Errorf("invalid scheduler private key: %s", cfg.schedulerSigningKey)
	}
	if !env.IsValidEthAddress(cfg.schedulerSigningAddress) {
		return fmt.Errorf("invalid scheduler address: %s", cfg.schedulerSigningAddress)
	}
	return nil
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return cfg.devMode
}

// GetSchedulerRPCPort returns the scheduler RPC port
func GetSchedulerRPCPort() string {
	return cfg.schedulerRPCPort
}

// GetDBServerURL returns the database server URL
func GetDBServerURL() string {
	return cfg.dbServerURL
}

// GetMaxWorkers returns the maximum number of concurrent workers allowed
func GetMaxWorkers() int {
	return cfg.maxWorkers
}

func GetSchedulerSigningKey() string {
	return cfg.schedulerSigningKey
}

func GetSchedulerSigningAddress() string {
	return cfg.schedulerSigningAddress
}
