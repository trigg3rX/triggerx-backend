package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Primary: Cloud Redis (Upstash) settings
	upstashURL   string
	upstashToken string

	// Fallback: Local Redis settings (optional)
	localAddr     string
	localPassword string
	localEnabled  bool

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

	aggregatorRPCURL string
	dispatcherPrivateKey string
	dispatcherAddress    string

	// Stream settings
	streamMaxLen int
	streamTTL    time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:       env.GetEnvBool("DEV_MODE", false),
		upstashURL:    env.GetEnv("UPSTASH_REDIS_URL", ""),
		upstashToken:  env.GetEnv("UPSTASH_REDIS_REST_TOKEN", ""),
		localAddr:     env.GetEnv("REDIS_ADDR", "localhost:6379"),
		localPassword: env.GetEnv("REDIS_PASSWORD", ""),
		localEnabled:  env.GetEnvBool("REDIS_LOCAL_ENABLED", false),
		db:            0,
		poolSize:      env.GetEnvInt("REDIS_POOL_SIZE", 10),
		minIdleConns:  env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		maxRetries:    env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		dialTimeout:   env.GetEnvDuration("REDIS_DIAL_TIMEOUT_SEC", 5*time.Second),
		readTimeout:   env.GetEnvDuration("REDIS_READ_TIMEOUT_SEC", 3*time.Second),
		writeTimeout:  env.GetEnvDuration("REDIS_WRITE_TIMEOUT_SEC", 3*time.Second),
		poolTimeout:   env.GetEnvDuration("REDIS_POOL_TIMEOUT_SEC", 4*time.Second),
		aggregatorRPCURL: env.GetEnv("AGGREGATOR_RPC_URL", ""),
		dispatcherPrivateKey: env.GetEnv("DISPATCHER_PRIVATE_KEY", ""),
		dispatcherAddress:    env.GetEnv("DISPATCHER_ADDRESS", ""),
		streamMaxLen:  env.GetEnvInt("REDIS_STREAM_MAX_LEN", 10000),
		streamTTL:     env.GetEnvDuration("REDIS_STREAM_TTL_HOURS", 24*time.Hour),
	}
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func validateConfig() error {
	if env.IsEmpty(cfg.upstashURL) {
		return fmt.Errorf("invalid upstash url: %s", cfg.upstashURL)
	}
	if env.IsEmpty(cfg.upstashToken) {
		return fmt.Errorf("invalid upstash token: %s", cfg.upstashToken)
	}
	if !env.IsValidURL(cfg.aggregatorRPCURL) {
		return fmt.Errorf("invalid aggregator rpc url: %s", cfg.aggregatorRPCURL)
	}
	if !env.IsValidEthAddress(cfg.dispatcherAddress) {
		return fmt.Errorf("invalid dispatcher address: %s", cfg.dispatcherAddress)
	}
	if !env.IsValidPrivateKey(cfg.dispatcherPrivateKey) {
		return fmt.Errorf("invalid dispatcher private key: %s", cfg.dispatcherPrivateKey)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func IsUpstashEnabled() bool {
	return cfg.upstashURL != ""
}

func GetUpstashURL() string {
	return cfg.upstashURL
}

func GetUpstashToken() string {
	return cfg.upstashToken
}

func IsLocalRedisEnabled() bool {
	return cfg.localEnabled && cfg.localAddr != ""
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

func GetStreamMaxLen() int {
	return cfg.streamMaxLen
}

func GetStreamTTL() time.Duration {
	return cfg.streamTTL
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

func GetAggregatorRPCURL() string {
	return cfg.aggregatorRPCURL
}

func GetDispatcherPrivateKey() string {
	return cfg.dispatcherPrivateKey
}

func GetDispatcherAddress() string {
	return cfg.dispatcherAddress
}