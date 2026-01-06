package database

import (
	"crypto/tls"
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

	// Authentication fields
	Username string
	Password string

	// SSL/TLS fields
	EnableSSL          bool
	SSLConfig          *tls.Config
	CertPath           string
	KeyPath            string
	CAPath             string
	InsecureSkipVerify bool
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

// WithAuthentication sets the username and password for authentication.
func (c *Config) WithAuthentication(username, password string) *Config {
	c.Username = username
	c.Password = password
	return c
}

// WithSSL enables SSL/TLS encryption with the provided configuration.
func (c *Config) WithSSL(sslConfig *tls.Config) *Config {
	c.EnableSSL = true
	c.SSLConfig = sslConfig
	return c
}

// WithSSLCertificates enables SSL/TLS using certificate files.
// If insecureSkipVerify is true, certificate verification is skipped (not recommended for production).
func (c *Config) WithSSLCertificates(certPath, keyPath, caPath string, insecureSkipVerify bool) *Config {
	c.EnableSSL = true
	c.CertPath = certPath
	c.KeyPath = keyPath
	c.CAPath = caPath
	c.InsecureSkipVerify = insecureSkipVerify
	return c
}
