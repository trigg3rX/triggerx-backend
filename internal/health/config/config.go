package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Port at which health service will be running
	healthRPCPort string

	// Bot token for Telegram notifications
	botToken string
	// Email user for notifications
	emailUser     string
	emailPassword string

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// IPFS configuration
	pinataHost  string
	pinataJWT string

	// Etherscan API Key
	etherscanAPIKey string

	// Alchemy API Key
	alchemyAPIKey string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:             env.GetEnvBool("DEV_MODE", false),
		healthRPCPort:       env.GetEnv("HEALTH_RPC_PORT", "9003"),
		botToken:            env.GetEnv("BOT_TOKEN", ""),
		emailUser:           env.GetEnv("EMAIL_USER", ""),
		emailPassword:       env.GetEnv("EMAIL_PASS", ""),
		databaseHostAddress: env.GetEnv("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:    env.GetEnv("DATABASE_HOST_PORT", "9042"),
		pinataHost:            env.GetEnv("PINATA_HOST", ""),
		pinataJWT:           env.GetEnv("PINATA_JWT", ""),
		etherscanAPIKey:     env.GetEnv("ETHERSCAN_API_KEY", ""),
		alchemyAPIKey:       env.GetEnv("ALCHEMY_API_KEY", ""),
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
	if !env.IsValidPort(cfg.healthRPCPort) {
		return fmt.Errorf("invalid Health RPC Port: %s", cfg.healthRPCPort)
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("invalid Pinata Host: %s", cfg.pinataHost)
	}
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("invalid Pinata JWT: %s", cfg.pinataJWT)
	}
	if env.IsEmpty(cfg.etherscanAPIKey) {
		return fmt.Errorf("invalid Etherscan API Key: %s", cfg.etherscanAPIKey)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid Alchemy API Key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if !cfg.devMode {
		if !env.IsValidEmail(cfg.emailUser) {
			return fmt.Errorf("invalid email user: %s", cfg.emailUser)
		}
		if env.IsEmpty(cfg.emailPassword) {
			return fmt.Errorf("invalid email password: %s", cfg.emailPassword)
		}
		if env.IsEmpty(cfg.botToken) {
			return fmt.Errorf("invalid bot token: %s", cfg.botToken)
		}
	}
	return nil
}

func GetHealthRPCPort() string {
	return cfg.healthRPCPort
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetBotToken() string {
	return cfg.botToken
}

func GetEmailUser() string {
	return cfg.emailUser
}

func GetEmailPassword() string {
	return cfg.emailPassword
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}