package config

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

type Config struct {
	ManagerRPCAddress            string
	EventSchedulerRPCAddress     string
	ConditionSchedulerRPCAddress string
	DatabaseRPCPort              string

	DatabaseHost     string
	DatabaseHostPort string

	EmailUser     string
	EmailPassword string
	BotToken      string

	AlchemyAPIKey string

	FaucetPrivateKey string
	FaucetFundAmount string

	UpstashRedisUrl       string
	UpstashRedisRestToken string

	DevMode bool
}

var cfg Config

// getEnvWithDefault returns the environment variable value or a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode:                      os.Getenv("DEV_MODE") == "true",
		ManagerRPCAddress:            os.Getenv("MANAGER_RPC_ADDRESS"),
		EventSchedulerRPCAddress:     getEnvWithDefault("EVENT_SCHEDULER_RPC_ADDRESS", "http://localhost:9004"),
		ConditionSchedulerRPCAddress: getEnvWithDefault("CONDITION_SCHEDULER_RPC_ADDRESS", "http://localhost:9005"),
		DatabaseRPCPort:              os.Getenv("DATABASE_RPC_PORT"),
		DatabaseHost:                 os.Getenv("DATABASE_HOST"),
		DatabaseHostPort:             os.Getenv("DATABASE_HOST_PORT"),
		EmailUser:                    os.Getenv("EMAIL_USER"),
		EmailPassword:                os.Getenv("EMAIL_PASSWORD"),
		BotToken:                     os.Getenv("BOT_TOKEN"),
		AlchemyAPIKey:                os.Getenv("ALCHEMY_API_KEY"),
		FaucetPrivateKey:             os.Getenv("FAUCET_PRIVATE_KEY"),
		FaucetFundAmount:             os.Getenv("FAUCET_FUND_AMOUNT"),
		UpstashRedisUrl:              os.Getenv("UPSTASH_REDIS_URL"),
		UpstashRedisRestToken:        os.Getenv("UPSTASH_REDIS_REST_TOKEN"),
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func validateConfig(cfg Config) error {
	if !validator.IsValidRPCAddress(cfg.ManagerRPCAddress) {
		return fmt.Errorf("invalid manager rpc address: %s", cfg.ManagerRPCAddress)
	}
	if !validator.IsValidRPCAddress(cfg.EventSchedulerRPCAddress) {
		return fmt.Errorf("invalid event scheduler rpc address: %s", cfg.EventSchedulerRPCAddress)
	}
	if !validator.IsValidRPCAddress(cfg.ConditionSchedulerRPCAddress) {
		return fmt.Errorf("invalid condition scheduler rpc address: %s", cfg.ConditionSchedulerRPCAddress)
	}
	if !validator.IsValidPort(cfg.DatabaseRPCPort) {
		return fmt.Errorf("invalid database rpc port: %s", cfg.DatabaseRPCPort)
	}
	if !validator.IsValidIPAddress(cfg.DatabaseHost) {
		return fmt.Errorf("invalid database host: %s", cfg.DatabaseHost)
	}
	if !validator.IsValidPort(cfg.DatabaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.DatabaseHostPort)
	}
	if validator.IsValidEmail(cfg.EmailUser) {
		return fmt.Errorf("invalid email user: %s", cfg.EmailUser)
	}
	if !cfg.DevMode {
		if validator.IsEmpty(cfg.EmailPassword) {
			return fmt.Errorf("invalid email password: %s", cfg.EmailPassword)
		}
		if validator.IsEmpty(cfg.BotToken) {
			return fmt.Errorf("invalid bot token: %s", cfg.BotToken)
		}
		if validator.IsEmpty(cfg.AlchemyAPIKey) {
			return fmt.Errorf("invalid alchemy api key: %s", cfg.AlchemyAPIKey)
		}
	}
	if !validator.IsValidPrivateKey(cfg.FaucetPrivateKey) {
		return fmt.Errorf("invalid faucet private key: %s", cfg.FaucetPrivateKey)
	}
	if validator.IsEmpty(cfg.FaucetFundAmount) {
		return fmt.Errorf("invalid faucet fund amount: %s", cfg.FaucetFundAmount)
	}
	if validator.IsEmpty(cfg.UpstashRedisUrl) {
		return fmt.Errorf("invalid upstash redis url: %s", cfg.UpstashRedisUrl)
	}
	if validator.IsEmpty(cfg.UpstashRedisRestToken) {
		return fmt.Errorf("invalid upstash redis rest token: %s", cfg.UpstashRedisRestToken)
	}

	return nil
}

func GetManagerRPCAddress() string {
	return cfg.ManagerRPCAddress
}

func GetEventSchedulerRPCAddress() string {
	return cfg.EventSchedulerRPCAddress
}

func GetConditionSchedulerRPCAddress() string {
	return cfg.ConditionSchedulerRPCAddress
}

func GetDatabaseRPCPort() string {
	return cfg.DatabaseRPCPort
}

func GetDatabaseHost() string {
	return cfg.DatabaseHost
}

func GetDatabaseHostPort() string {
	return cfg.DatabaseHostPort
}

func GetEmailUser() string {
	return cfg.EmailUser
}

func GetEmailPassword() string {
	return cfg.EmailPassword
}

func GetBotToken() string {
	return cfg.BotToken
}

func GetAlchemyAPIKey() string {
	return cfg.AlchemyAPIKey
}

func GetFaucetPrivateKey() string {
	return cfg.FaucetPrivateKey
}

func GetFaucetFundAmount() string {
	return cfg.FaucetFundAmount
}

func GetUpstashRedisUrl() string {
	return cfg.UpstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.UpstashRedisRestToken
}

func IsDevMode() bool {
	return cfg.DevMode
}
