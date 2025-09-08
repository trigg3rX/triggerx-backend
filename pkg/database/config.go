package database

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// Config holds the configuration for the ScyllaDB connection.
type Config struct {
	Hosts              []string
	Keyspace           string
	Timeout            time.Duration
	Retries            int
	ConnectWait        time.Duration
	Consistency        gocql.Consistency
	ProtoVersion       int
	SocketKeepalive    time.Duration
	MaxPreparedStmts   int
	DefaultIdempotence bool
	RetryConfig        *retry.RetryConfig
}

func NewConfig(DatabaseHost string, DatabaseHostPort string) *Config {
	return &Config{
		Hosts:              []string{DatabaseHost + ":" + DatabaseHostPort},
		Keyspace:           "triggerx",
		Timeout:            time.Second * 30,
		Retries:            5,
		ConnectWait:        time.Second * 10,
		Consistency:        gocql.Quorum,
		ProtoVersion:       4,
		SocketKeepalive:    15 * time.Second,
		MaxPreparedStmts:   1000,
		DefaultIdempotence: true,
		RetryConfig:        retry.DefaultRetryConfig(),
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
