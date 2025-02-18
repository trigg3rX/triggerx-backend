package config

import (
    "time"
)

type Config struct {
    Hosts       []string
    Keyspace    string
    Timeout     time.Duration
    Retries     int
    ConnectWait time.Duration
}

func NewConfig() *Config {
    return &Config{
        Hosts:       []string{"localhost:9042"},
        Keyspace:    "triggerx",
        Timeout:     time.Second * 30,
        Retries:     5,
        ConnectWait: time.Second * 10,
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