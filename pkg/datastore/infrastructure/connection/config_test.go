package connection

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// TestConfig_NewConfig tests the NewConfig function
func TestConfig_NewConfig(t *testing.T) {
	host := "localhost"
	port := "9042"
	config := NewConfig(host, port)

	assert.Equal(t, []string{"localhost:9042"}, config.Hosts)
	assert.Equal(t, "triggerx", config.Keyspace)
	assert.Equal(t, time.Second*30, config.Timeout)
	assert.Equal(t, 5, config.Retries)
	assert.Equal(t, time.Second*10, config.ConnectWait)
	assert.Equal(t, gocql.Quorum, config.Consistency)
	assert.Equal(t, time.Second*15, config.HealthCheckInterval)
	assert.Equal(t, 4, config.ProtoVersion)
	assert.Equal(t, 15*time.Second, config.SocketKeepalive)
	assert.Equal(t, 1000, config.MaxPreparedStmts)
	assert.Equal(t, true, config.DefaultIdempotence)
	assert.NotNil(t, config.RetryConfig)
}

// TestConfig_WithHosts tests the WithHosts method
func TestConfig_WithHosts(t *testing.T) {
	config := &Config{}
	hosts := []string{"host1:9042", "host2:9042"}

	result := config.WithHosts(hosts)

	assert.Equal(t, config, result)
	assert.Equal(t, hosts, config.Hosts)
}

// TestConfig_WithKeyspace tests the WithKeyspace method
func TestConfig_WithKeyspace(t *testing.T) {
	config := &Config{}
	keyspace := "test_keyspace"

	result := config.WithKeyspace(keyspace)

	assert.Equal(t, config, result)
	assert.Equal(t, keyspace, config.Keyspace)
}

// TestConfig_WithTimeout tests the WithTimeout method
func TestConfig_WithTimeout(t *testing.T) {
	config := &Config{}
	timeout := time.Second * 60

	result := config.WithTimeout(timeout)

	assert.Equal(t, config, result)
	assert.Equal(t, timeout, config.Timeout)
}

// TestConfig_WithRetries tests the WithRetries method
func TestConfig_WithRetries(t *testing.T) {
	config := &Config{}
	retries := 10

	result := config.WithRetries(retries)

	assert.Equal(t, config, result)
	assert.Equal(t, retries, config.Retries)
}

// TestConfig_WithRetryConfig tests the WithRetryConfig method
func TestConfig_WithRetryConfig(t *testing.T) {
	config := &Config{}
	retryConfig := retry.DefaultRetryConfig()

	result := config.WithRetryConfig(retryConfig)

	assert.Equal(t, config, result)
	assert.Equal(t, retryConfig, config.RetryConfig)
}

// TestConfig_WithHealthCheckInterval tests the WithHealthCheckInterval method
func TestConfig_WithHealthCheckInterval(t *testing.T) {
	config := &Config{}
	interval := time.Second * 30

	result := config.WithHealthCheckInterval(interval)

	assert.Equal(t, config, result)
	assert.Equal(t, interval, config.HealthCheckInterval)
}

// TestConfig_WithNilRetryConfig tests WithRetryConfig with nil
func TestConfig_WithNilRetryConfig(t *testing.T) {
	config := &Config{}

	result := config.WithRetryConfig(nil)

	assert.Equal(t, config, result)
	assert.Nil(t, config.RetryConfig)
}

// TestConfig_WithEmptyHosts tests WithHosts with empty slice
func TestConfig_WithEmptyHosts(t *testing.T) {
	config := &Config{}
	hosts := []string{}

	result := config.WithHosts(hosts)

	assert.Equal(t, config, result)
	assert.Equal(t, hosts, config.Hosts)
	assert.Len(t, config.Hosts, 0)
}

