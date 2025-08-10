package redis

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// SetRetryConfig sets custom retry configuration
func (c *Client) SetRetryConfig(config *RetryConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryConfig = config
}

// SetConnectionRecoveryConfig sets custom connection recovery configuration
func (c *Client) SetConnectionRecoveryConfig(config *ConnectionRecoveryConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recoveryConfig = config
}

// executeWithRetry executes an operation with retry logic
func (c *Client) executeWithRetry(ctx context.Context, operation func() error, operationName string) error {
	return c.executeWithRetryAndKey(ctx, operation, operationName, "")
}

// executeWithRetryAndKey executes an operation with retry logic and tracks by key
func (c *Client) executeWithRetryAndKey(ctx context.Context, operation func() error, operationName string, key string) error {
	config := c.retryConfig
	if config == nil {
		config = DefaultRetryConfig()
	}

	// Convert Redis RetryConfig to generic RetryConfig
	retryConfig := &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		LogRetryAttempt: config.LogRetryAttempt,
		ShouldRetry:     c.isRetryableError,
	}

	// Create a wrapper operation that includes monitoring
	wrappedOperation := func() error {
		start := time.Now()
		c.trackOperationStart(operationName, key)

		err := operation()

		duration := time.Since(start)
		c.trackOperationEnd(operationName, key, duration, err)

		return err
	}

	// Use the generic retry package
	return retry.RetryFunc(ctx, wrappedOperation, retryConfig, c.logger)
}

// isRetryableError determines if an error is retryable using type assertions and specific checks
// instead of relying solely on string matching.
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Case 1: redis.Nil is a "not found" error, which is never retryable.
	if errors.Is(err, redis.Nil) {
		return false
	}

	// Case 2: The context was canceled or its deadline was exceeded. Not retryable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Case 3: Check for common network error types.
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			c.logger.Debugf("Redis error is a retryable network timeout: %v", err)
			return true // It's a timeout, so we should retry.
		}
	}

	// Case 4: Check for specific system-level errors (e.g., connection refused).
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			if sysErr == syscall.ECONNREFUSED || sysErr == syscall.ECONNRESET {
				c.logger.Debugf("Redis error is a retryable syscall error (%s): %v", sysErr.Error(), err)
				return true
			}
		}
	}

	// Case 5: EOF often indicates a connection was closed by the other side, which is retryable.
	if errors.Is(err, io.EOF) {
		c.logger.Debugf("Redis error is a retryable EOF: %v", err)
		return true
	}

	// Default to non-retryable for unknown errors.
	c.logger.Debugf("Redis error is considered non-retryable by default: %v (type: %T)", err, err)
	return false
}
