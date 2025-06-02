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

	// ScyllaDB Host and Port for Job Polling
	databaseHostAddress string
	databaseHostPort    string

	// Scheduler RPC Port
	schedulerRPCPort string

	// Database RPC URL
	dbServerURL string

	// Scheduler Private Key and Address
	schedulerPrivateKey string
	schedulerAddress    string

	// Maximum number of workers
	maxWorkers int

	// Time Durations
	pollingInterval  time.Duration
	pollingLookAhead time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:             env.GetEnvBool("DEV_MODE", false),
		databaseHostAddress: env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort:    env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort:    env.GetEnv("SCHEDULER_RPC_PORT", "9004"),
		dbServerURL:         env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		schedulerPrivateKey: env.GetEnv("SCHEDULER_PRIVATE_KEY", ""),
		schedulerAddress:    env.GetEnv("SCHEDULER_ADDRESS", ""),
		maxWorkers:          env.GetEnvInt("MAX_WORKERS", 10),
		pollingInterval:     env.GetEnvDuration("POLLING_INTERVAL", 50*time.Second),
		pollingLookAhead:    env.GetEnvDuration("POLLING_LOOK_AHEAD", 60*time.Second),
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

func IsDevMode() bool {
	return cfg.devMode
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
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

func GetPollingInterval() time.Duration {
	return cfg.pollingInterval
}

func GetPollingLookAhead() time.Duration {
	return cfg.pollingLookAhead
}

func GetMaxWorkers() int {
	return cfg.maxWorkers
}
