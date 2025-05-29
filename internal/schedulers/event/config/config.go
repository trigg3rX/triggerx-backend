package config

import (
	"fmt"

	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode          bool
	databaseHost     string
	databaseHostPort string
	schedulerRPCPort string
	dbServerURL      string
	// Chain RPC URLs
	opSepoliaRPC   string
	baseSepoliaRPC string
	ethSepoliaRPC  string
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		devMode:          env.GetEnv("DEV_MODE", "false") == "true",
		databaseHost:     env.GetEnv("DATABASE_HOST", "localhost"),
		databaseHostPort: env.GetEnv("DATABASE_HOST_PORT", "9042"),
		schedulerRPCPort: env.GetEnv("SCHEDULER_RPC_PORT", "9004"),
		dbServerURL:      env.GetEnv("DATABASE_RPC_URL", "http://localhost:9002"),
		// Chain RPC URLs with default values
		opSepoliaRPC:   env.GetEnv("OP_SEPOLIA_RPC_URL", "https://sepolia.optimism.io"),
		baseSepoliaRPC: env.GetEnv("BASE_SEPOLIA_RPC_URL", "https://sepolia.base.org"),
		ethSepoliaRPC:  env.GetEnv("ETH_SEPOLIA_RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com"),
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
		"11155420": cfg.opSepoliaRPC,   // OP Sepolia
		"84532":    cfg.baseSepoliaRPC, // Base Sepolia
		"11155111": cfg.ethSepoliaRPC,  // Ethereum Sepolia
	}
}

// GetOPSepoliaRPC returns the OP Sepolia RPC URL
func GetOPSepoliaRPC() string {
	return cfg.opSepoliaRPC
}

// GetBaseSepoliaRPC returns the Base Sepolia RPC URL
func GetBaseSepoliaRPC() string {
	return cfg.baseSepoliaRPC
}

// GetEthSepoliaRPC returns the Ethereum Sepolia RPC URL
func GetEthSepoliaRPC() string {
	return cfg.ethSepoliaRPC
}
