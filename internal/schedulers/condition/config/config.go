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

	// Scheduler Private Key and Address
	schedulerPrivateKey string
	schedulerAddress    string

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
		devMode:             env.GetEnvBool("DEV_MODE", false),
		databaseHost:        env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort:    env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort:    env.GetEnv("CONDITION_SCHEDULER_RPC_PORT", "9006"),
		dbServerURL:         env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		schedulerPrivateKey: env.GetEnv("SCHEDULER_PRIVATE_KEY", ""),
		schedulerAddress:    env.GetEnv("SCHEDULER_ADDRESS", ""),
		maxWorkers:          env.GetEnvInt("MAX_WORKERS", 100),
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
	if !env.IsValidPrivateKey(cfg.schedulerPrivateKey) {
		return fmt.Errorf("invalid scheduler private key: %s", cfg.schedulerPrivateKey)
	}
	if !env.IsValidEthAddress(cfg.schedulerAddress) {
		return fmt.Errorf("invalid scheduler address: %s", cfg.schedulerAddress)
	}
	return nil
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return cfg.devMode
}

// GetDatabaseHost returns the database host
func GetDatabaseHost() string {
	return cfg.databaseHost
}

// GetDatabaseHostPort returns the database host port
func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
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
