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

	// Scheduler RPC URLs
	timeSchedulerRPCUrl      string
	conditionSchedulerRPCUrl string

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

	// Task Execution Address
	taskExecutionAddress     string
	testTaskExecutionAddress string
	imuaTaskExecutionAddress string

	// YAML-loaded settings
	databaseOperations   yaml.DatabaseOperationsConfig
	metrics              yaml.MetricsConfig
	shutdown             yaml.ShutdownConfig
	version              yaml.VersionConfig
}

type YAMLConfig struct {
	DatabaseOperations yaml.DatabaseOperationsConfig `yaml:"database"`
	Metrics              yaml.MetricsConfig          `yaml:"metrics"`
	Shutdown             yaml.ShutdownConfig         `yaml:"shutdown"`
	Version              yaml.VersionConfig          `yaml:"version"`
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
		httpPort:                      env.GetEnvString("DBSERVER_HTTP_PORT", "9002"),
		grpcPort:                      env.GetEnvString("DBSERVER_GRPC_PORT", "9012"),
		dbConnection:                  env.GetDatabaseConfig(),
		otelExporterEndpoint:          env.GetOTELExporterEndpoint(),
		timeSchedulerRPCUrl:           env.GetEnvString("TIME_SCHEDULER_RPC_URL", "localhost:9015"),
		conditionSchedulerRPCUrl:      env.GetEnvString("CONDITION_SCHEDULER_RPC_URL", "localhost:9016"),
		emailUser:                     env.GetEnvString("EMAIL_USER", ""),
		emailPassword:                 env.GetEnvString("EMAIL_PASS", ""),
		botToken:                      env.GetEnvString("BOT_TOKEN", ""),
		alchemyAPIKey:                 env.GetEnvString("DBSERVER_ALCHEMY_API_KEY", ""),
		faucetPrivateKey:              env.GetEnvString("FAUCET_PRIVATE_KEY", ""),
		faucetFundAmount:              env.GetEnvString("FAUCET_FUND_AMOUNT", "30000000000000000"),
		upstashRedisUrl:               env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken:         env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		taskExecutionAddress:          env.GetEnvString("TASK_EXECUTION_ADDRESS", ""),
		testTaskExecutionAddress:      env.GetEnvString("TEST_TASK_EXECUTION_ADDRESS", ""),
		imuaTaskExecutionAddress:      env.GetEnvString("IMUA_TASK_EXECUTION_ADDRESS", ""),
		databaseOperations:            yamlConfig.DatabaseOperations,
		metrics:                       yamlConfig.Metrics,
		shutdown:                      yamlConfig.Shutdown,
		version:                       yamlConfig.Version,
	}
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if err := yaml.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func validateConfig(cfg Config) error {
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
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	if !env.IsValidHostPort(cfg.timeSchedulerRPCUrl) {
		return fmt.Errorf("invalid time scheduler RPC URL (expected host:port): %s", cfg.timeSchedulerRPCUrl)
	}
	if !env.IsValidHostPort(cfg.conditionSchedulerRPCUrl) {
		return fmt.Errorf("invalid condition scheduler RPC URL (expected host:port): %s", cfg.conditionSchedulerRPCUrl)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidPrivateKey(cfg.faucetPrivateKey) {
		return fmt.Errorf("invalid faucet private key: %s", cfg.faucetPrivateKey)
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

func GetTimeSchedulerRPCUrl() string {
	return cfg.timeSchedulerRPCUrl
}

func GetConditionSchedulerRPCUrl() string {
	return cfg.conditionSchedulerRPCUrl
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

func GetTaskExecutionAddress() string {
	return cfg.taskExecutionAddress
}

func GetTestTaskExecutionAddress() string {
	return cfg.testTaskExecutionAddress
}

func GetImuaTaskExecutionAddress() string {
	return cfg.imuaTaskExecutionAddress
}
