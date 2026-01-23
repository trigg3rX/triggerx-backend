package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"

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

	// Service ID for OpenTelemetry
	serviceID string

	// Health RPC URL
	healthRPCUrl string
	// Aggregator RPC URL
	aggregatorRPCUrl     string
	testAggregatorRPCUrl string

	// Performer API URLs (for direct HTTP calls to keeper)
	performerAPIUrl     string
	testPerformerAPIUrl string

	// Task Dispatcher signing key
	signingKey     string
	signingAddress string

	// Redis (Upstash) connection settings
	upstashURL   string
	upstashToken string

	// YAML-loaded settings
	redis    RedisConfig
	databaseOperations yaml.DatabaseOperationsConfig
	metrics              yaml.MetricsConfig
	shutdown             yaml.ShutdownConfig
	version              yaml.VersionConfig
}

type RedisConfig struct {
	PoolSize              int           `yaml:"pool_size"`
	MinIdleConns          int           `yaml:"min_idle_conns"`
	MaxRetries            int           `yaml:"max_retries"`
	DialTimeout           yaml.Duration `yaml:"dial_timeout"`
	ReadTimeout           yaml.Duration `yaml:"read_timeout"`
	WriteTimeout          yaml.Duration `yaml:"write_timeout"`
	PoolTimeout           yaml.Duration `yaml:"pool_timeout"`
	RetryDelay            yaml.Duration `yaml:"retry_delay"`
	RequestTimeout        yaml.Duration `yaml:"request_timeout"`
	InitializationTimeout yaml.Duration `yaml:"initialization_timeout"`
	MaxRetryBackoff       yaml.Duration `yaml:"max_retry_backoff"`
}

