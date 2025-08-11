package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Task Dispatcher RPC port
	taskDispatcherRPCPort int

	// Health RPC URL
	healthRPCUrl string
	// Aggregator RPC URL
	aggregatorRPCUrl string

	// Task Dispatcher signing key
	signingKey     string
	signingAddress string

	// Redis (Upstash) connection settings
	upstashURL   string
	upstashToken string

	// OpenTelemetry endpoint
	ottempoEndpoint string

	// Common settings
	poolSize     int
	minIdleConns int
	maxRetries   int

	// Timeout settings
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	poolTimeout  time.Duration

	// Metrics settings
	metricsUpdateInterval time.Duration

	// Timeout and retry settings
	retryDelay            time.Duration
	requestTimeout        time.Duration
	initializationTimeout time.Duration
	maxRetryBackoff       time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:               env.GetEnvBool("DEV_MODE", false),
		taskDispatcherRPCPort: env.GetEnvInt("TASK_DISPATCHER_RPC_PORT", 9003),
		healthRPCUrl:          env.GetEnvString("HEALTH_RPC_URL", "http://localhost:9004"),
		aggregatorRPCUrl:      env.GetEnvString("AGGREGATOR_RPC_URL", "http://localhost:9001"),
		signingKey:            env.GetEnvString("TASK_DISPATCHER_SIGNING_KEY", ""),
		signingAddress:        env.GetEnvString("TASK_DISPATCHER_SIGNING_ADDRESS", ""),
		upstashURL:            env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashToken:          env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		poolSize:              env.GetEnvInt("REDIS_POOL_SIZE", 10),
		minIdleConns:          env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		maxRetries:            env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		dialTimeout:           env.GetEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		readTimeout:           env.GetEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		writeTimeout:          env.GetEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		poolTimeout:           env.GetEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
		metricsUpdateInterval: env.GetEnvDuration("REDIS_METRICS_UPDATE_INTERVAL", 30*time.Second),
		retryDelay:            env.GetEnvDuration("REDIS_RETRY_DELAY", 2*time.Second),
		requestTimeout:        env.GetEnvDuration("REDIS_REQUEST_TIMEOUT", 10*time.Second),
		initializationTimeout: env.GetEnvDuration("REDIS_INITIALIZATION_TIMEOUT", 10*time.Second),
		maxRetryBackoff:       env.GetEnvDuration("REDIS_MAX_RETRY_BACKOFF", 5*time.Minute),
		ottempoEndpoint:       env.GetEnvString("TEMPO_OTLP_ENDPOINT", "localhost:4318"),
	}

	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func GetTaskDispatcherRPCPort() int {
	return cfg.taskDispatcherRPCPort
}

func GetAggregatorRPCUrl() string {
	return cfg.aggregatorRPCUrl
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
	return cfg.poolSize
}

func GetMinIdleConns() int {
	return cfg.minIdleConns
}

func GetMaxRetries() int {
	return cfg.maxRetries
}

func GetDialTimeout() time.Duration {
	return cfg.dialTimeout
}

func GetReadTimeout() time.Duration {
	return cfg.readTimeout
}

func GetWriteTimeout() time.Duration {
	return cfg.writeTimeout
}

func GetPoolTimeout() time.Duration {
	return cfg.poolTimeout
}

func GetMetricsUpdateInterval() time.Duration {
	return cfg.metricsUpdateInterval
}

func GetRetryDelay() time.Duration {
	return cfg.retryDelay
}

func GetRequestTimeout() time.Duration {
	return cfg.requestTimeout
}

func GetInitializationTimeout() time.Duration {
	return cfg.initializationTimeout
}

func GetMaxRetryBackoff() time.Duration {
	return cfg.maxRetryBackoff
}

func GetOTTempoEndpoint() string {
	return cfg.ottempoEndpoint
}

// GetRedisClientConfig returns a RedisConfig for the new Redis client
func GetRedisClientConfig() redisClient.RedisConfig {
	return redisClient.RedisConfig{
		UpstashConfig: redisClient.UpstashConfig{
			URL:   cfg.upstashURL,
			Token: cfg.upstashToken,
		},
		ConnectionSettings: redisClient.ConnectionSettings{
			PoolSize:         cfg.poolSize,
			MaxIdleConns:     0, // Let Redis client manage this
			MinIdleConns:     cfg.minIdleConns,
			MaxRetries:       cfg.maxRetries,
			DialTimeout:      cfg.dialTimeout,
			ReadTimeout:      cfg.readTimeout,
			WriteTimeout:     cfg.writeTimeout,
			PoolTimeout:      cfg.poolTimeout,
			PingTimeout:      2 * time.Second,  // Default ping timeout
			HealthTimeout:    5 * time.Second,  // Default health check timeout
			OperationTimeout: 10 * time.Second, // Default operation timeout
		},
	}
}
