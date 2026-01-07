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

	// Upstash Redis URL and Rest Token
	upstashRedisUrl       string
	upstashRedisRestToken string

	// Pinata JWT and Host
	pinataJWT  string
	pinataHost string

	// Notification webhook
	notifyWebhookURL   string
	notifyWebhookToken string

	// SMTP email settings
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPass     string
	smtpFrom     string
	smtpStartTLS bool

	// YAML-loaded settings
	redis   RedisConfig
	stream  StreamConfig
	metrics MetricsConfig
	version yaml.VersionConfig
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

type StreamConfig struct {
	MaxLen          int           `yaml:"max_len"`
	JobStreamTTL    yaml.Duration `yaml:"job_stream_ttl"`
	TaskStreamTTL   yaml.Duration `yaml:"task_stream_ttl"`
	CacheTTL        yaml.Duration `yaml:"cache_ttl"`
	CleanupInterval yaml.Duration `yaml:"cleanup_interval"`
}

type MetricsConfig struct {
	UpdateInterval yaml.Duration `yaml:"update_interval"`
}

type YAMLConfig struct {
	Redis   RedisConfig        `yaml:"redis"`
	Stream  StreamConfig       `yaml:"stream"`
	Metrics MetricsConfig      `yaml:"metrics"`
	Version yaml.VersionConfig `yaml:"version"`
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
		httpPort:              env.GetEnvString("TASK_MONITOR_HTTP_PORT", "9003"),
		grpcPort:              env.GetEnvString("TASK_MONITOR_GRPC_PORT", "9013"),
		dbConnection:          env.GetDatabaseConfig(),
		otelExporterEndpoint:  env.GetOTELExporterEndpoint(),
		serviceID:             env.GetEnvString("TASK_MONITOR_SERVICE_ID", "1"),
		upstashRedisUrl:       env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken: env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		pinataJWT:             env.GetEnvString("PINATA_JWT", ""),
		pinataHost:            env.GetEnvString("PINATA_HOST", ""),
		notifyWebhookURL:      env.GetEnvString("TASK_NOTIFY_WEBHOOK_URL", ""),
		notifyWebhookToken:    env.GetEnvString("TASK_NOTIFY_WEBHOOK_TOKEN", ""),
		smtpHost:              env.GetEnvString("SMTP_HOST", ""),
		smtpPort:              env.GetEnvInt("SMTP_PORT", 587),
		smtpUser:              env.GetEnvString("SMTP_USER", ""),
		smtpPass:              env.GetEnvString("SMTP_PASS", ""),
		smtpFrom:              env.GetEnvString("SMTP_FROM", ""),
		smtpStartTLS:          env.GetEnvBool("SMTP_STARTTLS", true),
		redis:                 yamlConfig.Redis,
		stream:                yamlConfig.Stream,
		metrics:               yamlConfig.Metrics,
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
		return fmt.Errorf("invalid Task Monitor HTTP Port: %s", cfg.httpPort)
	}
	if !env.IsValidPort(cfg.grpcPort) {
		return fmt.Errorf("invalid Task Monitor gRPC Port: %s", cfg.grpcPort)
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
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("invalid pinata JWT: %s", cfg.pinataJWT)
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("invalid pinata host: %s", cfg.pinataHost)
	}
	if env.IsEmpty(cfg.upstashRedisUrl) {
		return fmt.Errorf("invalid upstash redis url: %s", cfg.upstashRedisUrl)
	}
	if env.IsEmpty(cfg.upstashRedisRestToken) {
		return fmt.Errorf("invalid upstash redis rest token: %s", cfg.upstashRedisRestToken)
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

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetUpstashRedisUrl() string {
	return cfg.upstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.upstashRedisRestToken
}

func GetStreamMaxLen() int {
	return cfg.stream.MaxLen
}

func GetJobStreamTTL() time.Duration {
	return cfg.stream.JobStreamTTL.ToDuration()
}

func GetTaskStreamTTL() time.Duration {
	return cfg.stream.TaskStreamTTL.ToDuration()
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

func GetCacheTTL() time.Duration {
	return cfg.stream.CacheTTL.ToDuration()
}

func GetCleanupInterval() time.Duration {
	return cfg.stream.CleanupInterval.ToDuration()
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

func GetNotifyWebhookURL() string {
	return cfg.notifyWebhookURL
}

func GetNotifyWebhookToken() string {
	return cfg.notifyWebhookToken
}

func GetSMTPHost() string   { return cfg.smtpHost }
func GetSMTPPort() int      { return cfg.smtpPort }
func GetSMTPUser() string   { return cfg.smtpUser }
func GetSMTPPass() string   { return cfg.smtpPass }
func GetSMTPFrom() string   { return cfg.smtpFrom }
func GetSMTPStartTLS() bool { return cfg.smtpStartTLS }

// GetRedisClientConfig returns a RedisConfig for the new Redis client
func GetRedisClientConfig() redisClient.RedisConfig {
	return redisClient.RedisConfig{
		UpstashConfig: redisClient.UpstashConfig{
			URL:   cfg.upstashRedisUrl,
			Token: cfg.upstashRedisRestToken,
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
