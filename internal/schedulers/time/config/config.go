package config

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode          bool
	databaseHost     string
	databaseHostPort string
	schedulerRPCPort string
	dbServerURL      string
	maxWorkers       int
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	maxWorkersStr := env.GetEnv("MAX_WORKERS", "10")
	maxWorkers, err := strconv.Atoi(maxWorkersStr)
	if err != nil {
		maxWorkers = 10 // fallback to default
	}

	cfg = Config{
		devMode:          env.GetEnvBool("DEV_MODE", false),
		databaseHost:     env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort: env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort: env.GetEnv("SCHEDULER_RPC_PORT", "9004"),
		dbServerURL:      env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		maxWorkers:       maxWorkers,
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

// GetMaxWorkers returns the maximum number of workers
func GetMaxWorkers() int {
	return cfg.maxWorkers
}
