package database

import (
	"time"

	"github.com/gocql/gocql"
)

type Connection struct {
	session *gocql.Session
	config  *Config
}

func NewConnection(config *Config) (*Connection, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Timeout = config.Timeout
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{
		Min:        config.ConnectWait,
		Max:        config.ConnectWait * 10,
		NumRetries: config.Retries,
	}
	cluster.ConnectTimeout = config.ConnectWait
	cluster.ReconnectionPolicy = &gocql.ConstantReconnectionPolicy{
		MaxRetries: 10,
		Interval:   config.ConnectWait,
	}

	// Set consistency level based on config
	consistencyLevel := gocql.One // default to ONE for better availability
	if config.Consistency != "" {
		switch config.Consistency {
		case "ONE":
			consistencyLevel = gocql.One
		case "QUORUM":
			consistencyLevel = gocql.Quorum
		case "LOCAL_QUORUM":
			consistencyLevel = gocql.LocalQuorum
		case "LOCAL_ONE":
			consistencyLevel = gocql.LocalOne
		}
	}
	cluster.Consistency = consistencyLevel

	// Enable automatic host discovery and failover
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.Keyspace = "triggerx"
	cluster.NumConns = 4 // Increase number of connections per host

	// Configure retry and failover behavior
	cluster.DisableInitialHostLookup = false
	cluster.Port = 9042

	// Set timeouts
	cluster.Timeout = 5 * time.Second         // Query timeout
	cluster.ConnectTimeout = 10 * time.Second // Initial connection timeout

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

// GetSession returns the current session with the specified consistency level
func (c *Connection) GetSession(consistency gocql.Consistency) *gocql.Session {
	if c.session != nil {
		c.session.SetConsistency(consistency)
	}
	return c.session
}

// Session returns the default session
func (c *Connection) Session() *gocql.Session {
	return c.session
}

func (c *Connection) Close() {
	if c.session != nil {
		c.session.Close()
	}
}