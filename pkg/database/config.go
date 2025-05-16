package database

import (
	"time"
)

type Config struct {
	Hosts       []string
	Keyspace    string
	Timeout     time.Duration
	Retries     int
	ConnectWait time.Duration
	Consistency string
}

// NewConfig returns configuration for both ScyllaDB nodes
func NewConfig() *Config {
	return &Config{
		Hosts:       []string{"localhost:9042", "localhost:9043"},
		Keyspace:    "triggerx",
		Timeout:     time.Second * 30,
		Retries:     5,
		ConnectWait: time.Second * 10,
		Consistency: "LOCAL_QUORUM",
	}
}

// NewScylla1Config returns configuration specifically for ScyllaDB node 1
func NewScylla1Config() *Config {
	return &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "triggerx",
		Timeout:     time.Second * 30,
		Retries:     5,
		ConnectWait: time.Second * 10,
		Consistency: "ONE",
	}
}

// NewScylla2Config returns configuration specifically for ScyllaDB node 2
func NewScylla2Config() *Config {
	return &Config{
		Hosts:       []string{"localhost:9043"},
		Keyspace:    "triggerx",
		Timeout:     time.Second * 30,
		Retries:     5,
		ConnectWait: time.Second * 10,
		Consistency: "ONE",
	}
}

func (c *Config) WithHosts(hosts []string) *Config {
	c.Hosts = hosts
	return c
}

func (c *Config) WithKeyspace(keyspace string) *Config {
	c.Keyspace = keyspace
	return c
}

// NewAlternativeConfig creates a new configuration for an alternative database connection
// with different settings optimized for higher availability
func NewAlternativeConfig() *Config {
	return &Config{
		Hosts:       []string{"triggerx-scylla-1:9042", "triggerx-scylla-2:9042"},
		Keyspace:    "triggerx_alt",   // Different keyspace for separation
		Timeout:     time.Second * 15, // Lower timeout for faster failure detection
		Retries:     3,                // Fewer retries but faster failover
		ConnectWait: time.Second * 5,  // Faster connection attempts
		Consistency: "ONE",            // Lower consistency for better availability
	}
}