package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Scheduler RPC URLs
	timeSchedulerRPCUrl      string
	conditionSchedulerRPCUrl string

	// Database RPC Port
	dbserverRPCPort string

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// Email User and Password
	emailUser     string
	emailPassword string
	botToken      string

	// API Keys for Alchemy
	alchemyAPIKey string

	// Faucet Private Key and Fund Amount
	faucetPrivateKey string
	faucetFundAmount string

	// Upstash Redis URL and Rest Token
	upstashRedisUrl       string
	upstashRedisRestToken string

	// Polling Look Ahead
	timeSchedulerPollingLookAhead int
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		timeSchedulerRPCUrl:      env.GetEnv("TIME_SCHEDULER_RPC_URL", "http://localhost:9005"),
		conditionSchedulerRPCUrl: env.GetEnv("CONDITION_SCHEDULER_RPC_URL", "http://localhost:9006"),
		dbserverRPCPort:          env.GetEnv("DBSERVER_RPC_PORT", "9002"),
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
		timeSchedulerPollingLookAhead: env.GetEnvInt("TIME_SCHEDULER_POLLING_LOOKAHEAD", 40),
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
	if !env.IsValidURL(cfg.timeSchedulerRPCUrl) {
		return fmt.Errorf("invalid time scheduler RPC URL: %s", cfg.timeSchedulerRPCUrl)
	}
	if !env.IsValidURL(cfg.conditionSchedulerRPCUrl) {
		return fmt.Errorf("invalid condition scheduler RPC URL: %s", cfg.conditionSchedulerRPCUrl)
	}
	if !env.IsValidPort(cfg.dbserverRPCPort) {
		return fmt.Errorf("invalid database RPC port: %s", cfg.dbserverRPCPort)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidPrivateKey(cfg.faucetPrivateKey) {
		return fmt.Errorf("invalid faucet private key: %s", cfg.faucetPrivateKey)
	}
	if env.IsEmpty(cfg.upstashRedisUrl) {
		return fmt.Errorf("invalid upstash redis url: %s", cfg.upstashRedisUrl)
	}
	if env.IsEmpty(cfg.upstashRedisRestToken) {
		return fmt.Errorf("invalid upstash redis rest token: %s", cfg.upstashRedisRestToken)
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

func GetTimeSchedulerRPCUrl() string {
	return cfg.timeSchedulerRPCUrl
}

func GetConditionSchedulerRPCUrl() string {
	return cfg.conditionSchedulerRPCUrl
}

func GetDBServerRPCPort() string {
	return cfg.dbserverRPCPort
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

func GetPollingLookAhead() int {
	return cfg.timeSchedulerPollingLookAhead
}
