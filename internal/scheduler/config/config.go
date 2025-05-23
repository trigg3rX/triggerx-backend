package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	devMode          bool
	databaseHost     string
	databaseHostPort string
	schedulerRPCPort string
)

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	devMode = getEnvBool("DEV_MODE", false)
	databaseHost = getEnv("DATABASE_HOST", "localhost")
	databaseHostPort = getEnv("DATABASE_HOST_PORT", "9042")
	schedulerRPCPort = getEnv("SCHEDULER_RPC_PORT", "9005")

	return nil
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return devMode
}

// GetDatabaseHost returns the database host
func GetDatabaseHost() string {
	return databaseHost
}

// GetDatabaseHostPort returns the database host port
func GetDatabaseHostPort() string {
	return databaseHostPort
}

// GetSchedulerRPCPort returns the scheduler RPC port
func GetSchedulerRPCPort() string {
	return schedulerRPCPort
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolValue
	}
	return defaultValue
}
