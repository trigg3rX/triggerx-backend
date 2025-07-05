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

	// Redis RPC port
	redisRPCPort string

	// DBServer RPC URL
	dbServerRPCUrl string
	// Health RPC URL
	healthRPCUrl string
	// Aggregator RPC URL
	aggregatorRPCUrl string

	// Redis signing key
	redisSigningKey     string
	redisSigningAddress string

	// Primary: Cloud Redis (Upstash) settings
	upstashURL   string
	upstashToken string

	// Pinata Host
	pinataHost string

	// Fallback: Local Redis settings (optional)
	localAddr     string
	localPassword string

	// Common settings
	db           int
	poolSize     int
	minIdleConns int
	maxRetries   int

	// Timeout settings
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	poolTimeout  time.Duration

	// Stream settings
	streamMaxLen    int
	jobStreamTTL    time.Duration
	taskStreamTTL   time.Duration
	cacheTTL        time.Duration
	cleanupInterval time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:             env.GetEnvBool("DEV_MODE", false),
		redisRPCPort:        env.GetEnv("REDIS_RPC_PORT", "9003"),
		healthRPCUrl:        env.GetEnv("HEALTH_RPC_URL", "http://localhost:9004"),
		dbServerRPCUrl:      env.GetEnv("DBSERVER_RPC_URL", "http://localhost:9002"),
		aggregatorRPCUrl:    env.GetEnv("AGGREGATOR_RPC_URL", "http://localhost:9001"),
		redisSigningKey:     env.GetEnv("REDIS_SIGNING_KEY", ""),
		redisSigningAddress: env.GetEnv("REDIS_SIGNING_ADDRESS", ""),
		upstashURL:          env.GetEnv("UPSTASH_REDIS_URL", ""),
		upstashToken:        env.GetEnv("UPSTASH_REDIS_REST_TOKEN", ""),
		localAddr:           env.GetEnv("REDIS_ADDR", "localhost:6379"),
		localPassword:       env.GetEnv("REDIS_PASSWORD", ""),
		db:                  0,
		poolSize:            env.GetEnvInt("REDIS_POOL_SIZE", 10),
		minIdleConns:        env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		maxRetries:          env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		dialTimeout:         env.GetEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		readTimeout:         env.GetEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		writeTimeout:        env.GetEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		poolTimeout:         env.GetEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
		streamMaxLen:        env.GetEnvInt("REDIS_STREAM_MAX_LEN", 10000),
		jobStreamTTL:        env.GetEnvDuration("REDIS_JOB_STREAM_TTL", 120*time.Hour),
		taskStreamTTL:       env.GetEnvDuration("REDIS_TASK_STREAM_TTL", 1*time.Hour),
		cacheTTL:            env.GetEnvDuration("REDIS_CACHE_TTL", 24*time.Hour),
		cleanupInterval:     env.GetEnvDuration("REDIS_CLEANUP_INTERVAL", 10*time.Minute),
		pinataHost:          env.GetEnv("PINATA_HOST", ""),
	}

	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func IsUpstashEnabled() bool {
	return cfg.upstashURL != ""
}

func IsLocalRedisEnabled() bool {
	return cfg.localAddr != ""
}

func IsRedisAvailable() bool {
	return IsUpstashEnabled() || IsLocalRedisEnabled()
}

func GetRedisType() string {
	if IsUpstashEnabled() {
		return "upstash"
	}
	if IsLocalRedisEnabled() {
		return "local"
	}
	return "none"
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func GetRedisRPCPort() string {
	return cfg.redisRPCPort
}

func GetDBServerRPCUrl() string {
	return cfg.dbServerRPCUrl
}

func GetAggregatorRPCUrl() string {
	return cfg.aggregatorRPCUrl
}

func GetRedisSigningKey() string {
	return cfg.redisSigningKey
}

func GetRedisSigningAddress() string {
	return cfg.redisSigningAddress
}

func GetUpstashURL() string {
	return cfg.upstashURL
}

func GetUpstashToken() string {
	return cfg.upstashToken
}

func GetRedisAddr() string {
	return cfg.localAddr
}

func GetRedisPassword() string {
	return cfg.localPassword
}

func GetRedisDB() int {
	return cfg.db
}

func GetStreamMaxLen() int {
	return cfg.streamMaxLen
}

func GetJobStreamTTL() time.Duration {
	return cfg.jobStreamTTL
}

func GetTaskStreamTTL() time.Duration {
	return cfg.taskStreamTTL
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

func GetCacheTTL() time.Duration {
	return cfg.cacheTTL
}

func GetCleanupInterval() time.Duration {
	return cfg.cleanupInterval
}

// GetRedisClientConfig returns a RedisConfig for the new Redis client
func GetRedisClientConfig() redisClient.RedisConfig {
	return redisClient.RedisConfig{
		IsUpstash: IsUpstashEnabled(),
		UpstashConfig: redisClient.UpstashConfig{
			URL:   cfg.upstashURL,
			Token: cfg.upstashToken,
		},
		LocalRedisConfig: redisClient.LocalRedisConfig{
			Addr:     cfg.localAddr,
			Password: cfg.localPassword,
			DB:       cfg.db,
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
		StreamsConfig: redisClient.StreamsConfig{
			JobStreamTTL:       cfg.jobStreamTTL,
			TaskStreamTTL:      cfg.taskStreamTTL,
			KeeperStreamTTL:    24 * time.Hour, // Default keeper stream TTL
			RegistrarStreamTTL: 24 * time.Hour, // Default registrar stream TTL
		},
	}
}
