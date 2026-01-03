package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

type Config struct {
	devMode bool

	// Service ports
	httpPort string
	grpcPort string

	// Database Connection Configuration (from env)
	dbConnection env.DatabaseConfig

	// OTel exporter endpoint
	otelExporterEndpoint string

	// Service RPC URL
	taskDispatcherRPCUrl string
	eventMonitorRPCUrl   string

	// Scheduler ID for consumer groups
	conditionSchedulerID int

	// YAML-loaded settings
	workers WorkersConfig
	metrics              yaml.MetricsConfig
	shutdown             yaml.ShutdownConfig
	version              yaml.VersionConfig
}

type WorkersConfig struct {
	MaxWorkers int `yaml:"max_workers"`
}

type YAMLConfig struct {
	Workers  WorkersConfig  `yaml:"workers"`
	Metrics  yaml.MetricsConfig `yaml:"metrics"`
	Shutdown yaml.ShutdownConfig `yaml:"shutdown"`
	Version  yaml.VersionConfig  `yaml:"version"`
}

var cfg *Config

// Init initializes the configuration
func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = &Config{
		devMode:                   env.GetEnvBool("DEV_MODE", false),
		httpPort:                  env.GetEnvString("CONDITION_SCHEDULER_HTTP_PORT", "9006"),
		grpcPort:                  env.GetEnvString("CONDITION_SCHEDULER_GRPC_PORT", "9016"),
		dbConnection:              env.GetDatabaseConfig(),
		otelExporterEndpoint:      env.GetOTELExporterEndpoint(),
		taskDispatcherRPCUrl:      env.GetEnvString("TASK_DISPATCHER_RPC_URL", "localhost:9017"),
		eventMonitorRPCUrl:        env.GetEnvString("EVENT_MONITOR_RPC_URL", "localhost:9018"),
		conditionSchedulerID:      env.GetEnvInt("CONDITION_SCHEDULER_ID", 1234),
		workers:                   yamlConfig.Workers,
		metrics:                   yamlConfig.Metrics,
		shutdown:                  yamlConfig.Shutdown,
		version:                   yamlConfig.Version,
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
	if !env.IsValidPort(cfg.httpPort) {
		return fmt.Errorf("invalid condition scheduler HTTP Port: %s", cfg.httpPort)
	}
	if !env.IsValidPort(cfg.grpcPort) {
		return fmt.Errorf("invalid condition scheduler gRPC Port: %s", cfg.grpcPort)
	}
	if !env.IsValidIPAddress(cfg.dbConnection.HostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.dbConnection.HostAddress)
	}
	if !env.IsValidPort(cfg.dbConnection.HostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.dbConnection.HostPort)
	}
	if !env.IsValidHostPort(cfg.taskDispatcherRPCUrl) {
		return fmt.Errorf("invalid task dispatcher RPC URL: %s", cfg.taskDispatcherRPCUrl)
	}
	if !env.IsValidHostPort(cfg.eventMonitorRPCUrl) {
		return fmt.Errorf("invalid event monitor RPC URL: %s", cfg.eventMonitorRPCUrl)
	}
	if !env.IsValidInt(cfg.conditionSchedulerID) {
		return fmt.Errorf("invalid condition scheduler ID: %d", cfg.conditionSchedulerID)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetVersion() string {
	return cfg.version.Version
}

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetHTTPPort() string {
	return cfg.httpPort
}

func GetGRPCPort() string {
	return cfg.grpcPort
}

func GetDatabaseHostAddress() string {
	return cfg.dbConnection.HostAddress
}

func GetDatabaseHostPort() string {
	return cfg.dbConnection.HostPort
}

func GetDatabaseUsername() string {
	return cfg.dbConnection.Username
}

func GetDatabasePassword() string {
	return cfg.dbConnection.Password
}

func GetDatabaseSSLEnabled() bool {
	return cfg.dbConnection.SSLEnabled
}

func GetDatabaseSSLCertPath() string {
	return cfg.dbConnection.SSLCertPath
}

func GetDatabaseSSLKeyPath() string {
	return cfg.dbConnection.SSLKeyPath
}

func GetDatabaseSSLCAPath() string {
	return cfg.dbConnection.SSLCAPath
}

func GetDatabaseSSLInsecureSkipVerify() bool {
	return cfg.dbConnection.SSLInsecureSkipVerify
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

func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}
