package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging" // Assuming you have a logger, otherwise use a mock/simple one
)

var (
	testClient *Client
)

// TestMain runs before any other tests in this package.
func TestMain(m *testing.M) {
	// Use a mock logger for tests
	logger := logging.NewNoOpLogger() // Replace with your actual logger initialization if needed

	// Configuration for local Redis test server
	config := RedisConfig{
		UpstashConfig: UpstashConfig{ // We use the URL parser which works for standard redis:// URLs
			URL: "redis://localhost:6379/0",
		},
		ConnectionSettings: ConnectionSettings{
			PoolSize:     10,
			MinIdleConns: 2,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}

	var err error
	testClient, err = NewRedisClient(logger, config)
	if err != nil {
		fmt.Printf("Failed to connect to local Redis for testing: %v\n", err)
		fmt.Println("Please ensure Redis is running on localhost:6379. You can use 'docker compose up -d'.")
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	// Teardown
	err = testClient.Close()
	if err != nil {
		fmt.Printf("Failed to close Redis client: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

// flushDB cleans the entire database. Used to ensure tests are isolated.
func flushDB(t *testing.T) {
	t.Helper()
	err := testClient.Client().FlushDB(context.Background()).Err()
	if err != nil {
		t.Fatalf("Failed to flush Redis DB: %v", err)
	}
}

// createKey is a simple test helper
func createKey(t *testing.T, key, value string) {
	t.Helper()
	err := testClient.Set(context.Background(), key, value, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create test key '%s': %v", key, err)
	}
}
