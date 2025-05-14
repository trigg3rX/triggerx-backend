package database

import (
    "fmt"
    "os"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    Hosts       []string
    Keyspace    string
    Timeout     time.Duration
    Retries     int
    ConnectWait time.Duration
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", "error", err)
	}

	dbHost := os.Getenv("DATABASE_DOCKER_IP_ADDRESS")
	dbPort := os.Getenv("DATABASE_DOCKER_PORT")

	return &Config{
		Hosts:       []string{dbHost + ":" + dbPort},
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