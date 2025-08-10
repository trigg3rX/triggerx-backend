package redis

import (
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// PipelineFunc is a function that accepts a Redis pipeline object.
// Users will queue their commands (e.g., pipe.Get, pipe.Set) inside this function.
type PipelineFunc func(pipe redis.Pipeliner) error

type ConnectionSettings struct {
	PoolSize     int
	MaxIdleConns int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration // Used for socket reads, including for PING. This is the preferred way to set command timeouts.
	WriteTimeout time.Duration // Used for socket writes.
	PoolTimeout  time.Duration

	// The fields below are now DEPRECATED for creating timeouts, as using Read/Write timeouts
	// is better for connection pool health. They are kept for reference but are no longer
	// used to create a context.WithTimeout in the Ping or CheckConnection methods.
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

// RetryConfig is an alias for the generic retry configuration
type RetryConfig = retry.RetryConfig

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	config := retry.DefaultRetryConfig()
	// Override with Redis-specific defaults
	config.MaxRetries = 3
	config.InitialDelay = 100 * time.Millisecond
	config.MaxDelay = 5 * time.Second
	config.JitterFactor = 0.1
	return config
}

// ServerInfo holds parsed information from the Redis server.
type ServerInfo struct {
	RawInfo         string    `json:"raw_info"`
	Timestamp       time.Time `json:"timestamp"`
	MemoryInfo      string    `json:"memory_info,omitempty"`
	StatsInfo       string    `json:"stats_info,omitempty"`
	MemoryInfoError string    `json:"memory_info_error,omitempty"`
	StatsInfoError  string    `json:"stats_info_error,omitempty"`
}

// HealthStatus represents the health status of Redis connection
type HealthStatus struct {
	Connected   bool
	LastPing    time.Time
	PingLatency time.Duration
	Errors      []string
	ServerInfo  *ServerInfo
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

// PoolHealthStats holds the statistics of the connection pool.
type PoolHealthStats struct {
	Hits       uint32 `json:"hits"`
	Misses     uint32 `json:"misses"`
	Timeouts   uint32 `json:"timeouts"`
	TotalConns uint32 `json:"total_conns"`
	IdleConns  uint32 `json:"idle_conns"`
	StaleConns uint32 `json:"stale_conns"`
}

// ConnectionStatus represents the current state of the client's connection.
type ConnectionStatus struct {
	IsRecovering     bool            `json:"is_recovering"`
	LastHealthCheck  time.Time       `json:"last_health_check"`
	RecoveryEnabled  bool            `json:"recovery_enabled"`
	RecoveryInterval time.Duration   `json:"recovery_interval"`
	PoolStats        PoolHealthStats `json:"pool_stats"`
}

// HealthCheckResult represents the detailed outcome of a full health check.
type HealthCheckResult struct {
	Ping      PingCheckResult      `json:"ping"`
	Set       OperationCheckResult `json:"set"`
	Get       GetCheckResult       `json:"get"`
	Cleanup   CleanupCheckResult   `json:"cleanup,omitempty"`
	PoolStats PoolHealthStats      `json:"pool_stats"`
	Overall   OverallHealth        `json:"overall"`
}

// PingCheckResult holds the result of a ping test.
type PingCheckResult struct {
	Success bool          `json:"success"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

// OperationCheckResult holds the result of a generic operation test (like SET).
type OperationCheckResult struct {
	Success bool          `json:"success"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

// GetCheckResult holds the result of a GET test.
type GetCheckResult struct {
	Success    bool          `json:"success"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
	ValueMatch bool          `json:"value_match"`
}

// CleanupCheckResult holds the result of the cleanup operation.
type CleanupCheckResult struct {
	Error string `json:"error,omitempty"`
}

// OverallHealth holds the summary of the health check.
type OverallHealth struct {
	Healthy   bool      `json:"healthy"`
	Timestamp time.Time `json:"timestamp"`
}

// ScanOptions defines options for the Scan operation
type ScanOptions struct {
	Pattern string `json:"pattern"` // Redis pattern for key matching (e.g., "user:*", "*:active")
	Count   int64  `json:"count"`   // Number of keys to scan per iteration (suggested: 100-1000)
	Type    string `json:"type"`    // Optional: filter by type (string, list, set, zset, hash, stream)
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	Cursor  uint64   `json:"cursor"`   // Next cursor position for continuation
	Keys    []string `json:"keys"`     // Keys found in this iteration
	HasMore bool     `json:"has_more"` // Whether there are more keys to scan
}
