package redis

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTrackRetryAttempt_Integration tests that trackRetryAttempt is called during retry operations
func TestTrackRetryAttempt_Integration(t *testing.T) {
	// Reset metrics to ensure clean state
	testClient.ResetOperationMetrics()

	// Use the existing testClient but reset its monitoring hooks
	originalHooks := testClient.monitoringHooks
	defer func() {
		testClient.SetMonitoringHooks(originalHooks)
	}()

	// Set up monitoring hooks to track retry attempts
	var retryAttempts []struct {
		operation string
		attempt   int
		err       error
	}
	var mu sync.Mutex

	hooks := &MonitoringHooks{
		OnRetryAttempt: func(operation string, attempt int, err error) {
			mu.Lock()
			defer mu.Unlock()
			retryAttempts = append(retryAttempts, struct {
				operation string
				attempt   int
				err       error
			}{operation, attempt, err})
		},
	}
	testClient.SetMonitoringHooks(hooks)

	// Create a retryable error (io.EOF is retryable)
	retryableErr := io.EOF

	// Test operation that will retry
	ctx := context.Background()
	err := testClient.executeWithRetry(ctx, func() error {
		return retryableErr
	}, "test_operation")

	// Verify that the operation failed after retries
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after")

	// Verify that retry attempts were tracked
	mu.Lock()
	assert.Greater(t, len(retryAttempts), 0, "Should have tracked retry attempts")
	for i, attempt := range retryAttempts {
		assert.Equal(t, "test_operation", attempt.operation)
		assert.Equal(t, i+1, attempt.attempt)
		assert.Equal(t, retryableErr, attempt.err)
	}
	mu.Unlock()

	// Verify that operation metrics were updated
	metrics := testClient.GetOperationMetrics()
	operationMetrics, exists := metrics["test_operation"]
	assert.True(t, exists, "Operation metrics should exist")
	assert.Greater(t, operationMetrics.TotalCalls, int64(1), "Should have multiple total calls")
	assert.Greater(t, operationMetrics.RetryCount, int64(0), "Should have retry attempts")
}

// TestTrackRetryAttempt_NonRetryableError tests that trackRetryAttempt is not called for non-retryable errors
func TestTrackRetryAttempt_NonRetryableError(t *testing.T) {
	// Reset metrics to ensure clean state
	testClient.ResetOperationMetrics()

	// Use the existing testClient but reset its monitoring hooks
	originalHooks := testClient.monitoringHooks
	defer func() {
		testClient.SetMonitoringHooks(originalHooks)
	}()

	// Set up monitoring hooks to track retry attempts
	var retryAttempts []struct {
		operation string
		attempt   int
		err       error
	}
	var mu sync.Mutex

	hooks := &MonitoringHooks{
		OnRetryAttempt: func(operation string, attempt int, err error) {
			mu.Lock()
			defer mu.Unlock()
			retryAttempts = append(retryAttempts, struct {
				operation string
				attempt   int
				err       error
			}{operation, attempt, err})
		},
	}
	testClient.SetMonitoringHooks(hooks)

	// Use a non-retryable error that will actually execute the operation
	// redis.Nil is a good example of a non-retryable error
	nonRetryableErr := errors.New("non-retryable business logic error")

	// Test operation that should not retry
	ctx := context.Background()
	err := testClient.executeWithRetry(ctx, func() error {
		return nonRetryableErr
	}, "test_operation")

	// Verify that the operation failed without retries
	assert.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)

	// Verify that no retry attempts were tracked
	mu.Lock()
	assert.Len(t, retryAttempts, 0, "Should not have tracked any retry attempts for non-retryable error")
	mu.Unlock()

	// Verify that operation metrics were updated
	metrics := testClient.GetOperationMetrics()
	operationMetrics, exists := metrics["test_operation"]
	assert.True(t, exists, "Operation metrics should exist")
	assert.Equal(t, int64(1), operationMetrics.TotalCalls, "Should have 1 total call")
	assert.Equal(t, int64(1), operationMetrics.ErrorCount, "Should have 1 error call")
	assert.Equal(t, int64(0), operationMetrics.RetryCount, "Should have 0 retry attempts")
}

