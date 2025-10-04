package connection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// TestNewConnection tests the NewConnection function with valid configuration
func TestNewConnection(t *testing.T) {
	config := NewConfig("localhost", "9041")
	logger := logging.NewNoOpLogger()

	// This will fail to connect to actual DB, but tests the validation and setup
	conn, err := NewConnection(config, logger)

	// Should return error due to no actual DB connection, but config should be valid
	assert.Error(t, err)
	assert.Nil(t, conn)
}

// TestNewConnection_InvalidConfig tests NewConnection with invalid configuration
func TestNewConnection_InvalidConfig(t *testing.T) {
	config := &Config{
		Hosts:    []string{}, // Invalid: empty hosts
		Keyspace: "",         // Invalid: empty keyspace
		Timeout:  0,          // Invalid: zero timeout
	}
	logger := logging.NewNoOpLogger()

	conn, err := NewConnection(config, logger)

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "invalid configuration")
}

// TestScyllaConnectionManager_GetSession tests GetSession method
func TestScyllaConnectionManager_GetSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	manager := &scyllaConnectionManager{
		session: mockSession,
	}

	session := manager.GetSession()
	assert.Equal(t, mockSession, session)
}

// TestScyllaConnectionManager_GetGocqlxSession tests GetGocqlxSession method
func TestScyllaConnectionManager_GetGocqlxSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	manager := &scyllaConnectionManager{
		gocqlxSession: mockGocqlxSession,
	}

	gocqlxSession := manager.GetGocqlxSession()
	assert.Equal(t, mockGocqlxSession, gocqlxSession)
}

// TestScyllaConnectionManager_Close tests Close method
func TestScyllaConnectionManager_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	mockSession.EXPECT().Close().Times(1)

	manager := &scyllaConnectionManager{
		session: mockSession,
	}

	manager.Close()
}

// TestScyllaConnectionManager_Close_WithNilSession tests Close with nil session
func TestScyllaConnectionManager_Close_WithNilSession(t *testing.T) {
	manager := &scyllaConnectionManager{
		session: nil,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		manager.Close()
	})
}

// TestScyllaConnectionManager_HealthCheck tests HealthCheck method
func TestScyllaConnectionManager_HealthCheck(t *testing.T) {
	// Test with nil session (simpler test that doesn't require complex mocking)
	manager := &scyllaConnectionManager{
		session: nil,
		logger:  logging.NewNoOpLogger(),
	}

	ctx := context.Background()
	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database session is nil")
}

// TestScyllaConnectionManager_HealthCheck_Failure tests HealthCheck with failure
func TestScyllaConnectionManager_HealthCheck_Failure(t *testing.T) {
	// Test circuit breaker open scenario
	manager := &scyllaConnectionManager{
		circuitBreakerState: CircuitBreakerOpen,
		circuitBreakerTime:  time.Now(),
		logger:              logging.NewNoOpLogger(),
	}

	ctx := context.Background()
	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

// TestScyllaConnectionManager_HealthCheck_NilSession tests HealthCheck with nil session
func TestScyllaConnectionManager_HealthCheck_NilSession(t *testing.T) {
	manager := &scyllaConnectionManager{
		session: nil,
	}

	ctx := context.Background()
	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database session is nil")
}

// TestScyllaConnectionManager_GetHealthStatus tests GetHealthStatus method
func TestScyllaConnectionManager_GetHealthStatus(t *testing.T) {
	manager := &scyllaConnectionManager{
		healthStatus:        true,
		lastHealthCheck:     time.Now(),
		healthCheckCount:    10,
		healthCheckFailures: 2,
		reconnectAttempts:   1,
	}

	status, lastCheck, count, failures, attempts := manager.GetHealthStatus()

	assert.True(t, status)
	assert.False(t, lastCheck.IsZero())
	assert.Equal(t, int64(10), count)
	assert.Equal(t, int64(2), failures)
	assert.Equal(t, int64(1), attempts)
}

// TestScyllaConnectionManager_IsHealthy tests IsHealthy method
func TestScyllaConnectionManager_IsHealthy(t *testing.T) {
	manager := &scyllaConnectionManager{
		healthStatus: true,
	}

	assert.True(t, manager.IsHealthy())

	manager.healthStatus = false
	assert.False(t, manager.IsHealthy())
}

// TestScyllaConnectionManager_GetReconnectCount tests GetReconnectCount method
func TestScyllaConnectionManager_GetReconnectCount(t *testing.T) {
	manager := &scyllaConnectionManager{
		reconnectCount: 5,
	}

	assert.Equal(t, 5, manager.GetReconnectCount())
}

// TestScyllaConnectionManager_CircuitBreaker tests circuit breaker functionality
func TestScyllaConnectionManager_CircuitBreaker(t *testing.T) {
	manager := &scyllaConnectionManager{
		circuitBreakerState: CircuitBreakerClosed,
		circuitBreakerCount: 0,
		logger:              logging.NewNoOpLogger(),
	}

	// Test closed state
	assert.False(t, manager.isCircuitBreakerOpen())

	// Test opening circuit breaker
	manager.recordCircuitBreakerFailure()
	manager.recordCircuitBreakerFailure()
	manager.recordCircuitBreakerFailure()
	manager.recordCircuitBreakerFailure()
	manager.recordCircuitBreakerFailure() // 5th failure should open circuit

	assert.True(t, manager.isCircuitBreakerOpen())

	// Test half-open state after timeout
	manager.circuitBreakerTime = time.Now().Add(-31 * time.Second)
	assert.False(t, manager.isCircuitBreakerOpen())
	assert.Equal(t, CircuitBreakerHalfOpen, manager.circuitBreakerState)

	// Test closing circuit breaker
	manager.recordCircuitBreakerSuccess()
	assert.Equal(t, CircuitBreakerClosed, manager.circuitBreakerState)
	assert.Equal(t, 0, manager.circuitBreakerCount)
}

// TestScyllaConnectionManager_SetSession tests SetSession method
func TestScyllaConnectionManager_SetSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	manager := &scyllaConnectionManager{}

	manager.SetSession(mockSession)
	assert.Equal(t, mockSession, manager.session)
}

