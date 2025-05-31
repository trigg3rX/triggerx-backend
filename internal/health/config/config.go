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

	BotToken      string
	EmailUser     string
	EmailPassword string

	DatabaseHost     string
	DatabaseHostPort string

	DevMode bool
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
		BotToken:           os.Getenv("BOT_TOKEN"),
		EmailUser:          os.Getenv("EMAIL_USER"),
		EmailPassword:      os.Getenv("EMAIL_PASSWORD"),
		DatabaseHost:       os.Getenv("DATABASE_HOST"),
		DatabaseHostPort:   os.Getenv("DATABASE_HOST_PORT"),
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

// GetDatabaseHost returns the configured database host
func GetDatabaseHost() string {
	return cfg.DatabaseHost
}

// GetDatabaseHostPort returns the configured database host port
func GetDatabaseHostPort() string {
	return cfg.DatabaseHostPort
}

// GetBotToken returns the configured bot token
func GetBotToken() string {
	return cfg.BotToken
}

// GetEmailUser returns the configured email user
func GetEmailUser() string {
	return cfg.EmailUser
}

// GetEmailPassword returns the configured email password
func GetEmailPassword() string {
	return cfg.EmailPassword
}

// IsDevMode returns whether the service is running in development mode
func IsDevMode() bool {
	return cfg.DevMode
}
