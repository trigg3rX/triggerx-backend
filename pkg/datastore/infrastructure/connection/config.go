package connection

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// Config holds the configuration for the ScyllaDB connection.
type Config struct {
	Hosts               []string
	Keyspace            string
	Timeout             time.Duration
	Retries             int
	ConnectWait         time.Duration
	Consistency         gocql.Consistency
	HealthCheckInterval time.Duration
	ProtoVersion        int
	SocketKeepalive     time.Duration
	MaxPreparedStmts    int
	DefaultIdempotence  bool
	RetryConfig         *retry.RetryConfig
}

// NewConfig creates a new Config with sensible defaults.
func NewConfig(DatabaseHost string, DatabaseHostPort string) *Config {
	return &Config{
		Hosts:               []string{DatabaseHost + ":" + DatabaseHostPort},
		Keyspace:            "triggerx",
		Timeout:             time.Second * 30,
		Retries:             5,
		ConnectWait:         time.Second * 10,
		Consistency:         gocql.Quorum,
		HealthCheckInterval: time.Second * 15,
		ProtoVersion:        4,
		SocketKeepalive:     15 * time.Second,
		MaxPreparedStmts:    1000,
		DefaultIdempotence:  true,
		RetryConfig:         retry.DefaultRetryConfig(),
	}
}

// WithHosts sets the hosts for the database connection.
func (c *Config) WithHosts(hosts []string) *Config {
	c.Hosts = hosts
	return c
}

// WithKeyspace sets the keyspace for the database connection.
func (c *Config) WithKeyspace(keyspace string) *Config {
	c.Keyspace = keyspace
	return c
}

// WithTimeout sets the timeout for queries.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithRetries sets the number of retries for a query.
func (c *Config) WithRetries(retries int) *Config {
	c.Retries = retries
	return c
}

// WithRetryConfig sets the retry configuration.
func (c *Config) WithRetryConfig(retryConfig *retry.RetryConfig) *Config {
	c.RetryConfig = retryConfig
	return c
}

// WithHealthCheckInterval sets the health check interval.
func (c *Config) WithHealthCheckInterval(interval time.Duration) *Config {
	c.HealthCheckInterval = interval
	return c
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	if len(c.Hosts) == 0 {
		return fmt.Errorf("at least one host must be specified")
	}

	if c.Keyspace == "" {
		return fmt.Errorf("keyspace cannot be empty")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", c.Timeout)
	}

	if c.ConnectWait <= 0 {
		return fmt.Errorf("connect wait must be positive, got: %v", c.ConnectWait)
	}

	if c.Retries < 0 {
		return fmt.Errorf("retries cannot be negative, got: %d", c.Retries)
	}

	if c.ProtoVersion < 1 || c.ProtoVersion > 4 {
		return fmt.Errorf("protocol version must be between 1 and 4, got: %d", c.ProtoVersion)
	}

	if c.MaxPreparedStmts < 0 {
		return fmt.Errorf("max prepared statements cannot be negative, got: %d", c.MaxPreparedStmts)
	}

	if c.HealthCheckInterval < 0 {
		return fmt.Errorf("health check interval cannot be negative, got: %v", c.HealthCheckInterval)
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	hosts := make([]string, len(c.Hosts))
	copy(hosts, c.Hosts)

	return &Config{
		Hosts:               hosts,
		Keyspace:            c.Keyspace,
		Timeout:             c.Timeout,
		Retries:             c.Retries,
		ConnectWait:         c.ConnectWait,
		Consistency:         c.Consistency,
		HealthCheckInterval: c.HealthCheckInterval,
		ProtoVersion:        c.ProtoVersion,
		SocketKeepalive:     c.SocketKeepalive,
		MaxPreparedStmts:    c.MaxPreparedStmts,
		DefaultIdempotence:  c.DefaultIdempotence,
		RetryConfig:         c.RetryConfig,
	}
}