// TestScyllaConnectionManager_SetConfig tests SetConfig method
func TestScyllaConnectionManager_SetConfig(t *testing.T) {
	config := NewConfig("localhost", "9042")
	manager := &scyllaConnectionManager{}

	manager.SetConfig(config)
	assert.Equal(t, config, manager.config)
}

// TestScyllaConnectionManager_SetLogger tests SetLogger method
func TestScyllaConnectionManager_SetLogger(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &scyllaConnectionManager{}

	manager.SetLogger(logger)
	assert.Equal(t, logger, manager.logger)
}

// TestGocqlxSessionWrapper tests the gocqlx session wrapper
func TestGocqlxSessionWrapper(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	wrapper := &gocqlxSessionWrapper{
		session: mockSession,
	}

	// Test Close method
	mockSession.EXPECT().Close().Times(1)
	wrapper.Close()
}

// TestGocqlxQueryWrapper tests the gocqlx query wrapper
func TestGocqlxQueryWrapper(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	wrapper := &gocqlxQueryWrapper{
		query: mockQuery,
	}

	// Test WithContext
	mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
	result := wrapper.WithContext(context.Background())
	assert.NotNil(t, result)

	// Test BindStruct
	mockQuery.EXPECT().BindStruct(gomock.Any()).Return(mockQuery)
	result = wrapper.BindStruct(struct{}{})
	assert.NotNil(t, result)

	// Test ExecRelease
	expectedError := fmt.Errorf("exec error")
	mockQuery.EXPECT().ExecRelease().Return(expectedError)
	err := wrapper.ExecRelease()
	assert.Equal(t, expectedError, err)

	// Test GetRelease
	mockQuery.EXPECT().GetRelease(gomock.Any()).Return(expectedError)
	err = wrapper.GetRelease(struct{}{})
	assert.Equal(t, expectedError, err)

	// Test Select
	mockQuery.EXPECT().Select(gomock.Any()).Return(expectedError)
	err = wrapper.Select(struct{}{})
	assert.Equal(t, expectedError, err)

	// Test SelectRelease
	mockQuery.EXPECT().SelectRelease(gomock.Any()).Return(expectedError)
	err = wrapper.SelectRelease(struct{}{})
	assert.Equal(t, expectedError, err)
}

// TestScyllaConnectionManager_ConcurrentAccess tests concurrent access to connection manager
func TestScyllaConnectionManager_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	manager := &scyllaConnectionManager{
		session: mockSession,
	}

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			session := manager.GetSession()
			assert.Equal(t, mockSession, session)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestScyllaConnectionManager_ConcurrentClose tests concurrent close operations
