package database

import (
	"github.com/gocql/gocql"
	// "log"
	// "time"
)

type Connection struct {
	session *gocql.Session
	config  *Config
}

func NewConnection(config *Config) (*Connection, error) {
	cluster := gocql.NewCluster(config.Hosts...)
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
	}

	return conn, nil
}

func (c *Connection) Session() *gocql.Session {
	return c.session
}

func (c *Connection) Close() {
	if c.session != nil {
		c.session.Close()
	}
} 