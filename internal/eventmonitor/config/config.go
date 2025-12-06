package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	// Service Configuration
	Port string
	Host string

	// RPC Configuration
	AlchemyAPIKey string

	// Polling Configuration
	PollInterval   time.Duration
	MaxBlockRange  uint64
	LookbackBlocks uint64

	// Webhook Configuration
	WebhookTimeout    time.Duration
	WebhookMaxRetries int
	WebhookRetryDelay time.Duration

	// Logging
	LogLevel string
	DevMode  bool
}

var cfg Config

// Init initializes the configuration
func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		Port:              env.GetEnvString("EVENT_MONITOR_PORT", "9007"),
		Host:              env.GetEnvString("EVENT_MONITOR_HOST", "0.0.0.0"),
		AlchemyAPIKey:     env.GetEnvString("ALCHEMY_API_KEY", ""),
		PollInterval:      parseDuration(env.GetEnvString("POLL_INTERVAL", "1s")),
		MaxBlockRange:     uint64(env.GetEnvInt("MAX_BLOCK_RANGE", 10)),
		LookbackBlocks:    uint64(env.GetEnvInt("LOOKBACK_BLOCKS", 100)),
		WebhookTimeout:    parseDuration(env.GetEnvString("WEBHOOK_TIMEOUT", "5s")),
		WebhookMaxRetries: env.GetEnvInt("WEBHOOK_MAX_RETRIES", 3),
		WebhookRetryDelay: parseDuration(env.GetEnvString("WEBHOOK_RETRY_DELAY", "1s")),
		LogLevel:          env.GetEnvString("LOG_LEVEL", "info"),
		DevMode:           env.GetEnvBool("DEV_MODE", false),
	}

	return validateConfig()
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Second // default to 1 second
	}
	return d
}

func validateConfig() error {
	if !env.IsValidPort(cfg.Port) {
		return fmt.Errorf("invalid port: %s", cfg.Port)
	}
	return nil
}

// GetPort returns the service port
func GetPort() string {
	return cfg.Port
}

// GetHost returns the service host
func GetHost() string {
	return cfg.Host
}

// GetAlchemyAPIKey returns the Alchemy API key
func GetAlchemyAPIKey() string {
	return cfg.AlchemyAPIKey
}

// GetPollInterval returns the polling interval
func GetPollInterval() time.Duration {
	return cfg.PollInterval
}

// GetMaxBlockRange returns the maximum block range per query
func GetMaxBlockRange() uint64 {
	return cfg.MaxBlockRange
}

// GetLookbackBlocks returns the number of blocks to look back on startup
func GetLookbackBlocks() uint64 {
	return cfg.LookbackBlocks
}

// GetWebhookTimeout returns the webhook timeout
func GetWebhookTimeout() time.Duration {
	return cfg.WebhookTimeout
}

// GetWebhookMaxRetries returns the maximum webhook retries
func GetWebhookMaxRetries() int {
	return cfg.WebhookMaxRetries
}

// GetWebhookRetryDelay returns the webhook retry delay
func GetWebhookRetryDelay() time.Duration {
	return cfg.WebhookRetryDelay
}

// GetLogLevel returns the log level
func GetLogLevel() string {
	return cfg.LogLevel
}

// IsDevMode returns whether the service is in dev mode
func IsDevMode() bool {
	return cfg.DevMode
}

// GetChainRPCUrls returns chain RPC URLs
func GetChainRPCUrls() map[string]string {
	if cfg.AlchemyAPIKey == "" {
		// Fallback to public endpoints if no Alchemy key
		return map[string]string{
			"11155420": "https://sepolia.optimism.io",
			"84532":    "https://sepolia.base.org",
			"11155111": "https://ethereum-sepolia.publicnode.com",
			"421614":   "https://sepolia-rollup.arbitrum.io/rpc",
		}
	}

	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.AlchemyAPIKey),
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.AlchemyAPIKey),
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.AlchemyAPIKey),
		"421614":   fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", cfg.AlchemyAPIKey),
	}
}
