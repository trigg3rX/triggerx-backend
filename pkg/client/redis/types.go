package redis

import (
	"time"
)

type ConnectionSettings struct {
	PoolSize         int
	MaxIdleConns     int
	MinIdleConns     int
	MaxRetries       int
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	PoolTimeout      time.Duration
	PingTimeout      time.Duration
	HealthTimeout    time.Duration
	OperationTimeout time.Duration
}

type UpstashConfig struct {
	URL   string
	Token string
}

type RedisConfig struct {
	UpstashConfig      UpstashConfig
	ConnectionSettings ConnectionSettings
}

// RetryConfig defines configuration for retry mechanisms
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	LogRetryAttempt bool
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
	}
}

// HealthStatus represents the health status of Redis connection
type HealthStatus struct {
	Connected   bool
	LastPing    time.Time
	PingLatency time.Duration
	Errors      []string
	ServerInfo  map[string]interface{}
}

// ConnectionRecoveryConfig defines configuration for connection recovery
type ConnectionRecoveryConfig struct {
	Enabled         bool
	CheckInterval   time.Duration
	MaxRetries      int
	BackoffFactor   float64
	MaxBackoffDelay time.Duration
}

// DefaultConnectionRecoveryConfig returns default connection recovery configuration
func DefaultConnectionRecoveryConfig() *ConnectionRecoveryConfig {
	return &ConnectionRecoveryConfig{
		Enabled:         true,
		CheckInterval:   30 * time.Second,
		MaxRetries:      5,
		BackoffFactor:   2.0,
		MaxBackoffDelay: 5 * time.Minute,
	}
}