// TestConfig_WithEmptyKeyspace tests WithKeyspace with empty string
func TestConfig_WithEmptyKeyspace(t *testing.T) {
	config := &Config{}
	keyspace := ""

	result := config.WithKeyspace(keyspace)

	assert.Equal(t, config, result)
	assert.Equal(t, keyspace, config.Keyspace)
}

// TestConfig_WithZeroTimeout tests WithTimeout with zero duration
func TestConfig_WithZeroTimeout(t *testing.T) {
	config := &Config{}
	timeout := time.Duration(0)

	result := config.WithTimeout(timeout)

	assert.Equal(t, config, result)
	assert.Equal(t, timeout, config.Timeout)
}

// TestConfig_WithNegativeRetries tests WithRetries with negative value
func TestConfig_WithNegativeRetries(t *testing.T) {
	config := &Config{}
	retries := -1

	result := config.WithRetries(retries)

	assert.Equal(t, config, result)
	assert.Equal(t, retries, config.Retries)
}

// TestConfig_WithZeroHealthCheckInterval tests WithHealthCheckInterval with zero duration
func TestConfig_WithZeroHealthCheckInterval(t *testing.T) {
	config := &Config{}
	interval := time.Duration(0)

	result := config.WithHealthCheckInterval(interval)

	assert.Equal(t, config, result)
	assert.Equal(t, interval, config.HealthCheckInterval)
}

// TestNewConfig_WithEmptyHost tests NewConfig with empty host
func TestNewConfig_WithEmptyHost(t *testing.T) {
	host := ""
	port := "9042"
	config := NewConfig(host, port)

	expectedHosts := []string{":9042"}
	assert.Equal(t, expectedHosts, config.Hosts)
	assert.Equal(t, "triggerx", config.Keyspace)
}

// TestNewConfig_WithEmptyPort tests NewConfig with empty port
func TestNewConfig_WithEmptyPort(t *testing.T) {
	host := "localhost"
	port := ""
	config := NewConfig(host, port)

	expectedHosts := []string{"localhost:"}
	assert.Equal(t, expectedHosts, config.Hosts)
	assert.Equal(t, "triggerx", config.Keyspace)
}

// TestConfig_MethodChaining tests method chaining on Config
func TestConfig_MethodChaining(t *testing.T) {
	config := &Config{}
	hosts := []string{"host1:9042", "host2:9042"}
	keyspace := "test_keyspace"
	timeout := time.Second * 60
	retries := 10
	retryConfig := retry.DefaultRetryConfig()
	interval := time.Second * 30

	// Test method chaining
	result := config.
		WithHosts(hosts).
		WithKeyspace(keyspace).
		WithTimeout(timeout).
		WithRetries(retries).
		WithRetryConfig(retryConfig).
		WithHealthCheckInterval(interval)

	// All methods should return the same instance for chaining
	assert.Equal(t, config, result)
	assert.Equal(t, hosts, config.Hosts)
	assert.Equal(t, keyspace, config.Keyspace)
	assert.Equal(t, timeout, config.Timeout)
	assert.Equal(t, retries, config.Retries)
	assert.Equal(t, retryConfig, config.RetryConfig)
	assert.Equal(t, interval, config.HealthCheckInterval)
}

// TestConfig_Validate tests the Validate method with valid configuration
func TestConfig_Validate(t *testing.T) {
	config := &Config{
		Hosts:               []string{"localhost:9042"},
		Keyspace:            "test_keyspace",
		Timeout:             time.Second * 30,
		ConnectWait:         time.Second * 10,
		Retries:             5,
		ProtoVersion:        4,
		MaxPreparedStmts:    1000,
		HealthCheckInterval: time.Second * 15,
	}

	err := config.Validate()
	assert.NoError(t, err)
}

