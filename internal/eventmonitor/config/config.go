package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

type Config struct {
	devMode bool

	// Service ports
	httpPort string
	grpcPort string

	// OTel exporter endpoint
	otelExporterEndpoint string

	// Condition Scheduler RPC URL
	conditionSchedulerRPCUrl string

	// Alchemy API Key
	alchemyAPIKey string

	// YAML-loaded settings
	polling PollingConfig
	webhook WebhookConfig
	metrics              yaml.MetricsConfig
	shutdown             yaml.ShutdownConfig
	version              yaml.VersionConfig
}

type PollingConfig struct {
	Interval       yaml.Duration `yaml:"interval"`
	BlockRange     int           `yaml:"block_range"`
	LookbackBlocks int           `yaml:"lookback_blocks"`
}

type WebhookConfig struct {
	Timeout    yaml.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
	RetryDelay yaml.Duration `yaml:"retry_delay"`
}

type YAMLConfig struct {
	Polling  PollingConfig  `yaml:"polling"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	Metrics  yaml.MetricsConfig `yaml:"metrics"`
	Shutdown yaml.ShutdownConfig `yaml:"shutdown"`
	Version  yaml.VersionConfig  `yaml:"version"`
}

var cfg Config

func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Load YAML config
	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = Config{
		devMode:                       env.GetEnvBool("DEV_MODE", false),
		httpPort:                      env.GetEnvString("EVENT_MONITOR_HTTP_PORT", "9008"),
		grpcPort:                      env.GetEnvString("EVENT_MONITOR_GRPC_PORT", "9018"),
		otelExporterEndpoint:          env.GetOTELExporterEndpoint(),
		conditionSchedulerRPCUrl: env.GetEnvString("CONDITION_SCHEDULER_RPC_URL", "localhost:9016"),
		alchemyAPIKey:            env.GetEnvString("EVENT_MONITOR_ALCHEMY_API_KEY", ""),
		polling:                  yamlConfig.Polling,
		webhook:                  yamlConfig.Webhook,
		metrics:                  yamlConfig.Metrics,
		shutdown:                 yamlConfig.Shutdown,
		version:                  yamlConfig.Version,
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
	if !env.IsValidPort(cfg.httpPort) {
		return fmt.Errorf("invalid DB Server HTTP Port: %s", cfg.httpPort)
	}
	if !env.IsValidPort(cfg.grpcPort) {
		return fmt.Errorf("invalid DB Server gRPC Port: %s", cfg.grpcPort)
	}
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	if !env.IsValidHostPort(cfg.conditionSchedulerRPCUrl) {
		return fmt.Errorf("invalid condition scheduler RPC URL: %s", cfg.conditionSchedulerRPCUrl)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy API key: %s", cfg.alchemyAPIKey)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetVersion() string {
	return cfg.version.Version
}

func GetHTTPPort() string {
	return cfg.httpPort
}

func GetGRPCPort() string {
	return cfg.grpcPort
}

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetMetricsUpdateInterval() time.Duration {
	return cfg.metrics.UpdateInterval.ToDuration()
}

func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func GetPollInterval() time.Duration {
	return cfg.polling.Interval.ToDuration()
}

func GetMaxBlockRange() uint64 {
	return uint64(cfg.polling.BlockRange)
}

func GetLookbackBlocks() uint64 {
	return uint64(cfg.polling.LookbackBlocks)
}

func GetWebhookTimeout() time.Duration {
	return cfg.webhook.Timeout.ToDuration()
}

func GetWebhookMaxRetries() int {
	return cfg.webhook.MaxRetries
}

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
