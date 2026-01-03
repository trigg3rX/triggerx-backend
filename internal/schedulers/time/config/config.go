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

	// Database Connection Configuration (from env)
	dbConnection env.DatabaseConfig

	// OTel exporter endpoint
	otelExporterEndpoint string

	// Task Dispatcher RPC URL
	taskDispatcherRPCUrl string

	// Scheduler ID
	timeSchedulerID int

	// YAML-loaded settings	
	polling PollingConfig
	metrics              yaml.MetricsConfig
	shutdown             yaml.ShutdownConfig
	version              yaml.VersionConfig
}

type PollingConfig struct {
	Interval            yaml.Duration `yaml:"interval"`
	LookAhead           yaml.Duration `yaml:"look_ahead"`
	BatchSize           int           `yaml:"batch_size"`
	PerformerLockTTL    yaml.Duration `yaml:"performer_lock_ttl"`
	TaskCacheTTL        yaml.Duration `yaml:"task_cache_ttl"`
	DuplicateTaskWindow yaml.Duration `yaml:"duplicate_task_window"`
}

type YAMLConfig struct {
	Polling  PollingConfig  `yaml:"polling"`
	Metrics  yaml.MetricsConfig  `yaml:"metrics"`
	Shutdown yaml.ShutdownConfig `yaml:"shutdown"`
	Version  yaml.VersionConfig  `yaml:"version"`
}

var cfg *Config

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
		devMode:              env.GetEnvBool("DEV_MODE", false),
		httpPort:             env.GetEnvString("TIME_SCHEDULER_HTTP_PORT", "9005"),
		grpcPort:             env.GetEnvString("TIME_SCHEDULER_GRPC_PORT", "9015"),
		dbConnection:         env.GetDatabaseConfig(),
		otelExporterEndpoint: env.GetOTELExporterEndpoint(),
		taskDispatcherRPCUrl: env.GetEnvString("TASK_DISPATCHER_RPC_URL", "localhost:9017"),
		timeSchedulerID:      env.GetEnvInt("TIME_SCHEDULER_ID", 1234),
		polling:              yamlConfig.Polling,
		shutdown:             yamlConfig.Shutdown,
		version:              yamlConfig.Version,
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
		return fmt.Errorf("invalid time scheduler HTTP Port: %s", cfg.httpPort)
	}
	if !env.IsValidPort(cfg.grpcPort) {
		return fmt.Errorf("invalid time scheduler gRPC Port: %s", cfg.grpcPort)
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
	if !env.IsValidInt(cfg.timeSchedulerID) {
		return fmt.Errorf("invalid time scheduler ID: %d", cfg.timeSchedulerID)
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

func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}
