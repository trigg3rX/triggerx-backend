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

	// Redis Configuration (from env)
	upstashRedisUrl       string
	upstashRedisRestToken string

	// OTel exporter endpoint
	otelExporterEndpoint string

	// Service ID for OpenTelemetry
	serviceID string

	// Bot token for Telegram notifications
	botToken string
	// Email user for notifications
	emailUser     string
	emailPassword string

	// IPFS configuration
	pinataHost string
	pinataJWT  string

	// Dispatcher Signing Address
	dispatcherSigningAddress string

	// Etherscan API Key
	etherscanAPIKey string

	// Alchemy API Key
	alchemyAPIKey string

	// Task Execution Address
	taskExecutionAddress     string
	testTaskExecutionAddress string
	imuaTaskExecutionAddress string

	// YAML-loaded settings
	healthCheck        HealthCheckConfig
	notification       NotificationConfig
	rpc                RPCConfig
	databaseOperations yaml.DatabaseOperationsConfig
	metrics            yaml.MetricsConfig
	shutdown           yaml.ShutdownConfig
	version            yaml.VersionConfig
	keeperVersions     KeeperVersionsConfig
}

type HealthCheckConfig struct {
	KeeperTimeout yaml.Duration `yaml:"keeper_timeout"`
	CheckInterval yaml.Duration `yaml:"check_interval"`
	SyncInterval  yaml.Duration `yaml:"sync_interval"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryBackoff  yaml.Duration `yaml:"retry_backoff"`
}

type NotificationConfig struct {
	Timeout       yaml.Duration `yaml:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts"`
	OfflineDelay  yaml.Duration `yaml:"offline_delay"`
}

type RPCConfig struct {
	GetPerformerTimeout yaml.Duration `yaml:"get_performer_timeout"`
	HealthTimeout       yaml.Duration `yaml:"health_timeout"`
}

type KeeperVersionsConfig struct {
	Latest                  []string `yaml:"latest"`
	UseTaskExecutionAddress []string `yaml:"use_task_execution_address"`
	UpgradeMessage          string   `yaml:"upgrade_message"`
}

