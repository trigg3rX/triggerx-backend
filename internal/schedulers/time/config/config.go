package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

const (
	version = "0.0.1"
)

type Config struct {
	devMode              bool
	otelExporterEndpoint string

	// Scheduler RPC Port
	timeSchedulerRPCPort string

	// Task Dispatcher RPC URL
	taskDispatcherRPCUrl string

	// Scheduler ID
	timeSchedulerID int

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// ScyllaDB Authentication
	databaseUsername string
	databasePassword string

	// ScyllaDB SSL/TLS Configuration
	databaseSSLEnabled            bool
	databaseSSLCertPath           string
	databaseSSLKeyPath            string
	databaseSSLCAPath             string
	databaseSSLInsecureSkipVerify bool

	polling PollingConfig
}

type PollingConfig struct {
	Interval            yaml.Duration `yaml:"interval"`
	LookAhead           yaml.Duration `yaml:"look_ahead"`
	BatchSize           int           `yaml:"batch_size"`
	PerformerLockTTL    yaml.Duration `yaml:"performer_lock_ttl"`
	TaskCacheTTL        yaml.Duration `yaml:"task_cache_ttl"`
	DuplicateTaskWindow yaml.Duration `yaml:"duplicate_task_window"`
}

var cfg *Config

func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Load YAML config - need wrapper struct to match YAML structure
	type YAMLConfig struct {
		Polling PollingConfig `yaml:"polling"`
	}
	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = &Config{
		devMode:                       env.GetEnvBool("DEV_MODE", false),
		otelExporterEndpoint:          env.GetEnvString("OTEL_EXPORTER_ENDPOINT", "localhost:4318"),
		timeSchedulerRPCPort:          env.GetEnvString("TIME_SCHEDULER_RPC_PORT", "9005"),
		taskDispatcherRPCUrl:          env.GetEnvString("TASK_DISPATCHER_RPC_URL", "localhost:9003"),
		timeSchedulerID:               env.GetEnvInt("TIME_SCHEDULER_ID", 1234),
		databaseHostAddress:           env.GetEnvString("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:              env.GetEnvString("DATABASE_HOST_PORT", "9042"),
		databaseUsername:              env.GetEnvString("DATABASE_USERNAME", ""),
		databasePassword:              env.GetEnvString("DATABASE_PASSWORD", ""),
		databaseSSLEnabled:            env.GetEnvBool("DATABASE_SSL_ENABLED", false),
		databaseSSLCertPath:           env.GetEnvString("DATABASE_SSL_CERT_PATH", ""),
		databaseSSLKeyPath:            env.GetEnvString("DATABASE_SSL_KEY_PATH", ""),
		databaseSSLCAPath:             env.GetEnvString("DATABASE_SSL_CA_PATH", ""),
		databaseSSLInsecureSkipVerify: env.GetEnvBool("DATABASE_SSL_INSECURE_SKIP_VERIFY", false),
		polling:                       yamlConfig.Polling,
	}
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if err := yaml.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func validateConfig() error {
	if !env.IsValidPort(cfg.timeSchedulerRPCPort) {
		return fmt.Errorf("invalid time scheduler RPC port: %s", cfg.timeSchedulerRPCPort)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	// Validate polling configuration
	if cfg.polling.Interval.ToDuration() <= 0 {
		return fmt.Errorf("polling interval must be positive, got: %v", cfg.polling.Interval.ToDuration())
	}
	if cfg.polling.LookAhead.ToDuration() <= 0 {
		return fmt.Errorf("polling look ahead must be positive, got: %v", cfg.polling.LookAhead.ToDuration())
	}
	if cfg.polling.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive, got: %d", cfg.polling.BatchSize)
	}
	// Note: taskDispatcherRPCUrl is a gRPC endpoint (host:port format), not an HTTP URL
	// so we don't validate it as a URL
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

func GetSchedulerRPCPort() string {
	return cfg.timeSchedulerRPCPort
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetDatabaseUsername() string {
	return cfg.databaseUsername
}

func GetDatabasePassword() string {
	return cfg.databasePassword
}

func GetDatabaseSSLEnabled() bool {
	return cfg.databaseSSLEnabled
}

func GetDatabaseSSLCertPath() string {
	return cfg.databaseSSLCertPath
}

func GetDatabaseSSLKeyPath() string {
	return cfg.databaseSSLKeyPath
}

func GetDatabaseSSLCAPath() string {
	return cfg.databaseSSLCAPath
}

func GetDatabaseSSLInsecureSkipVerify() bool {
	return cfg.databaseSSLInsecureSkipVerify
}

func GetTaskDispatcherRPCUrl() string {
	return cfg.taskDispatcherRPCUrl
}

func GetSchedulerID() int {
	return cfg.timeSchedulerID
}

func GetPollingInterval() time.Duration {
	return cfg.polling.Interval.ToDuration()
}

func GetPollingLookAhead() time.Duration {
	return cfg.polling.LookAhead.ToDuration()
}

func GetTaskBatchSize() int {
	return cfg.polling.BatchSize
}

func GetPerformerLockTTL() time.Duration {
	return cfg.polling.PerformerLockTTL.ToDuration()
}

func GetTaskCacheTTL() time.Duration {
	return cfg.polling.TaskCacheTTL.ToDuration()
}

func GetDuplicateTaskWindow() time.Duration {
	return cfg.polling.DuplicateTaskWindow.ToDuration()
}
