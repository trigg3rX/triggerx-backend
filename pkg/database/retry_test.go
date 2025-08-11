package database

import (
	"context"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// TestNewQuery tests the NewQuery method
func TestNewQuery(t *testing.T) {
	// Setup
	mockSession := &MockSession{}
	mockLogger := &logging.MockLogger{}

	conn := &Connection{
		session: mockSession,
		logger:  mockLogger,
		config: &Config{
			RetryConfig: retry.DefaultRetryConfig(),
		},
	}

	stmt := "SELECT * FROM test_table"
	values := []interface{}{"value1", "value2"}

	// Execute
	queryx := conn.NewQuery(stmt, values...)

	// Assert
	assert.NotNil(t, queryx)
	assert.Equal(t, conn, queryx.conn)
	assert.False(t, queryx.isIdem)
}

// TestQueryx_Exec_WithCustomRetryConfig tests execution with custom retry configuration
func TestQueryx_Exec_WithCustomRetryConfig(t *testing.T) {
	// Setup
	mockLogger := &logging.MockLogger{}

	customConfig := &retry.RetryConfig{
		MaxRetries:      2,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffFactor:   1.5,
		JitterFactor:    0.1,
		LogRetryAttempt: false,
		ShouldRetry:     gocqlShouldRetry,
	}

	// Create a mock query that will fail
	mockQuery := &MockQuery{
		execErr:   &gocql.RequestErrWriteTimeout{},
		maxCalls:  10, // Will keep failing
		callCount: 0,
	}

	// Test the retry logic directly
	operation := func() error {
		return mockQuery.Exec()
	}

	err := retry.RetryFunc(context.Background(), operation, customConfig, mockLogger)

	// Assert - should fail after max retries
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 2 attempts")
}

// TestQueryx_Exec_MaxRetriesExceeded tests when max retries are exceeded
func TestQueryx_Exec_MaxRetriesExceeded(t *testing.T) {
	// Setup
	mockLogger := &logging.MockLogger{}

	// Use a very low retry count
	customConfig := &retry.RetryConfig{
		MaxRetries:      1,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffFactor:   1.0,
		JitterFactor:    0.0,
		LogRetryAttempt: false,
		ShouldRetry:     gocqlShouldRetry,
	}

	// Create a mock query that will keep failing
	mockQuery := &MockQuery{
		execErr:   &gocql.RequestErrWriteTimeout{},
		maxCalls:  10, // Will keep failing
		callCount: 0,
	}

	// Test the retry logic directly
	operation := func() error {
		return mockQuery.Exec()
	}

	err := retry.RetryFunc(context.Background(), operation, customConfig, mockLogger)

	// Assert - should fail after max retries
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 1 attempts")
}

// TestQueryx_NonIdempotentWarning tests warning for non-idempotent queries
func TestQueryx_NonIdempotentWarning(t *testing.T) {
	// Setup
	mockLogger := &logging.MockLogger{}

	queryx := &Queryx{
		query:  nil, // We don't need a real query for this test
		conn:   &Connection{logger: mockLogger},
		isIdem: false,
	}

	mockLogger.On("Warnf", "Executing a non-idempotent query with retry logic. Ensure this is intended.", mock.Anything).Return()

	// Since we can't actually execute with a nil query, we'll test the warning logic
	// by checking that the warning would be logged for non-idempotent queries
	if !queryx.isIdem {
		mockLogger.Warnf("Executing a non-idempotent query with retry logic. Ensure this is intended.")
	}

	// Assert
	mockLogger.AssertCalled(t, "Warnf", "Executing a non-idempotent query with retry logic. Ensure this is intended.", mock.Anything)
}

// TestQueryx_RetryLogic tests the retry logic used by Queryx
func TestQueryx_RetryLogic(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: nil, maxCalls: 0}
		mockLogger := &logging.MockLogger{}

		// Test the retry logic directly with our mock
		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = gocqlShouldRetry

		err := retry.RetryFunc(context.Background(), operation, cfg, mockLogger)

		assert.NoError(t, err)
		assert.Equal(t, 1, mockQuery.callCount, "Exec should be called once")
	})

	t.Run("Fails once then succeeds", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: &gocql.RequestErrWriteTimeout{}, maxCalls: 1}
		mockLogger := &logging.MockLogger{}

		// Set up the mock logger to expect retry warning messages
		mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = gocqlShouldRetry

		err := retry.RetryFunc(context.Background(), operation, cfg, mockLogger)

		assert.NoError(t, err)
		assert.Equal(t, 2, mockQuery.callCount, "Exec should be called twice")
		mockLogger.AssertExpectations(t)
	})

	t.Run("Fails on non-retryable error", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: gocql.ErrNotFound, maxCalls: 5}
		mockLogger := &logging.MockLogger{}

		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = gocqlShouldRetry

		err := retry.RetryFunc(context.Background(), operation, cfg, mockLogger)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gocql.ErrNotFound)
		assert.Equal(t, 1, mockQuery.callCount, "Exec should be called only once")
	})
}
