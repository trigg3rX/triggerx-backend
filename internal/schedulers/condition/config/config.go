package config

import (
	"fmt"

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
	conditionSchedulerRPCPort string

	// Service RPC URL
	taskDispatcherRPCUrl string
	eventMonitorRPCUrl   string

	// Scheduler ID for consumer groups
	conditionSchedulerID int

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

	// Workers Config
	workers WorkersConfig
}

type WorkersConfig struct {
	MaxWorkers int `yaml:"max_workers"`
}

var cfg *Config

// Init initializes the configuration
func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Load YAML config - need wrapper struct to match YAML structure
	type YAMLConfig struct {
		Workers WorkersConfig `yaml:"workers"`
	}
	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = &Config{
		devMode:                       env.GetEnvBool("DEV_MODE", false),
		otelExporterEndpoint:          env.GetEnvString("OTEL_EXPORTER_ENDPOINT", "localhost:4318"),
		conditionSchedulerRPCPort:     env.GetEnvString("CONDITION_SCHEDULER_RPC_PORT", "9006"),
		taskDispatcherRPCUrl:          env.GetEnvString("TASK_DISPATCHER_RPC_URL", "localhost:9003"),
		eventMonitorRPCUrl:            env.GetEnvString("EVENT_MONITOR_RPC_URL", "localhost:9009"),
		conditionSchedulerID:          env.GetEnvInt("CONDITION_SCHEDULER_ID", 1234),
		databaseHostAddress:           env.GetEnvString("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:              env.GetEnvString("DATABASE_HOST_PORT", "9042"),
		databaseUsername:              env.GetEnvString("DATABASE_USERNAME", ""),
		databasePassword:              env.GetEnvString("DATABASE_PASSWORD", ""),
		databaseSSLEnabled:            env.GetEnvBool("DATABASE_SSL_ENABLED", false),
		databaseSSLCertPath:           env.GetEnvString("DATABASE_SSL_CERT_PATH", ""),
		databaseSSLKeyPath:            env.GetEnvString("DATABASE_SSL_KEY_PATH", ""),
		databaseSSLCAPath:             env.GetEnvString("DATABASE_SSL_CA_PATH", ""),
		databaseSSLInsecureSkipVerify: env.GetEnvBool("DATABASE_SSL_INSECURE_SKIP_VERIFY", false),
		workers:                       yamlConfig.Workers,
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
	if !env.IsValidPort(cfg.conditionSchedulerRPCPort) {
		return fmt.Errorf("invalid condition scheduler RPC port: %s", cfg.conditionSchedulerRPCPort)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if cfg.workers.MaxWorkers <= 0 {
		return fmt.Errorf("max workers must be positive, got: %d", cfg.workers.MaxWorkers)
	}
	if !env.IsValidHostPort(cfg.taskDispatcherRPCUrl) {
		return fmt.Errorf("invalid task dispatcher RPC URL: %s", cfg.taskDispatcherRPCUrl)
	}
	if !env.IsValidHostPort(cfg.eventMonitorRPCUrl) {
		return fmt.Errorf("invalid event monitor RPC URL: %s", cfg.eventMonitorRPCUrl)
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

func GetSchedulerRPCPort() string {
	return cfg.conditionSchedulerRPCPort
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

func GetEventMonitorRPCUrl() string {
	return cfg.eventMonitorRPCUrl
}

func GetSchedulerID() int {
	return cfg.conditionSchedulerID
}

func GetMaxWorkers() int {
	return cfg.workers.MaxWorkers
}
