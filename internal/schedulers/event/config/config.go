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
	// Chain RPC URLs
	alchemyAPIKey string
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	maxWorkersStr := env.GetEnv("MAX_WORKERS", "100")
	maxWorkers, err := strconv.Atoi(maxWorkersStr)
	if err != nil {
		maxWorkers = 100 // fallback to default
	}

	cfg = Config{
		devMode:          env.GetEnv("DEV_MODE", "false") == "true",
		databaseHost:     env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort: env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort: env.GetEnv("SCHEDULER_RPC_PORT", "9005"),
		dbServerURL:      env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		maxWorkers:       maxWorkers,
		// Chain RPC URLs with default values
		alchemyAPIKey: env.GetEnv("ALCHEMY_API_KEY", ""),
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

// GetChainRPCUrls returns a map of chain IDs to RPC URLs
func GetChainRPCUrls() map[string]string {
	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),   // OP Sepolia
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey), // Base Sepolia
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),  // Ethereum Sepolia
	}
}

// GetMaxWorkers returns the maximum number of concurrent workers allowed
func GetMaxWorkers() int {
	return cfg.maxWorkers
}