type YAMLConfig struct {
	HealthCheck        HealthCheckConfig             `yaml:"health_check"`
	Notification       NotificationConfig            `yaml:"notification"`
	RPC                RPCConfig                     `yaml:"rpc"`
	DatabaseOperations yaml.DatabaseOperationsConfig `yaml:"database"`
	Metrics            yaml.MetricsConfig            `yaml:"metrics"`
	Shutdown           yaml.ShutdownConfig           `yaml:"shutdown"`
	Version            yaml.VersionConfig            `yaml:"version"`
	KeeperVersions     KeeperVersionsConfig          `yaml:"keeper_versions"`
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
		devMode:                  env.GetEnvBool("DEV_MODE", false),
		httpPort:                 env.GetEnvString("HEALTH_HTTP_PORT", "9004"),
		grpcPort:                 env.GetEnvString("HEALTH_GRPC_PORT", "9014"),
		dbConnection:             env.GetDatabaseConfig(),
		upstashRedisUrl:          env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken:    env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		otelExporterEndpoint:     env.GetOTELExporterEndpoint(),
		serviceID:                env.GetEnvString("HEALTH_SERVICE_ID", "1"),
		botToken:                 env.GetEnvString("BOT_TOKEN", ""),
		emailUser:                env.GetEnvString("EMAIL_USER", ""),
		emailPassword:            env.GetEnvString("EMAIL_PASS", ""),
		pinataHost:               env.GetEnvString("PINATA_HOST", ""),
		pinataJWT:                env.GetEnvString("PINATA_JWT", ""),
		dispatcherSigningAddress: env.GetEnvString("TASK_DISPATCHER_SIGNING_ADDRESS", ""),
		etherscanAPIKey:          env.GetEnvString("ETHERSCAN_API_KEY", ""),
		alchemyAPIKey:            env.GetEnvString("HEALTH_ALCHEMY_API_KEY", ""),
		taskExecutionAddress:     env.GetEnvString("TASK_EXECUTION_ADDRESS", ""),
		testTaskExecutionAddress: env.GetEnvString("TEST_TASK_EXECUTION_ADDRESS", ""),
		imuaTaskExecutionAddress: env.GetEnvString("IMUA_TASK_EXECUTION_ADDRESS", ""),
		healthCheck:              yamlConfig.HealthCheck,
		notification:             yamlConfig.Notification,
		rpc:                      yamlConfig.RPC,
		databaseOperations:       yamlConfig.DatabaseOperations,
		metrics:                  yamlConfig.Metrics,
		shutdown:                 yamlConfig.Shutdown,
		version:                  yamlConfig.Version,
		keeperVersions:           yamlConfig.KeeperVersions,
	}
	if err := validateConfig(); err != nil {
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
	if !env.IsValidIPAddress(cfg.dbConnection.HostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.dbConnection.HostAddress)
	}
	if !env.IsValidPort(cfg.dbConnection.HostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.dbConnection.HostPort)
	}
	if env.IsEmpty(cfg.upstashRedisUrl) {
		return fmt.Errorf("invalid upstash redis url: %s", cfg.upstashRedisUrl)
	}
	if env.IsEmpty(cfg.upstashRedisRestToken) {
		return fmt.Errorf("invalid upstash redis rest token: %s", cfg.upstashRedisRestToken)
	}
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("invalid Pinata Host: %s", cfg.pinataHost)
	}
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("invalid Pinata JWT: %s", cfg.pinataJWT)
	}
	if env.IsEmpty(cfg.etherscanAPIKey) {
		return fmt.Errorf("invalid Etherscan API Key: %s", cfg.etherscanAPIKey)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid Alchemy API Key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidEthAddress(cfg.taskExecutionAddress) {
		return fmt.Errorf("invalid task execution address: %s", cfg.taskExecutionAddress)
	}
	if !env.IsValidEthAddress(cfg.testTaskExecutionAddress) {
		return fmt.Errorf("invalid test task execution address: %s", cfg.testTaskExecutionAddress)
	}
	if !env.IsValidEthAddress(cfg.imuaTaskExecutionAddress) {
		return fmt.Errorf("invalid Imua task execution address: %s", cfg.imuaTaskExecutionAddress)
	}
	if !env.IsValidEthAddress(cfg.dispatcherSigningAddress) {
		return fmt.Errorf("invalid dispatcher signing address: %s", cfg.dispatcherSigningAddress)
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

func GetServiceID() string {
	return cfg.serviceID
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

func GetDatabaseTimeout() time.Duration {
	return cfg.databaseOperations.Timeout.ToDuration()
}

func GetDatabaseConnectWait() time.Duration {
	return cfg.databaseOperations.ConnectWait.ToDuration()
}

func GetDatabaseRetries() int {
	return cfg.databaseOperations.Retries
}

func GetMetricsUpdateInterval() time.Duration {
	return cfg.metrics.UpdateInterval.ToDuration()
}

func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}

func GetBotToken() string {
	return cfg.botToken
}

func GetEmailUser() string {
	return cfg.emailUser
}

func GetEmailPassword() string {
	return cfg.emailPassword
}

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func GetTaskExecutionAddress() string {
	return cfg.taskExecutionAddress
}

func GetTestTaskExecutionAddress() string {
	return cfg.testTaskExecutionAddress
}

func GetImuaTaskExecutionAddress() string {
	return cfg.imuaTaskExecutionAddress
}

func GetDispatcherSigningAddress() string {
	return cfg.dispatcherSigningAddress
}

func GetHealthCheckKeeperTimeout() time.Duration {
	return cfg.healthCheck.KeeperTimeout.ToDuration()
}

func GetHealthCheckInterval() time.Duration {
	return cfg.healthCheck.CheckInterval.ToDuration()
}

func GetHealthCheckSyncInterval() time.Duration {
	return cfg.healthCheck.SyncInterval.ToDuration()
}

func GetHealthCheckMaxRetries() int {
	return cfg.healthCheck.MaxRetries
}

func GetHealthCheckRetryBackoff() time.Duration {
	return cfg.healthCheck.RetryBackoff.ToDuration()
}

func GetNotificationTimeout() time.Duration {
	return cfg.notification.Timeout.ToDuration()
}

func GetNotificationRetryAttempts() int {
	return cfg.notification.RetryAttempts
}

func GetNotificationOfflineDelay() time.Duration {
	return cfg.notification.OfflineDelay.ToDuration()
}

func GetRPCGetPerformerTimeout() time.Duration {
	return cfg.rpc.GetPerformerTimeout.ToDuration()
}

func GetRPCHealthTimeout() time.Duration {
	return cfg.rpc.HealthTimeout.ToDuration()
}

// Keeper version configuration getters
func GetKeeperLatestVersions() []string {
	return cfg.keeperVersions.Latest
}

func GetKeeperVersionsWithTaskExecutionAddress() []string {
	return cfg.keeperVersions.UseTaskExecutionAddress
}

func GetKeeperUpgradeMessage() string {
	return cfg.keeperVersions.UpgradeMessage
}

// Helper function to check if a version is in a list
func IsKeeperVersionInList(version string, versionList []string) bool {
	for _, v := range versionList {
		if v == version {
			return true
		}
	}
	return false
}

// Redis configuration getters
func GetUpstashRedisUrl() string {
	return cfg.upstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.upstashRedisRestToken
}
