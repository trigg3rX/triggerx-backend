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
	streamMaxLen  int
	jobStreamTTL  time.Duration
	taskStreamTTL time.Duration

	// API settings
	apiPort string
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
		db:            0,
		poolSize:      env.GetEnvInt("REDIS_POOL_SIZE", 10),
		minIdleConns:  env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		maxRetries:    env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		dialTimeout:   env.GetEnvDuration("REDIS_DIAL_TIMEOUT_SEC", 5*time.Second),
		readTimeout:   env.GetEnvDuration("REDIS_READ_TIMEOUT_SEC", 3*time.Second),
		writeTimeout:  env.GetEnvDuration("REDIS_WRITE_TIMEOUT_SEC", 3*time.Second),
		poolTimeout:   env.GetEnvDuration("REDIS_POOL_TIMEOUT_SEC", 4*time.Second),
		streamMaxLen:  env.GetEnvInt("REDIS_STREAM_MAX_LEN", 10000),
		jobStreamTTL:  env.GetEnvDuration("REDIS_JOB_STREAM_TTL_HOURS", 120*time.Hour),
		taskStreamTTL: env.GetEnvDuration("REDIS_TASK_STREAM_TTL_HOURS", 1*time.Hour),
		apiPort:       env.GetEnv("REDIS_API_PORT", "9009"),
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

func GetRedisAPIPort() string {
	return cfg.apiPort
}