type YAMLConfig struct {
	Redis    RedisConfig    `yaml:"redis"`
	DatabaseOperations yaml.DatabaseOperationsConfig `yaml:"database"`
	Metrics  yaml.MetricsConfig  `yaml:"metrics"`
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
		devMode:               env.GetEnvBool("DEV_MODE", false),
		httpPort:              env.GetEnvString("TASK_DISPATCHER_HTTP_PORT", "9007"),
		grpcPort:              env.GetEnvString("TASK_DISPATCHER_GRPC_PORT", "9017"),
		dbConnection:          env.GetDatabaseConfig(),
		otelExporterEndpoint:  env.GetOTELExporterEndpoint(),
		serviceID:             env.GetEnvString("TASK_DISPATCHER_SERVICE_ID", "1"),
		healthRPCUrl:          env.GetEnvString("HEALTH_RPC_URL", "localhost:9014"),
		aggregatorRPCUrl:      env.GetEnvString("AGGREGATOR_RPC_URL", "localhost:9001"),
		testAggregatorRPCUrl:  env.GetEnvString("TEST_AGGREGATOR_RPC_URL", "localhost:9001"),
		performerAPIUrl:       env.GetEnvString("PERFORMER_API_URL", "localhost:9021"),
		testPerformerAPIUrl:   env.GetEnvString("TEST_PERFORMER_API_URL", "localhost:9021"),
		signingKey:            env.GetEnvString("TASK_DISPATCHER_SIGNING_KEY", ""),
		signingAddress:        env.GetEnvString("TASK_DISPATCHER_SIGNING_ADDRESS", ""),
		upstashURL:            env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashToken:          env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		redis:                 yamlConfig.Redis,
		databaseOperations:    yamlConfig.DatabaseOperations,
		metrics:               yamlConfig.Metrics,
		shutdown:              yamlConfig.Shutdown,
		version:               yamlConfig.Version,
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
		return fmt.Errorf("invalid Task Dispatcher HTTP Port: %s", cfg.httpPort)
	}
	if !env.IsValidPort(cfg.grpcPort) {
		return fmt.Errorf("invalid Task Dispatcher gRPC Port: %s", cfg.grpcPort)
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
	if !env.IsValidHostPort(cfg.healthRPCUrl) {
		return fmt.Errorf("invalid health RPC URL: %s", cfg.healthRPCUrl)
	}
	if !env.IsValidHostPort(cfg.aggregatorRPCUrl) {
		return fmt.Errorf("invalid aggregator RPC URL: %s", cfg.aggregatorRPCUrl)
	}
	if !env.IsValidHostPort(cfg.testAggregatorRPCUrl) {
		return fmt.Errorf("invalid test aggregator RPC URL: %s", cfg.testAggregatorRPCUrl)
	}
	if !env.IsValidURL(cfg.performerAPIUrl) {
		return fmt.Errorf("invalid performer API URL: %s", cfg.performerAPIUrl)
	}
	if !env.IsValidURL(cfg.testPerformerAPIUrl) {
		return fmt.Errorf("invalid test performer API URL: %s", cfg.testPerformerAPIUrl)
	}
	if !env.IsValidPrivateKey(cfg.signingKey) {
		return fmt.Errorf("invalid signing key: %s", cfg.signingKey)
	}
	if !env.IsValidEthAddress(cfg.signingAddress) {
		return fmt.Errorf("invalid signing address: %s", cfg.signingAddress)
	}
	if env.IsEmpty(cfg.upstashURL) {
		return fmt.Errorf("invalid upstash URL: %s", cfg.upstashURL)
	}
	if env.IsEmpty(cfg.upstashToken) {
		return fmt.Errorf("invalid upstash token: %s", cfg.upstashToken)
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

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetServiceID() string {
	return cfg.serviceID
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func GetAggregatorRPCUrl() string {
	url := cfg.aggregatorRPCUrl
	// Auto-prepend http:// if scheme is missing (for backward compatibility)
	if url != "" && !hasScheme(url) {
		return "http://" + url
	}
	return url
}

func GetTestAggregatorRPCUrl() string {
	url := cfg.testAggregatorRPCUrl
	// Auto-prepend http:// if scheme is missing (for backward compatibility)
	if url != "" && !hasScheme(url) {
		return "http://" + url
	}
	return url
}

// hasScheme checks if a URL string has a scheme (http://, https://, etc.)
func hasScheme(url string) bool {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' {
			// Check if it's followed by // (scheme separator)
			if i+2 < len(url) && url[i+1] == '/' && url[i+2] == '/' {
				return true
			}
			// If we hit a colon before //, it's likely a port, not a scheme
			return false
		}
		if url[i] == '/' {
			// If we hit / before :, no scheme
			return false
		}
	}
	return false
}

func GetPerformerAPIUrl() string {
	return cfg.performerAPIUrl
}

func GetTestPerformerAPIUrl() string {
	return cfg.testPerformerAPIUrl
}

func GetTaskDispatcherSigningKey() string {
	return cfg.signingKey
}

func GetTaskDispatcherSigningAddress() string {
	return cfg.signingAddress
}

func GetUpstashURL() string {
	return cfg.upstashURL
}

func GetUpstashToken() string {
	return cfg.upstashToken
}

func GetPoolSize() int {
	return cfg.redis.PoolSize
}

func GetMinIdleConns() int {
	return cfg.redis.MinIdleConns
}

func GetMaxRetries() int {
	return cfg.redis.MaxRetries
}

func GetDialTimeout() time.Duration {
	return cfg.redis.DialTimeout.ToDuration()
}

func GetReadTimeout() time.Duration {
	return cfg.redis.ReadTimeout.ToDuration()
}

func GetWriteTimeout() time.Duration {
	return cfg.redis.WriteTimeout.ToDuration()
}

func GetPoolTimeout() time.Duration {
	return cfg.redis.PoolTimeout.ToDuration()
}

func GetMetricsUpdateInterval() time.Duration {
	return cfg.metrics.UpdateInterval.ToDuration()
}

func GetRetryDelay() time.Duration {
	return cfg.redis.RetryDelay.ToDuration()
}

func GetRequestTimeout() time.Duration {
	return cfg.redis.RequestTimeout.ToDuration()
}

func GetInitializationTimeout() time.Duration {
	return cfg.redis.InitializationTimeout.ToDuration()
}

func GetMaxRetryBackoff() time.Duration {
	return cfg.redis.MaxRetryBackoff.ToDuration()
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



func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}

// GetRedisClientConfig returns a RedisConfig for the new Redis client
func GetRedisClientConfig() redisClient.RedisConfig {
	return redisClient.RedisConfig{
		UpstashConfig: redisClient.UpstashConfig{
			URL:   cfg.upstashURL,
			Token: cfg.upstashToken,
		},
		ConnectionSettings: redisClient.ConnectionSettings{
			PoolSize:         cfg.redis.PoolSize,
			MaxIdleConns:     0, // Let Redis client manage this
			MinIdleConns:     cfg.redis.MinIdleConns,
			MaxRetries:       cfg.redis.MaxRetries,
			DialTimeout:      cfg.redis.DialTimeout.ToDuration(),
			ReadTimeout:      cfg.redis.ReadTimeout.ToDuration(),
			WriteTimeout:     cfg.redis.WriteTimeout.ToDuration(),
			PoolTimeout:      cfg.redis.PoolTimeout.ToDuration(),
			PingTimeout:      2 * time.Second,  // Default ping timeout
			HealthTimeout:    5 * time.Second,  // Default health check timeout
			OperationTimeout: 10 * time.Second, // Default operation timeout
		},
	}
}
