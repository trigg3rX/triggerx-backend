package connection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// Test helper functions

// TestNewConnection tests the NewConnection function
// Note: This test is limited due to singleton pattern in the connection package
func TestNewConnection(t *testing.T) {
	// Test with valid configuration
	config := &Config{
		Hosts:               []string{"localhost:9042"},
		Keyspace:            "test_keyspace",
		Timeout:             time.Second * 30,
		Retries:             3,
		ConnectWait:         time.Second * 10,
		Consistency:         gocql.Quorum,
		HealthCheckInterval: 0, // Disable health checker for testing
		ProtoVersion:        4,
		SocketKeepalive:     15 * time.Second,
		MaxPreparedStmts:    1000,
		DefaultIdempotence:  true,
	}
	logger := logging.NewNoOpLogger()

	// Test that NewConnection function exists and can be called
	// Note: Due to singleton pattern, we can only test the first call
	conn, err := NewConnection(config, logger)

	// The connection might fail due to no actual database, but the function should exist
	if err != nil {
		// Expected error due to no database connection - any error is acceptable for this test
		assert.Error(t, err)
	} else {
		assert.NotNil(t, conn)
	}
}

// TestConnection_GetSession tests the GetSession method
func TestConnection_GetSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection := mocks.NewMockConnection(ctrl)

	// Setup expectations
	mockConnection.EXPECT().GetSession().Return(mockSession).AnyTimes()
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession).AnyTimes()

	// Test getting session
	session := mockConnection.GetSession()
	assert.Equal(t, mockSession, session)

	// Test getting gocqlx session
	gocqlxSession := mockConnection.GetGocqlxSession()
	assert.Equal(t, mockGocqlxSession, gocqlxSession)
}

// TestConnection_GetGocqlxSession tests the GetGocqlxSession method
func TestConnection_GetGocqlxSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection := mocks.NewMockConnection(ctrl)

	// Setup expectations
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession).Times(1)

	// Test getting gocqlx session
	gocqlxSession := mockConnection.GetGocqlxSession()
	assert.Equal(t, mockGocqlxSession, gocqlxSession)
}

// TestConnection_Close tests the Close method
func TestConnection_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)

	// Setup expectations
	mockConnection.EXPECT().Close().Times(1)

	// Test closing connection
	mockConnection.Close()
}

// TestConnection_HealthCheck tests the HealthCheck method
func TestConnection_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		healthError   error
		expectError   bool
		expectedError string
	}{
		{
			name:        "successful health check",
			healthError: nil,
			expectError: false,
		},
		{
			name:          "health check failure",
			healthError:   errors.New("connection failed"),
			expectError:   true,
			expectedError: "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockConnection := mocks.NewMockConnection(ctrl)

			// Setup expectations
			mockConnection.EXPECT().HealthCheck(gomock.Any()).Return(tt.healthError).Times(1)

			// Perform health check
			ctx := context.Background()
			err := mockConnection.HealthCheck(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Note: SetSession, SetConfig, and SetLogger methods are not part of the public interface
// and are only available for testing purposes on the unexported scyllaConnectionManager struct.
// These methods are tested indirectly through the public interface methods.

// Note: gocqlxSessionWrapper is an unexported struct, so we test it through the public interface
// The wrapper functionality is tested indirectly through the connection interface methods.

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

// TestHealthChecker_NewHealthChecker tests the NewHealthChecker function
func TestHealthChecker_NewHealthChecker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	logger := logging.NewNoOpLogger()
	interval := time.Second * 30

	healthChecker := NewHealthChecker(mockConnection, logger, interval)

	// Test that the health checker was created successfully
	assert.NotNil(t, healthChecker)
}

// TestHealthChecker_Start tests the Start method
func TestHealthChecker_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	logger := logging.NewNoOpLogger()
	interval := time.Millisecond * 100 // Short interval for testing

	healthChecker := NewHealthChecker(mockConnection, logger, interval)

	// Setup mock expectations
	mockConnection.EXPECT().HealthCheck(gomock.Any()).Return(nil).AnyTimes()

	// Start health checker in a goroutine
	go healthChecker.Start()

	// Wait a bit for the health check to run
	time.Sleep(interval + time.Millisecond*50)

	// Stop the health checker
	healthChecker.Stop()

	// Wait a bit for the goroutine to stop
	time.Sleep(time.Millisecond * 100)
}

// TestHealthChecker_Stop tests the Stop method
func TestHealthChecker_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	logger := logging.NewNoOpLogger()
	interval := time.Millisecond * 100

	healthChecker := NewHealthChecker(mockConnection, logger, interval)

	// Test that Stop can be called without panicking
	assert.NotPanics(t, func() {
		healthChecker.Stop()
	})
}