// TestConfig_Validate_EmptyHosts tests Validate with empty hosts
func TestConfig_Validate_EmptyHosts(t *testing.T) {
	config := &Config{
		Hosts:    []string{},
		Keyspace: "test_keyspace",
		Timeout:  time.Second * 30,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one host must be specified")
}

// TestConfig_Validate_EmptyKeyspace tests Validate with empty keyspace
func TestConfig_Validate_EmptyKeyspace(t *testing.T) {
	config := &Config{
		Hosts:    []string{"localhost:9042"},
		Keyspace: "",
		Timeout:  time.Second * 30,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyspace cannot be empty")
}

// TestConfig_Validate_ZeroTimeout tests Validate with zero timeout
func TestConfig_Validate_ZeroTimeout(t *testing.T) {
	config := &Config{
		Hosts:    []string{"localhost:9042"},
		Keyspace: "test_keyspace",
		Timeout:  0,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")
}

// TestConfig_Validate_NegativeTimeout tests Validate with negative timeout
func TestConfig_Validate_NegativeTimeout(t *testing.T) {
	config := &Config{
		Hosts:    []string{"localhost:9042"},
		Keyspace: "test_keyspace",
		Timeout:  -time.Second,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")
}

// TestConfig_Validate_ZeroConnectWait tests Validate with zero connect wait
func TestConfig_Validate_ZeroConnectWait(t *testing.T) {
	config := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Second * 30,
		ConnectWait: 0,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect wait must be positive")
}

// TestConfig_Validate_NegativeRetries tests Validate with negative retries
func TestConfig_Validate_NegativeRetries(t *testing.T) {
	config := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Second * 30,
		ConnectWait: time.Second * 10,
		Retries:     -1,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retries cannot be negative")
}

// TestConfig_Validate_InvalidProtoVersion tests Validate with invalid protocol version
func TestConfig_Validate_InvalidProtoVersion(t *testing.T) {
	config := &Config{
		Hosts:        []string{"localhost:9042"},
		Keyspace:     "test_keyspace",
		Timeout:      time.Second * 30,
		ConnectWait:  time.Second * 10,
		ProtoVersion: 5, // Invalid version
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "protocol version must be between 1 and 4")
}

// TestConfig_Validate_ZeroProtoVersion tests Validate with zero protocol version
func TestConfig_Validate_ZeroProtoVersion(t *testing.T) {
	config := &Config{
		Hosts:        []string{"localhost:9042"},
		Keyspace:     "test_keyspace",
		Timeout:      time.Second * 30,
		ConnectWait:  time.Second * 10,
		ProtoVersion: 0, // Invalid version
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "protocol version must be between 1 and 4")
}

// TestConfig_Validate_NegativeMaxPreparedStmts tests Validate with negative max prepared statements
func TestConfig_Validate_NegativeMaxPreparedStmts(t *testing.T) {
	config := &Config{
		Hosts:            []string{"localhost:9042"},
		Keyspace:         "test_keyspace",
		Timeout:          time.Second * 30,
		ConnectWait:      time.Second * 10,
		ProtoVersion:     4,
		MaxPreparedStmts: -1,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max prepared statements cannot be negative")
}

// TestConfig_Validate_NegativeHealthCheckInterval tests Validate with negative health check interval
func TestConfig_Validate_NegativeHealthCheckInterval(t *testing.T) {
	config := &Config{
		Hosts:               []string{"localhost:9042"},
		Keyspace:            "test_keyspace",
		Timeout:             time.Second * 30,
		ConnectWait:         time.Second * 10,
		ProtoVersion:        4,
		HealthCheckInterval: -time.Second,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check interval cannot be negative")
}

// TestConfig_Validate_MultipleErrors tests Validate with multiple validation errors
func TestConfig_Validate_MultipleErrors(t *testing.T) {
	config := &Config{
		Hosts:               []string{},   // Empty hosts
		Keyspace:            "",           // Empty keyspace
		Timeout:             0,            // Zero timeout
		ConnectWait:         0,            // Zero connect wait
		Retries:             -1,           // Negative retries
		ProtoVersion:        5,            // Invalid protocol version
		MaxPreparedStmts:    -1,           // Negative max prepared statements
		HealthCheckInterval: -time.Second, // Negative health check interval
	}

	err := config.Validate()
	assert.Error(t, err)
	// Should contain at least one of the error messages
	errorMsg := err.Error()
	assert.True(t,
		contains(errorMsg, "at least one host must be specified") ||
			contains(errorMsg, "keyspace cannot be empty") ||
			contains(errorMsg, "timeout must be positive") ||
			contains(errorMsg, "connect wait must be positive") ||
			contains(errorMsg, "retries cannot be negative") ||
			contains(errorMsg, "protocol version must be between 1 and 4") ||
			contains(errorMsg, "max prepared statements cannot be negative") ||
			contains(errorMsg, "health check interval cannot be negative"),
		"Error message should contain at least one validation error: %s", errorMsg)
}

// TestConfig_Clone tests the Clone method
func TestConfig_Clone(t *testing.T) {
	original := &Config{
		Hosts:               []string{"host1:9042", "host2:9042"},
		Keyspace:            "test_keyspace",
		Timeout:             time.Second * 60,
		Retries:             10,
		ConnectWait:         time.Second * 20,
		Consistency:         gocql.Quorum,
		HealthCheckInterval: time.Second * 30,
		ProtoVersion:        4,
		SocketKeepalive:     30 * time.Second,
		MaxPreparedStmts:    2000,
		DefaultIdempotence:  true,
		RetryConfig:         retry.DefaultRetryConfig(),
	}

	cloned := original.Clone()

	// Should be a different instance
	assert.NotSame(t, original, cloned)

	// But should have the same values
	assert.Equal(t, original.Hosts, cloned.Hosts)
	assert.Equal(t, original.Keyspace, cloned.Keyspace)
	assert.Equal(t, original.Timeout, cloned.Timeout)
	assert.Equal(t, original.Retries, cloned.Retries)
	assert.Equal(t, original.ConnectWait, cloned.ConnectWait)
	assert.Equal(t, original.Consistency, cloned.Consistency)
	assert.Equal(t, original.HealthCheckInterval, cloned.HealthCheckInterval)
	assert.Equal(t, original.ProtoVersion, cloned.ProtoVersion)
	assert.Equal(t, original.SocketKeepalive, cloned.SocketKeepalive)
	assert.Equal(t, original.MaxPreparedStmts, cloned.MaxPreparedStmts)
	assert.Equal(t, original.DefaultIdempotence, cloned.DefaultIdempotence)
	assert.Equal(t, original.RetryConfig, cloned.RetryConfig)

	// Modifying the original should not affect the clone
	original.Hosts = []string{"modified:9042"}
	original.Keyspace = "modified_keyspace"
	original.Timeout = time.Second * 120

	assert.NotEqual(t, original.Hosts, cloned.Hosts)
	assert.NotEqual(t, original.Keyspace, cloned.Keyspace)
	assert.NotEqual(t, original.Timeout, cloned.Timeout)
}

// TestConfig_Clone_EmptyHosts tests Clone with empty hosts slice
func TestConfig_Clone_EmptyHosts(t *testing.T) {
	original := &Config{
		Hosts:    []string{},
		Keyspace: "test_keyspace",
		Timeout:  time.Second * 30,
	}

	cloned := original.Clone()

	assert.NotSame(t, original, cloned)
	assert.Equal(t, original.Hosts, cloned.Hosts)
	assert.Len(t, cloned.Hosts, 0)
}

// TestConfig_Clone_NilRetryConfig tests Clone with nil retry config
func TestConfig_Clone_NilRetryConfig(t *testing.T) {
	original := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Second * 30,
		RetryConfig: nil,
	}

	cloned := original.Clone()

	assert.NotSame(t, original, cloned)
	assert.Nil(t, cloned.RetryConfig)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}
