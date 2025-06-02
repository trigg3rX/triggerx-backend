package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	DevMode bool

	// Primary: Cloud Redis (Upstash) settings
	UpstashURL   string
	UpstashToken string

	// Fallback: Local Redis settings (optional)
	LocalAddr     string
	LocalPassword string
	LocalEnabled  bool

	// Common settings
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int

	// Timeout settings
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration

	// Stream settings
	StreamMaxLen int
	StreamTTL    time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode: env.GetEnvBool("DEV_MODE", false),
		UpstashURL:   env.GetEnv("UPSTASH_REDIS_URL", ""),
		UpstashToken: env.GetEnv("UPSTASH_REDIS_REST_TOKEN", ""),
		LocalAddr:     env.GetEnv("REDIS_ADDR", "localhost:6379"),
		LocalPassword: env.GetEnv("REDIS_PASSWORD", ""),
		LocalEnabled:  env.GetEnvBool("REDIS_LOCAL_ENABLED", false),
		DB:           0,
		PoolSize:     env.GetEnvInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		MaxRetries:   env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		DialTimeout:  time.Duration(env.GetEnvInt("REDIS_DIAL_TIMEOUT_SEC", 5)) * time.Second,
		ReadTimeout:  time.Duration(env.GetEnvInt("REDIS_READ_TIMEOUT_SEC", 3)) * time.Second,
		WriteTimeout: time.Duration(env.GetEnvInt("REDIS_WRITE_TIMEOUT_SEC", 3)) * time.Second,
		PoolTimeout:  time.Duration(env.GetEnvInt("REDIS_POOL_TIMEOUT_SEC", 4)) * time.Second,
		StreamMaxLen: env.GetEnvInt("REDIS_STREAM_MAX_LEN", 10000),
		StreamTTL:    time.Duration(env.GetEnvInt("REDIS_STREAM_TTL_HOURS", 24)) * time.Hour,
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func IsDevMode() bool {
	return cfg.DevMode
}

// Primary Redis configuration (Cloud-first)
func IsUpstashEnabled() bool {
	return cfg.UpstashURL != ""
}

func GetUpstashURL() string {
	return cfg.UpstashURL
}

func GetUpstashToken() string {
	return cfg.UpstashToken
}

// Fallback Redis configuration (Local)
func IsLocalRedisEnabled() bool {
	return cfg.LocalEnabled && cfg.LocalAddr != ""
}

func GetRedisAddr() string {
	return cfg.LocalAddr
}

func GetRedisPassword() string {
	return cfg.LocalPassword
}

func GetRedisDB() int {
	return cfg.DB
}

// Redis availability check
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

// Common Redis settings
func GetStreamMaxLen() int {
	return cfg.StreamMaxLen
}

func GetStreamTTL() time.Duration {
	return cfg.StreamTTL
}

func GetPoolSize() int {
	return cfg.PoolSize
}

func GetMinIdleConns() int {
	return cfg.MinIdleConns
}

func GetMaxRetries() int {
	return cfg.MaxRetries
}

func GetDialTimeout() time.Duration {
	return cfg.DialTimeout
}

func GetReadTimeout() time.Duration {
	return cfg.ReadTimeout
}

func GetWriteTimeout() time.Duration {
	return cfg.WriteTimeout
}

func GetPoolTimeout() time.Duration {
	return cfg.PoolTimeout
}