// TestTrackRetryAttempt_SuccessAfterRetry tests that trackRetryAttempt is called correctly when operation succeeds after retries
func TestTrackRetryAttempt_SuccessAfterRetry(t *testing.T) {
	// Reset metrics to ensure clean state
	testClient.ResetOperationMetrics()

	// Use the existing testClient but reset its monitoring hooks
	originalHooks := testClient.monitoringHooks
	defer func() {
		testClient.SetMonitoringHooks(originalHooks)
	}()

	// Set up monitoring hooks to track retry attempts
	var retryAttempts []struct {
		operation string
		attempt   int
		err       error
	}
	var mu sync.Mutex

	hooks := &MonitoringHooks{
		OnRetryAttempt: func(operation string, attempt int, err error) {
			mu.Lock()
			defer mu.Unlock()
			retryAttempts = append(retryAttempts, struct {
				operation string
				attempt   int
				err       error
			}{operation, attempt, err})
		},
	}
	testClient.SetMonitoringHooks(hooks)

	// Create an operation that fails twice then succeeds
	attemptCount := 0
	retryableErr := io.EOF // Use a retryable error

	ctx := context.Background()
	err := testClient.executeWithRetry(ctx, func() error {
		attemptCount++
		if attemptCount <= 2 {
			return retryableErr
		}
		return nil // Success on third attempt
	}, "test_operation")

	// Verify that the operation succeeded
	assert.NoError(t, err)

	// Verify that retry attempts were tracked
	mu.Lock()
	assert.Len(t, retryAttempts, 2, "Should have tracked 2 retry attempts")
	for i, attempt := range retryAttempts {
		assert.Equal(t, "test_operation", attempt.operation)
		assert.Equal(t, i+1, attempt.attempt)
		assert.Equal(t, retryableErr, attempt.err)
	}
	mu.Unlock()

	// Verify that operation metrics were updated
	metrics := testClient.GetOperationMetrics()
	operationMetrics, exists := metrics["test_operation"]
	assert.True(t, exists, "Operation metrics should exist")
	assert.Equal(t, int64(3), operationMetrics.TotalCalls, "Should have 3 total calls (1 initial + 2 retries)")
	assert.Equal(t, int64(2), operationMetrics.ErrorCount, "Should have 2 error calls")
	assert.Equal(t, int64(1), operationMetrics.SuccessCount, "Should have 1 success call")
	assert.Equal(t, int64(2), operationMetrics.RetryCount, "Should have 2 retry attempts")
}

// TestTrackRetryAttempt_WithKey tests that trackRetryAttempt works correctly with key tracking
func TestTrackRetryAttempt_WithKey(t *testing.T) {
	// Reset metrics to ensure clean state
	testClient.ResetOperationMetrics()

	// Use the existing testClient but reset its monitoring hooks
	originalHooks := testClient.monitoringHooks
	defer func() {
		testClient.SetMonitoringHooks(originalHooks)
	}()

	// Set up monitoring hooks to track operations
	var operationStarts []struct {
		operation string
		key       string
	}
	var operationEnds []struct {
		operation string
		key       string
		duration  time.Duration
		err       error
	}
	var retryAttempts []struct {
		operation string
		attempt   int
		err       error
	}
	var mu sync.Mutex

	hooks := &MonitoringHooks{
		OnOperationStart: func(operation string, key string) {
			mu.Lock()
			defer mu.Unlock()
			operationStarts = append(operationStarts, struct {
				operation string
				key       string
			}{operation, key})
		},
		OnOperationEnd: func(operation string, key string, duration time.Duration, err error) {
			mu.Lock()
			defer mu.Unlock()
			operationEnds = append(operationEnds, struct {
				operation string
				key       string
				duration  time.Duration
				err       error
			}{operation, key, duration, err})
		},
		OnRetryAttempt: func(operation string, attempt int, err error) {
			mu.Lock()
			defer mu.Unlock()
			retryAttempts = append(retryAttempts, struct {
				operation string
				attempt   int
				err       error
			}{operation, attempt, err})
		},
	}
	testClient.SetMonitoringHooks(hooks)

	// Test operation with key tracking that will retry
	ctx := context.Background()
	retryableErr := &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET} // Use a retryable network error
	testKey := "test-key-123"

	err := testClient.executeWithRetryAndKey(ctx, func() error {
		return retryableErr
	}, "test_operation_with_key", testKey)

	// Verify that the operation failed after retries
	assert.Error(t, err)

	// Verify that operation starts were tracked
	mu.Lock()
	assert.Greater(t, len(operationStarts), 0, "Should have tracked operation starts")
	for _, start := range operationStarts {
		assert.Equal(t, "test_operation_with_key", start.operation)
		assert.Equal(t, testKey, start.key)
	}
	mu.Unlock()

	// Verify that operation ends were tracked
	mu.Lock()
	assert.Greater(t, len(operationEnds), 0, "Should have tracked operation ends")
	for _, end := range operationEnds {
		assert.Equal(t, "test_operation_with_key", end.operation)
		assert.Equal(t, testKey, end.key)
		assert.Error(t, end.err)
	}
	mu.Unlock()

	// Verify that retry attempts were tracked
	mu.Lock()
	assert.Greater(t, len(retryAttempts), 0, "Should have tracked retry attempts")
	for i, attempt := range retryAttempts {
		assert.Equal(t, "test_operation_with_key", attempt.operation)
		assert.Equal(t, i+1, attempt.attempt)
		assert.Equal(t, retryableErr, attempt.err)
	}
	mu.Unlock()
}
