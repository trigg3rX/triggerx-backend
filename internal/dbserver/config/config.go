package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	timeSchedulerRPCUrl      string
	eventSchedulerRPCUrl     string
	conditionSchedulerRPCUrl string

	databaseRPCPort          string

	databaseHostAddress   	string
	databaseHostPort 		string

	emailUser     			string
	emailPassword 			string
	botToken     			string

	alchemyAPIKey 			string

	faucetPrivateKey 		string
	faucetFundAmount 		string

	upstashRedisUrl       	string
	upstashRedisRestToken 	string

	devMode bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		timeSchedulerRPCUrl:      env.GetEnv("TIME_SCHEDULER_RPC_URL", "http://localhost:9004"),
		eventSchedulerRPCUrl:     env.GetEnv("EVENT_SCHEDULER_RPC_URL", "http://localhost:9004"),
		conditionSchedulerRPCUrl: env.GetEnv("CONDITION_SCHEDULER_RPC_URL", "http://localhost:9005"),
		databaseRPCPort:          env.GetEnv("DATABASE_RPC_PORT", "9002"),
		databaseHostAddress:      env.GetEnv("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:         env.GetEnv("DATABASE_HOST_PORT", "9042"),
		emailUser:                env.GetEnv("EMAIL_USER", ""),
		emailPassword:            env.GetEnv("EMAIL_PASS", ""),
		botToken:                 env.GetEnv("BOT_TOKEN", ""),
		alchemyAPIKey:            env.GetEnv("ALCHEMY_API_KEY", ""),
		faucetPrivateKey:         env.GetEnv("FAUCET_PRIVATE_KEY", ""),
		faucetFundAmount:         env.GetEnv("FAUCET_FUND_AMOUNT", "30000000000000000"),
		upstashRedisUrl:          env.GetEnv("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken:    env.GetEnv("UPSTASH_REDIS_REST_TOKEN", ""),
		devMode:                  env.GetEnvBool("DEV_MODE", false),
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func validateConfig(cfg Config) error {
	if !validator.IsValidRPCUrl(cfg.timeSchedulerRPCUrl) {
		return fmt.Errorf("invalid time scheduler rpc address: %s", cfg.timeSchedulerRPCUrl)
	}
	if !validator.IsValidRPCUrl(cfg.eventSchedulerRPCUrl) {
		return fmt.Errorf("invalid event scheduler rpc address: %s", cfg.eventSchedulerRPCUrl)
	}
	if !validator.IsValidRPCUrl(cfg.conditionSchedulerRPCUrl) {
		return fmt.Errorf("invalid condition scheduler rpc address: %s", cfg.conditionSchedulerRPCUrl)
	}
	if !validator.IsValidPort(cfg.databaseRPCPort) {
		return fmt.Errorf("invalid database rpc port: %s", cfg.databaseRPCPort)
	}
	if !validator.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host: %s", cfg.databaseHostAddress)
	}
	if !validator.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if validator.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.alchemyAPIKey)
	}
	if !validator.IsValidPrivateKey(cfg.faucetPrivateKey) {
		return fmt.Errorf("invalid faucet private key: %s", cfg.faucetPrivateKey)
	}
	if validator.IsEmpty(cfg.faucetFundAmount) {
		return fmt.Errorf("invalid faucet fund amount: %s", cfg.faucetFundAmount)
	}
	if validator.IsEmpty(cfg.upstashRedisUrl) {
		return fmt.Errorf("invalid upstash redis url: %s", cfg.upstashRedisUrl)
	}
	if validator.IsEmpty(cfg.upstashRedisRestToken) {
		return fmt.Errorf("invalid upstash redis rest token: %s", cfg.upstashRedisRestToken)
	}
	if !cfg.devMode {
		if validator.IsValidEmail(cfg.emailUser) {
			return fmt.Errorf("invalid email user: %s", cfg.emailUser)
		}
		if validator.IsEmpty(cfg.emailPassword) {
			return fmt.Errorf("invalid email password: %s", cfg.emailPassword)
		}
		if validator.IsEmpty(cfg.botToken) {
			return fmt.Errorf("invalid bot token: %s", cfg.botToken)
		}
	}

	return nil
}

func GetTimeSchedulerRPCUrl() string {
	return cfg.timeSchedulerRPCUrl
}

func GetEventSchedulerRPCUrl() string {
	return cfg.eventSchedulerRPCUrl
}

func GetConditionSchedulerRPCUrl() string {
	return cfg.conditionSchedulerRPCUrl
}

func GetDatabaseRPCPort() string {
	return cfg.databaseRPCPort
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetEmailUser() string {
	return cfg.emailUser
}

func GetEmailPassword() string {
	return cfg.emailPassword
}

func GetBotToken() string {
	return cfg.botToken
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func GetFaucetPrivateKey() string {
	return cfg.faucetPrivateKey
}

func GetFaucetFundAmount() string {
	return cfg.faucetFundAmount
}

func GetUpstashRedisUrl() string {
	return cfg.upstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.upstashRedisRestToken
}

func IsDevMode() bool {
	return cfg.devMode
}