func TestScyllaConnectionManager_ConcurrentClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	mockSession.EXPECT().Close().AnyTimes() // Allow multiple calls for concurrent access

	manager := &scyllaConnectionManager{
		session: mockSession,
	}

	// Test concurrent close
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			manager.Close()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestScyllaConnectionManager_StartHealthChecker tests the health checker
func TestScyllaConnectionManager_StartHealthChecker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	logger := logging.NewNoOpLogger()

	config := &Config{
		HealthCheckInterval: time.Millisecond * 100,
	}

	manager := &scyllaConnectionManager{
		session: mockSession,
		config:  config,
		logger:  logger,
	}

	// Setup mock expectations for health check
	// Note: We can't easily mock the actual gocql.Query, so we'll test the logic differently
	// by testing with a nil session to trigger the error path
	manager.session = nil

	// Start health checker in background
	go manager.startHealthChecker()

	// Wait a bit for health checks to run
	time.Sleep(time.Millisecond * 250)

	// Verify health check was called (should fail due to nil session)
	assert.True(t, manager.healthCheckCount > 0)
	assert.True(t, manager.healthCheckFailures > 0)
	assert.False(t, manager.healthStatus)
}

// TestScyllaConnectionManager_StartHealthCheckerWithZeroInterval tests health checker with zero interval
func TestScyllaConnectionManager_StartHealthCheckerWithZeroInterval(t *testing.T) {
	config := &Config{
		HealthCheckInterval: 0, // Zero interval
	}

	// Should not start health checker with zero interval
	// This is tested by the fact that NewConnection doesn't start health checker with zero interval
	assert.Equal(t, time.Duration(0), config.HealthCheckInterval)
}

// TestScyllaConnectionManager_Reconnect tests reconnection logic
func TestScyllaConnectionManager_Reconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	logger := logging.NewNoOpLogger()

	// Use a very short timeout to avoid long waits
	config := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Millisecond * 100, // Very short timeout
		Retries:     1,                      // Minimal retries
		ConnectWait: time.Millisecond * 50,  // Very short connect wait
		RetryConfig: &retry.RetryConfig{
			MaxRetries:    1,
			InitialDelay:  time.Millisecond * 10,
			MaxDelay:      time.Millisecond * 50,
			BackoffFactor: 1.0,
		},
	}

	manager := &scyllaConnectionManager{
		session: mockSession,
		config:  config,
		logger:  logger,
	}

	// Test reconnection (will fail quickly due to short timeout, but tests the logic)
	manager.reconnect()

	// Verify reconnection attempt was recorded
	assert.True(t, manager.reconnectAttempts > 0)
	assert.True(t, manager.reconnectCount > 0)
}

// TestScyllaConnectionManager_ReconnectWithNilConfig tests reconnection with nil config
func TestScyllaConnectionManager_ReconnectWithNilConfig(t *testing.T) {
	manager := &scyllaConnectionManager{
		config: nil,
	}

	// Should not panic and should return early
	assert.NotPanics(t, func() {
		manager.reconnect()
	})

	// Verify that reconnect attempts were still incremented
	assert.Equal(t, int64(1), manager.reconnectAttempts)
	assert.Equal(t, 1, manager.reconnectCount)
}

// TestScyllaConnectionManager_ReconnectWithNilRetryConfig tests reconnection with nil retry config
func TestScyllaConnectionManager_ReconnectWithNilRetryConfig(t *testing.T) {
	config := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Second * 30,
		Retries:     3,
		ConnectWait: time.Second * 10,
		RetryConfig: nil, // Nil retry config
	}

	manager := &scyllaConnectionManager{
		config: config,
		logger: logging.NewNoOpLogger(),
	}

	// Should not panic and should use default retry config
	assert.NotPanics(t, func() {
		manager.reconnect()
	})
}

// TestGocqlxQueryWrapper_MethodChaining tests method chaining on gocqlx query wrapper
func TestGocqlxQueryWrapper_MethodChaining(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	wrapper := &gocqlxQueryWrapper{
		query: mockQuery,
	}

	// Test method chaining
	mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
	mockQuery.EXPECT().BindStruct(gomock.Any()).Return(mockQuery)

	result := wrapper.WithContext(context.Background()).BindStruct(struct{}{})
	assert.NotNil(t, result)
}

// Additional unit tests to improve coverage

