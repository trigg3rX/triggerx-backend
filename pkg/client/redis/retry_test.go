package redis

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteWithRetry_SuccessOnFirstAttempt verifies the operation is not retried if it succeeds.
func TestExecuteWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	var attempts int
	operation := func() error {
		attempts++
		return nil
	}

	err := testClient.executeWithRetry(context.Background(), operation, "test_success")
	require.NoError(t, err)
	assert.Equal(t, 1, attempts, "Operation should have been attempted exactly once")
}

// TestExecuteWithRetry_FailureWithNonRetryableError verifies that it fails immediately on non-retryable errors.
func TestExecuteWithRetry_FailureWithNonRetryableError(t *testing.T) {
	var attempts int
	// redis.Nil is a classic non-retryable error.
	nonRetryableErr := redis.Nil
	operation := func() error {
		attempts++
		return nonRetryableErr
	}

	err := testClient.executeWithRetry(context.Background(), operation, "test_non_retryable")
	require.Error(t, err)
	assert.ErrorIs(t, err, nonRetryableErr)
	assert.Equal(t, 1, attempts, "Operation should have been attempted exactly once")
}

// TestExecuteWithRetry_SuccessAfterRetries verifies that the client retries on retryable errors and eventually succeeds.
func TestExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	originalConfig := *testClient.retryConfig
	// Use a cleanup function to restore the original config after the test.
	t.Cleanup(func() {
		testClient.SetRetryConfig(&originalConfig)
	})

	// Configure a fast retry for the test.
	testClient.SetRetryConfig(&RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      5 * time.Millisecond,
		BackoffFactor: 1.5,
	})

	var attempts int
	retryableErr := io.EOF // io.EOF is a good example of a retryable network error.
	operation := func() error {
		attempts++
		if attempts <= 2 {
			return retryableErr // Fail the first two times
		}
		return nil // Succeed on the third attempt
	}

	err := testClient.executeWithRetry(context.Background(), operation, "test_success_after_retry")
	require.NoError(t, err)
	assert.Equal(t, 3, attempts, "Operation should have succeeded on the third attempt")
}

// TestExecuteWithRetry_FailureAfterAllAttempts verifies that the function fails after exhausting all retries.
func TestExecuteWithRetry_FailureAfterAllAttempts(t *testing.T) {
	originalConfig := *testClient.retryConfig
	t.Cleanup(func() {
		testClient.SetRetryConfig(&originalConfig)
	})

	// Configure a fast retry with 2 max retries.
	testClient.SetRetryConfig(&RetryConfig{
		MaxRetries:    2, // This means 2 total attempts (attempt 0 and attempt 1)
		InitialDelay:  1 * time.Millisecond,
		BackoffFactor: 1.5,
		MaxDelay:      5 * time.Millisecond,
	})

	var attempts int
	retryableErr := &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}
	operation := func() error {
		attempts++
		return retryableErr // Always fail
	}

	err := testClient.executeWithRetry(context.Background(), operation, "test_failure_after_retries")
	require.Error(t, err)
	assert.ErrorIs(t, err, syscall.ECONNRESET)
	assert.Equal(t, 2, attempts, "Should have been attempted 2 times (MaxRetries: 2)")
}

// TestIsRetryableError uses a table-driven test to check all conditions of the error-checking function.
func TestIsRetryableError(t *testing.T) {
	// Create specific error types for testing
	netTimeoutErr := &net.DNSError{Err: "i/o timeout", IsTimeout: true}
	connRefusedErr := &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "redis.Nil error is not retryable",
			err:      redis.Nil,
			expected: false,
		},
		{
			name:     "context.Canceled is not retryable",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "context.DeadlineExceeded is not retryable",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "generic error is not retryable",
			err:      errors.New("some unknown application error"),
			expected: false,
		},
		{
			name:     "network timeout error is retryable",
			err:      netTimeoutErr,
			expected: true,
		},
		{
			name:     "syscall.ECONNREFUSED is retryable",
			err:      connRefusedErr,
			expected: true,
		},
		{
			name:     "syscall.ECONNRESET is retryable",
			err:      &net.OpError{Err: syscall.ECONNRESET},
			expected: true,
		},
		{
			name:     "io.EOF is retryable",
			err:      io.EOF,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := testClient.isRetryableError(tc.err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
