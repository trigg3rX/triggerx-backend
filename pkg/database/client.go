package database

import (
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type Connection struct {
	session Sessioner
	config  *Config
	logger  logging.Logger
}

func NewConnection(config *Config, logger logging.Logger) (*Connection, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
	cluster.Timeout = config.Timeout
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: config.Retries}
	cluster.ConnectTimeout = config.ConnectWait
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		session: session,
		config:  config,
		logger:  logger,
	}

	return conn, nil
}

func (c *Connection) Session() Sessioner {
	return c.session
}

func (c *Connection) Close() {
	if c.session != nil {
		c.session.Close()
	}
}
