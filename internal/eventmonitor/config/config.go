package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

const (
	version = "0.0.1"
)

type Config struct {
	devMode bool
	otelExporterEndpoint string

	// Event Monitor RPC Port
	eventMonitorRPCPort string

	// Condition Scheduler RPC URL
	conditionSchedulerRPCUrl string

	// Alchemy API Key
	alchemyAPIKey string

	// Polling Configuration
	polling PollingConfig

	// Webhook Configuration
	webhook WebhookConfig
}

type PollingConfig struct {
	Interval            yaml.Duration `yaml:"interval"`
	BlockRange          int           `yaml:"block_range"`
	LookbackBlocks      int           `yaml:"lookback_blocks"`
}

type WebhookConfig struct {
	Timeout            yaml.Duration `yaml:"timeout"`
	MaxRetries         int           `yaml:"max_retries"`
	RetryDelay         yaml.Duration `yaml:"retry_delay"`
}

var cfg Config

// Init initializes the configuration
func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Load YAML config - need wrapper struct to match YAML structure
	type YAMLConfig struct {
		Polling PollingConfig `yaml:"polling"`
		Webhook WebhookConfig `yaml:"webhook"`
	}
	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = Config{
		devMode:           env.GetEnvBool("DEV_MODE", false),
		otelExporterEndpoint:         env.GetEnvString("OTEL_EXPORTER_ENDPOINT", "localhost:4318"),
		eventMonitorRPCPort:              env.GetEnvString("EVENT_MONITOR_PORT", "9009"),
		conditionSchedulerRPCUrl: env.GetEnvString("CONDITION_SCHEDULER_RPC_URL", "localhost:9006"),
		alchemyAPIKey: env.GetEnvString("ALCHEMY_API_KEY", ""),
		polling: yamlConfig.Polling,
		webhook: yamlConfig.Webhook,
	}

	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if err := yaml.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}

func validateConfig() error {
	if !env.IsValidPort(cfg.eventMonitorRPCPort) {
		return fmt.Errorf("invalid event monitor RPC port: %s", cfg.eventMonitorRPCPort)
	}
	if !env.IsValidHostPort(cfg.conditionSchedulerRPCUrl) {
		return fmt.Errorf("invalid condition scheduler RPC URL: %s", cfg.conditionSchedulerRPCUrl)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy API key: %s", cfg.alchemyAPIKey)
	}
	if cfg.polling.Interval.ToDuration() <= 0 {
		return fmt.Errorf("invalid polling interval: %s", cfg.polling.Interval.ToDuration())
	}
	if cfg.polling.BlockRange <= 0 {
		return fmt.Errorf("invalid block range: %d", cfg.polling.BlockRange)
	}
	if cfg.polling.LookbackBlocks <= 0 {
		return fmt.Errorf("invalid lookback blocks: %d", cfg.polling.LookbackBlocks)
	}
	if cfg.webhook.Timeout.ToDuration() <= 0 {
		return fmt.Errorf("invalid webhook timeout: %s", cfg.webhook.Timeout.ToDuration())
	}
	if cfg.webhook.MaxRetries <= 0 {
		return fmt.Errorf("invalid webhook max retries: %d", cfg.webhook.MaxRetries)
	}
	if cfg.webhook.RetryDelay.ToDuration() <= 0 {
		return fmt.Errorf("invalid webhook retry delay: %s", cfg.webhook.RetryDelay.ToDuration())
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetVersion() string {
	return version
}

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetEventMonitorRPCPort() string {
	return cfg.eventMonitorRPCPort
}

// GetAlchemyAPIKey returns the Alchemy API key
func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

// GetPollInterval returns the polling interval
func GetPollInterval() time.Duration {
	return cfg.polling.Interval.ToDuration()
}

// GetMaxBlockRange returns the maximum block range per query
func GetMaxBlockRange() uint64 {
	return uint64(cfg.polling.BlockRange)
}

// GetLookbackBlocks returns the number of blocks to look back on startup
func GetLookbackBlocks() uint64 {
	return uint64(cfg.polling.LookbackBlocks)
}

// GetWebhookTimeout returns the webhook timeout
func GetWebhookTimeout() time.Duration {
	return cfg.webhook.Timeout.ToDuration()
}

// GetWebhookMaxRetries returns the maximum webhook retries
func GetWebhookMaxRetries() int {
	return cfg.webhook.MaxRetries
}

// GetWebhookRetryDelay returns the webhook retry delay
func GetWebhookRetryDelay() time.Duration {
	return cfg.webhook.RetryDelay.ToDuration()
}

func GetChainRPCUrls() map[string]string {
	if cfg.alchemyAPIKey == "" {
		return map[string]string{
			"11155420": "https://sepolia.optimism.io",
			"84532":    "https://sepolia.base.org",
			"11155111": "https://ethereum-sepolia.publicnode.com",
			"421614":   "https://sepolia-rollup.arbitrum.io/rpc",
		}
	}
	
	return map[string]string{
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
		"84532":    fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
		"421614":   fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", cfg.alchemyAPIKey),
	}
}