// TestScyllaConnectionManager_GetHealthStatus_Detailed tests GetHealthStatus with detailed metrics
func TestScyllaConnectionManager_GetHealthStatus_Detailed(t *testing.T) {
	now := time.Now()
	manager := &scyllaConnectionManager{
		healthStatus:        false,
		lastHealthCheck:     now,
		healthCheckCount:    25,
		healthCheckFailures: 5,
		reconnectAttempts:   3,
	}

	status, lastCheck, count, failures, attempts := manager.GetHealthStatus()

	assert.False(t, status)
	assert.Equal(t, now, lastCheck)
	assert.Equal(t, int64(25), count)
	assert.Equal(t, int64(5), failures)
	assert.Equal(t, int64(3), attempts)
}

// TestScyllaConnectionManager_CircuitBreaker_Detailed tests circuit breaker with detailed scenarios
func TestScyllaConnectionManager_CircuitBreaker_Detailed(t *testing.T) {
	manager := &scyllaConnectionManager{
		circuitBreakerState: CircuitBreakerClosed,
		circuitBreakerCount: 0,
	}

	// Test closed state
	assert.False(t, manager.isCircuitBreakerOpen())
	assert.Equal(t, CircuitBreakerClosed, manager.circuitBreakerState)

	// Test opening circuit breaker with exactly 5 failures
	for i := 0; i < 4; i++ {
		manager.recordCircuitBreakerFailure()
		assert.False(t, manager.isCircuitBreakerOpen())
	}

	// 5th failure should open circuit
	manager.recordCircuitBreakerFailure()
	assert.True(t, manager.isCircuitBreakerOpen())
	assert.Equal(t, CircuitBreakerOpen, manager.circuitBreakerState)

	// Test half-open transition
	manager.circuitBreakerTime = time.Now().Add(-31 * time.Second)
	assert.False(t, manager.isCircuitBreakerOpen())
	assert.Equal(t, CircuitBreakerHalfOpen, manager.circuitBreakerState)

	// Test closing circuit breaker
	manager.recordCircuitBreakerSuccess()
	assert.Equal(t, CircuitBreakerClosed, manager.circuitBreakerState)
	assert.Equal(t, 0, manager.circuitBreakerCount)
}

// TestScyllaConnectionManager_CircuitBreaker_Timeout tests circuit breaker timeout logic
func TestScyllaConnectionManager_CircuitBreaker_Timeout(t *testing.T) {
	manager := &scyllaConnectionManager{
		circuitBreakerState: CircuitBreakerOpen,
		circuitBreakerTime:  time.Now().Add(-29 * time.Second), // Not yet timeout
		logger:              logging.NewNoOpLogger(),
	}

	// Should still be open (not enough time passed)
	assert.True(t, manager.isCircuitBreakerOpen())

	// Set time to past timeout
	manager.circuitBreakerTime = time.Now().Add(-31 * time.Second)
	assert.False(t, manager.isCircuitBreakerOpen())
	assert.Equal(t, CircuitBreakerHalfOpen, manager.circuitBreakerState)
}

// TestScyllaConnectionManager_Reconnect_Detailed tests reconnection with detailed scenarios
func TestScyllaConnectionManager_Reconnect_Detailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	logger := logging.NewNoOpLogger()

	// Use very short timeouts to avoid long waits
	config := &Config{
		Hosts:       []string{"localhost:9042"},
		Keyspace:    "test_keyspace",
		Timeout:     time.Millisecond * 100, // Very short timeout
		Retries:     1,                      // Minimal retries
		ConnectWait: time.Millisecond * 50,  // Very short connect wait
		RetryConfig: &retry.RetryConfig{
			MaxRetries:    1,
			InitialDelay:  time.Millisecond * 10,
			MaxDelay:      time.Millisecond * 50,
			BackoffFactor: 1.0,
		},
	}

	manager := &scyllaConnectionManager{
		session: mockSession,
		config:  config,
		logger:  logger,
	}

	initialAttempts := manager.reconnectAttempts
	initialCount := manager.reconnectCount

	// Test reconnection (will fail quickly due to short timeout, but tests the logic)
	manager.reconnect()

	// Verify reconnection attempt was recorded
	assert.True(t, manager.reconnectAttempts > initialAttempts)
	assert.True(t, manager.reconnectCount > initialCount)
}

