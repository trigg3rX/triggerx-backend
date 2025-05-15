package config

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

// Config holds all configuration for the health service
type Config struct {
	HealthRPCPort      string
	DatabaseRPCAddress string
	DevMode            bool
}

var cfg Config

// Init initializes the configuration for the health service
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode:            os.Getenv("DEV_MODE") == "true",
		HealthRPCPort:      os.Getenv("HEALTH_RPC_PORT"),
		DatabaseRPCAddress: os.Getenv("DATABASE_RPC_ADDRESS"),
	}

	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func validateConfig() error {
	if !validator.IsValidPort(cfg.HealthRPCPort) {
		return fmt.Errorf("invalid Health RPC Port: %s", cfg.HealthRPCPort)
	}

	if !validator.IsValidRPCAddress(cfg.DatabaseRPCAddress) {
		return fmt.Errorf("invalid Database RPC Address: %s", cfg.DatabaseRPCAddress)
	}

	return nil
}

// GetHealthRPCPort returns the configured health RPC port
func GetHealthRPCPort() string {
	return cfg.HealthRPCPort
}

// GetDatabaseRPCAddress returns the configured database RPC address
func GetDatabaseRPCAddress() string {
	return cfg.DatabaseRPCAddress
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return cfg.DevMode
}
