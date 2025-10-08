package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

// Config represents the complete health service configuration
type Config struct {
	// Configuration loaded from YAML
	Server       ServerConfig       `yaml:"server" validate:"required"`
	HealthCheck  HealthCheckConfig  `yaml:"health_check" validate:"required"`
	Database     DatabaseConfig     `yaml:"database" validate:"required"`
	Notification NotificationConfig `yaml:"notification"`

	// Secrets loaded from .env
	DevMode bool
	Secrets SecretsConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	HTTPPort string `yaml:"http_port" validate:"required,port"`
	GRPCPort string `yaml:"grpc_port" validate:"required,port"`
}

// HealthCheckConfig contains health check related configuration
type HealthCheckConfig struct {
	KeeperTimeout string `yaml:"keeper_timeout" validate:"required,duration"`
	CheckInterval string `yaml:"check_interval" validate:"required,duration"`
	SyncInterval  string `yaml:"sync_interval" validate:"required,duration"`
	MaxRetries    int    `yaml:"max_retries" validate:"min=1,max=10"`
	RetryBackoff  string `yaml:"retry_backoff" validate:"duration"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	Host              string `yaml:"host" validate:"required,ip"`
	Port              string `yaml:"port" validate:"required,port"`
	Keyspace          string `yaml:"keyspace" validate:"required,min=1"`
	ReplicationFactor int    `yaml:"replication_factor" validate:"min=1,max=10"`
	ConsistencyLevel  string `yaml:"consistency_level" validate:"oneof=one|two|three|quorum|all|local_quorum|each_quorum|local_one"`
	ConnectionTimeout string `yaml:"connection_timeout" validate:"duration"`
	QueryTimeout      string `yaml:"query_timeout" validate:"duration"`
}

// NotificationConfig contains notification-related configuration
type NotificationConfig struct {
	SMTPHost      string `yaml:"smtp_host" validate:"ip"`
	SMTPPort      string `yaml:"smtp_port" validate:"port"`
	Timeout       string `yaml:"timeout" validate:"duration"`
	RetryAttempts int    `yaml:"retry_attempts" validate:"min=1,max=10"`
}

// SecretsConfig contains sensitive configuration loaded from environment variables
type SecretsConfig struct {
	// External API keys
	BotToken        string
	EmailUser       string
	EmailPassword   string
	EtherscanAPIKey string
	AlchemyAPIKey   string
	PinataHost      string
	PinataJWT       string

	// Ethereum addresses
	ManagerSigningAddress    string
	TaskExecutionAddress     string
	TestTaskExecutionAddress string
	ImuaTaskExecutionAddress string
}

var cfg *Config

// Init initializes the configuration by loading YAML config and environment secrets
func Init(configPath string) error {
	return InitWithEnvFile(configPath, ".env")
}

// InitWithEnvFile initializes the configuration with a custom env file path (for testing)
func InitWithEnvFile(configPath string, envFilePath string) error {
	// Load secrets from env file
	if err := godotenv.Load(envFilePath); err != nil {
		return fmt.Errorf("error loading env file %s: %w", envFilePath, err)
	}

	// Set default config path if not provided
	if configPath == "" {
		configPath = "config/health.yaml"
	}

	// Load configuration from YAML
	config := &Config{}
	if err := yaml.LoadYAML(configPath, config); err != nil {
		return fmt.Errorf("error loading YAML config from %s: %w", configPath, err)
	}

	// Load secrets from environment variables
	config.Secrets, config.DevMode = loadSecretsFromEnv()

	// Validate the complete configuration
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cfg = config

	// Set Gin mode based on dev mode setting
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

// loadSecretsFromEnv loads sensitive configuration from environment variables
func loadSecretsFromEnv() (SecretsConfig, bool) {
	return SecretsConfig{
		// External API keys
		BotToken:        env.GetEnvString("BOT_TOKEN", ""),
		EmailUser:       env.GetEnvString("EMAIL_USER", ""),
		EmailPassword:   env.GetEnvString("EMAIL_PASS", ""),
		EtherscanAPIKey: env.GetEnvString("ETHERSCAN_API_KEY", ""),
		AlchemyAPIKey:   env.GetEnvString("ALCHEMY_API_KEY", ""),
		PinataHost:      env.GetEnvString("PINATA_HOST", ""),
		PinataJWT:       env.GetEnvString("PINATA_JWT", ""),

		// Ethereum addresses
		ManagerSigningAddress:    env.GetEnvString("MANAGER_SIGNING_ADDRESS", ""),
		TaskExecutionAddress:     env.GetEnvString("TASK_EXECUTION_ADDRESS", ""),
		TestTaskExecutionAddress: env.GetEnvString("TEST_TASK_EXECUTION_ADDRESS", ""),
		ImuaTaskExecutionAddress: env.GetEnvString("IMUA_TASK_EXECUTION_ADDRESS", ""),
	}, env.GetEnvBool("DEV_MODE", false)
}

// validateConfig validates the complete configuration
func validateConfig(config *Config) error {
	// Validate YAML configuration using the validator
	if err := yaml.ValidateConfig(config); err != nil {
		return fmt.Errorf("yaml validation failed: %w", err)
	}

	// Validate secrets if not in dev mode
	if !config.DevMode {
		if err := validateSecrets(config.Secrets); err != nil {
			return fmt.Errorf("secrets validation failed: %w", err)
		}
	}

	return nil
}

// validateSecrets validates sensitive configuration values
func validateSecrets(secrets SecretsConfig) error {
	// Validate required secrets
	if env.IsEmpty(secrets.PinataHost) {
		return fmt.Errorf("pinata host is required")
	}
	if env.IsEmpty(secrets.PinataJWT) {
		return fmt.Errorf("pinata JWT is required")
	}
	if env.IsEmpty(secrets.EtherscanAPIKey) {
		return fmt.Errorf("etherscan API key is required")
	}
	if env.IsEmpty(secrets.AlchemyAPIKey) {
		return fmt.Errorf("alchemy API key is required")
	}
	if !env.IsValidEthAddress(secrets.TaskExecutionAddress) {
		return fmt.Errorf("invalid task execution address: %s", secrets.TaskExecutionAddress)
	}
	if !env.IsValidEthAddress(secrets.TestTaskExecutionAddress) {
		return fmt.Errorf("invalid test task execution address: %s", secrets.TestTaskExecutionAddress)
	}
	if !env.IsValidEthAddress(secrets.ImuaTaskExecutionAddress) {
		return fmt.Errorf("invalid IMUA task execution address: %s", secrets.ImuaTaskExecutionAddress)
	}
	if !env.IsValidEthAddress(secrets.ManagerSigningAddress) {
		return fmt.Errorf("invalid manager signing address: %s", secrets.ManagerSigningAddress)
	}
	if !env.IsValidEmail(secrets.EmailUser) {
		return fmt.Errorf("invalid email user: %s", secrets.EmailUser)
	}
	if env.IsEmpty(secrets.EmailPassword) {
		return fmt.Errorf("email password is required")
	}
	if env.IsEmpty(secrets.BotToken) {
		return fmt.Errorf("bot token is required")
	}

	return nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	return cfg
}

// Server Configuration Getters
func GetHTTPPort() string {
	return cfg.Server.HTTPPort
}

func GetGRPCPort() string {
	return cfg.Server.GRPCPort
}

func IsDevMode() bool {
	return cfg.DevMode
}

// Health Check Configuration Getters
func GetKeeperTimeout() time.Duration {
	duration, _ := time.ParseDuration(cfg.HealthCheck.KeeperTimeout)
	return duration
}

func GetCheckInterval() time.Duration {
	duration, _ := time.ParseDuration(cfg.HealthCheck.CheckInterval)
	return duration
}

func GetSyncInterval() time.Duration {
	duration, _ := time.ParseDuration(cfg.HealthCheck.SyncInterval)
	return duration
}

func GetMaxRetries() int {
	return cfg.HealthCheck.MaxRetries
}

func GetRetryBackoff() time.Duration {
	duration, _ := time.ParseDuration(cfg.HealthCheck.RetryBackoff)
	return duration
}

// Database Configuration Getters
func GetDatabaseHost() string {
	return cfg.Database.Host
}

func GetDatabasePort() string {
	return cfg.Database.Port
}

func GetDatabaseKeyspace() string {
	return cfg.Database.Keyspace
}

func GetDatabaseReplicationFactor() int {
	return cfg.Database.ReplicationFactor
}

func GetDatabaseConsistencyLevel() string {
	return cfg.Database.ConsistencyLevel
}

func GetDatabaseConnectionTimeout() time.Duration {
	duration, _ := time.ParseDuration(cfg.Database.ConnectionTimeout)
	return duration
}

func GetDatabaseQueryTimeout() time.Duration {
	duration, _ := time.ParseDuration(cfg.Database.QueryTimeout)
	return duration
}

// Notification Configuration Getters
func GetSMTPHost() string {
	return cfg.Notification.SMTPHost
}

func GetSMTPPort() string {
	return cfg.Notification.SMTPPort
}
func GetNotificationTimeout() time.Duration {
	duration, _ := time.ParseDuration(cfg.Notification.Timeout)
	return duration
}

func GetNotificationRetryAttempts() int {
	return cfg.Notification.RetryAttempts
}

// Secrets Getters
func GetBotToken() string {
	return cfg.Secrets.BotToken
}

func GetEmailUser() string {
	return cfg.Secrets.EmailUser
}

func GetEmailPassword() string {
	return cfg.Secrets.EmailPassword
}

func GetEtherscanAPIKey() string {
	return cfg.Secrets.EtherscanAPIKey
}

func GetAlchemyAPIKey() string {
	return cfg.Secrets.AlchemyAPIKey
}

func GetPinataHost() string {
	return cfg.Secrets.PinataHost
}

func GetPinataJWT() string {
	return cfg.Secrets.PinataJWT
}

func GetManagerSigningAddress() string {
	return cfg.Secrets.ManagerSigningAddress
}

func GetTaskExecutionAddress() string {
	return cfg.Secrets.TaskExecutionAddress
}

func GetTestTaskExecutionAddress() string {
	return cfg.Secrets.TestTaskExecutionAddress
}

func GetImuaTaskExecutionAddress() string {
	return cfg.Secrets.ImuaTaskExecutionAddress
}
