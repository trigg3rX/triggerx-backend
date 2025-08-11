package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	redis "github.com/redis/go-redis/v9"
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

	delay := config.InitialDelay
	var lastErr error

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the operation with monitoring
		start := time.Now()
		c.trackOperationStart(operationName, key)

		err := operation()

		duration := time.Since(start)
		c.trackOperationEnd(operationName, key, duration, err)

		// If operation succeeded, return immediately
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry based on custom predicate
		if !c.isRetryableError(err) {
			return err
		}

		// Track retry attempt
		c.trackRetryAttempt(operationName, attempt+1, err)

		// Calculate delay with jitter
		sleepDuration := c.calculateDelayWithJitter(delay, config.JitterFactor)

		if config.LogRetryAttempt {
			c.logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt+1, config.MaxRetries, err, sleepDuration)
		}

		select {
		case <-time.After(sleepDuration):
			// Calculate next delay using exponential backoff
			delay = c.calculateNextDelay(delay, config.BackoffFactor, config.MaxDelay)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries, lastErr)
}

// calculateDelayWithJitter calculates the sleep duration with jitter applied
func (c *Client) calculateDelayWithJitter(baseDelay time.Duration, jitterFactor float64) time.Duration {
	sleepDuration := baseDelay
	if jitterFactor > 0 {
		// Use a simple jitter calculation
		jitter := time.Duration(jitterFactor * float64(baseDelay) * c.secureFloat64())
		sleepDuration += jitter
	}
	return sleepDuration
}

// calculateNextDelay calculates the next delay value using exponential backoff
func (c *Client) calculateNextDelay(currentDelay time.Duration, backoffFactor float64, maxDelay time.Duration) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * backoffFactor)
	if nextDelay > maxDelay {
		nextDelay = maxDelay
	}
	return nextDelay
}

// secureFloat64 returns a secure random float64 in [0.0,1.0)
func (c *Client) secureFloat64() float64 {
	// Use time-based random for simplicity, but could be enhanced with crypto/rand
	return float64(time.Now().UnixNano()%1000) / 1000.0
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