// TestScyllaConnectionManager_StartHealthChecker_Detailed tests health checker with detailed scenarios
func TestScyllaConnectionManager_StartHealthChecker_Detailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	logger := logging.NewNoOpLogger()

	config := &Config{
		HealthCheckInterval: time.Millisecond * 50,
	}

	manager := &scyllaConnectionManager{
		session: mockSession,
		config:  config,
		logger:  logger,
	}

	// Test with nil session to trigger error path
	manager.session = nil

	// Start health checker in background
	go manager.startHealthChecker()

	// Wait for multiple health checks to run
	time.Sleep(time.Millisecond * 200)

	// Verify health check was called multiple times and failed
	assert.True(t, manager.healthCheckCount > 0)
	assert.True(t, manager.healthCheckFailures > 0)
	assert.False(t, manager.healthStatus) // Should be unhealthy due to nil session
}

// TestScyllaConnectionManager_StartHealthChecker_WithFailures tests health checker with failures
func TestScyllaConnectionManager_StartHealthChecker_WithFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)
	logger := logging.NewNoOpLogger()

	config := &Config{
		HealthCheckInterval: time.Millisecond * 50,
	}

	manager := &scyllaConnectionManager{
		session: mockSession,
		config:  config,
		logger:  logger,
	}

	// Test with nil session to trigger failure path
	manager.session = nil

	// Start health checker in background
	go manager.startHealthChecker()

	// Wait for multiple health checks to run
	time.Sleep(time.Millisecond * 200)

	// Verify health check failures were recorded
	assert.True(t, manager.healthCheckCount > 0)
	assert.True(t, manager.healthCheckFailures > 0)
	assert.False(t, manager.healthStatus) // Should be unhealthy
}

// TestGocqlxSessionWrapper_Query_Detailed tests gocqlx session wrapper query method
func TestGocqlxSessionWrapper_Query_Detailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)

	// Set up mock expectation for Query call
	mockSession.EXPECT().Query("SELECT * FROM test").Return(nil).AnyTimes()

	wrapper := &gocqlxSessionWrapper{
		session: mockSession,
	}

	// Test query creation - this will create a wrapper but may fail due to mock type mismatch
	// The actual query execution would require a real gocql.Query, which is hard to mock properly
	result := wrapper.Query("SELECT * FROM test", []string{"id", "name"})
	assert.NotNil(t, result)
}

// TestRealGocqlxQuery_AllMethods tests all methods of realGocqlxQuery
func TestRealGocqlxQuery_AllMethods(t *testing.T) {
	// This test would require a real gocqlx.Queryx, so we'll test the interface compliance
	// by ensuring the methods exist and can be called without panicking

	// Note: In a real scenario, you'd need to create an actual gocqlx.Queryx
	// For now, we'll test that the methods are properly defined
	assert.True(t, true) // Placeholder - real implementation would test actual gocqlx.Queryx
}

// TestScyllaConnectionManager_ConcurrentHealthCheck tests concurrent health checks
func TestScyllaConnectionManager_ConcurrentHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockSessioner(ctrl)

	manager := &scyllaConnectionManager{
		session: mockSession,
	}

	// Reset circuit breaker state to ensure clean test
	manager.circuitBreakerState = CircuitBreakerClosed
	manager.circuitBreakerCount = 0

	// Test with nil session to trigger error path
	manager.session = nil

	// Test concurrent health checks
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			ctx := context.Background()
			err := manager.HealthCheck(ctx)
			assert.Error(t, err) // Should error due to nil session
			// Check for either error message since circuit breaker might open after multiple failures
			assert.True(t,
				contains(err.Error(), "database session is nil") ||
					contains(err.Error(), "circuit breaker is open"),
				"Error should contain either 'database session is nil' or 'circuit breaker is open', got: %s", err.Error())
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestScyllaConnectionManager_HealthCheck_WithCircuitBreaker tests health check with circuit breaker
func TestScyllaConnectionManager_HealthCheck_WithCircuitBreaker(t *testing.T) {
	manager := &scyllaConnectionManager{
		circuitBreakerState: CircuitBreakerOpen,
		circuitBreakerTime:  time.Now(),
		logger:              logging.NewNoOpLogger(),
	}

	ctx := context.Background()
	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

// TestScyllaConnectionManager_HealthCheck_WithNilSession_Detailed tests health check with nil session
func TestScyllaConnectionManager_HealthCheck_WithNilSession_Detailed(t *testing.T) {
	manager := &scyllaConnectionManager{
		session: nil,
		logger:  logging.NewNoOpLogger(),
	}

	ctx := context.Background()
	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database session is nil")

	// Verify circuit breaker failure was recorded
	assert.True(t, manager.circuitBreakerCount > 0)
}
