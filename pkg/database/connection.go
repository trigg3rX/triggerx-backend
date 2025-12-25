package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"sync"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Connection holds the database session and configuration.
type Connection struct {
	session Sessioner
	config  *Config
	logger  observability.Logger
}

var (
	once     sync.Once
	instance *Connection
)

// NewConnection creates a new ScyllaDB connection.
// It uses a singleton pattern to ensure only one connection is created.
func NewConnection(config *Config, logger observability.Logger) (*Connection, error) {
	var err error
	once.Do(func() {
		cluster := gocql.NewCluster(config.Hosts...)
		cluster.Keyspace = config.Keyspace
		cluster.Timeout = config.Timeout
		cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: config.Retries}
		cluster.ConnectTimeout = config.ConnectWait
		cluster.Consistency = config.Consistency
		cluster.ProtoVersion = config.ProtoVersion
		cluster.SocketKeepalive = config.SocketKeepalive
		cluster.MaxPreparedStmts = config.MaxPreparedStmts
		cluster.DefaultIdempotence = config.DefaultIdempotence

		// Configure authentication if provided
		if config.Username != "" && config.Password != "" {
			cluster.Authenticator = gocql.PasswordAuthenticator{
				Username: config.Username,
				Password: config.Password,
			}
			logger.Info(context.Background(), "ScyllaDB authentication enabled",
				observability.String("username", config.Username))
		}

		// Configure SSL/TLS if enabled
		if config.EnableSSL {
			var tlsConfig *tls.Config

			if config.SSLConfig != nil {
				// Use provided TLS config
				tlsConfig = config.SSLConfig
			} else if config.CertPath != "" && config.KeyPath != "" {
				// Load certificates from files
				tlsConfig, err = loadTLSConfigFromFiles(config.CertPath, config.KeyPath, config.CAPath, config.InsecureSkipVerify)
				if err != nil {
					logger.Error(context.Background(), "Failed to load SSL certificates", observability.Error(err))
					return
				}
			} else {
				// Basic TLS config without client certificates
				tlsConfig = &tls.Config{
					InsecureSkipVerify: config.InsecureSkipVerify,
				}
			}

			cluster.SslOpts = &gocql.SslOptions{
				Config: tlsConfig,
			}
			logger.Info(context.Background(), "ScyllaDB SSL/TLS encryption enabled")
		}

		session, sessionErr := cluster.CreateSession()
		if sessionErr != nil {
			err = sessionErr
			return
		}

		instance = &Connection{
			session: session,
			config:  config,
			logger:  logger,
		}
	})

	return instance, err
}

// loadTLSConfigFromFiles loads TLS configuration from certificate files.
func loadTLSConfigFromFiles(certPath, keyPath, caPath string, insecureSkipVerify bool) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Load client certificate and key if provided
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if caPath != "" {
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}
		config.RootCAs = caCertPool
	}

	return config, nil
}

// Session returns the underlying gocql session.

func (c *Connection) Session() Sessioner {
	return c.session
}

// Session returns the underlying gocql session.
func (c *Connection) Close() {
	if c.session != nil {
		c.session.Close()
	}
}

// HealthCheck performs a simple query to check the database connection.
func (c *Connection) HealthCheck(ctx context.Context) error {
	return c.session.Query("SELECT release_version FROM system.local").WithContext(ctx).Exec()
}

// SetSession sets the session for the connection.
func (c *Connection) SetSession(session Sessioner) {
	c.session = session
}

// SetConfig sets the config for the connection.
func (c *Connection) SetConfig(config *Config) {
	c.config = config
}

// SetLogger sets the logger for the connection.
func (c *Connection) SetLogger(logger observability.Logger) {
	c.logger = logger
}
