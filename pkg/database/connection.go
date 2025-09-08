package database

import (
	"context"
	"sync"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Connection holds the database session and configuration.
type Connection struct {
	session Sessioner
	config  *Config
	logger  logging.Logger
}

var (
	once     sync.Once
	instance *Connection
)

// NewConnection creates a new ScyllaDB connection.
// It uses a singleton pattern to ensure only one connection is created.
func NewConnection(config *Config, logger logging.Logger) (*Connection, error) {
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
func (c *Connection) SetLogger(logger logging.Logger) {
	c.logger = logger
}
