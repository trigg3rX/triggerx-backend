package redis

import (
	"context"
	"fmt"
	"math"
	"math/rand"
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

	start := time.Now()
	c.trackOperationStart(operationName, key)
	defer func() {
		duration := time.Since(start)
		c.trackOperationEnd(operationName, key, duration, nil)
	}()

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateBackoffDelay(attempt, config)
			if config.LogRetryAttempt && c.logger != nil {
				c.logger.Warnf("Retrying Redis operation %s (attempt %d/%d) after %v",
					operationName, attempt, config.MaxRetries, delay)
			}

			// Track retry attempt
			c.trackRetryAttempt(operationName, attempt, lastErr)

			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if err := operation(); err != nil {
			lastErr = err
			if !c.isRetryableError(err) {
				// if c.logger != nil {
					// c.logger.Errorf("Non-retryable error in Redis operation %s: %v (type: %T)", operationName, err, err)
				// }
				return err
			}
			continue
		}

		return nil
	}

	if c.logger != nil {
		c.logger.Errorf("Redis operation %s failed after %d retries: %v",
			operationName, config.MaxRetries, lastErr)
	}
	return fmt.Errorf("operation %s failed after %d retries: %w", operationName, config.MaxRetries, lastErr)
}

// calculateBackoffDelay calculates exponential backoff delay with jitter
func (c *Client) calculateBackoffDelay(attempt int, config *RetryConfig) time.Duration {
	backoff := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

	// Apply jitter
	jitter := backoff * config.JitterFactor * (rand.Float64() - 0.5)
	delay := time.Duration(backoff + jitter)

	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}

// isRetryableError determines if an error is retryable
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for redis.Nil error (non-retryable)
	if err == redis.Nil {
		// if c.logger != nil {
			// c.logger.Debugf("Redis error is redis.Nil (non-retryable): %v", err)
		// }
		return false
	}

	// Check for redis: nil string error (non-retryable)
	if err.Error() == "redis: nil" {
		if c.logger != nil {
			c.logger.Debugf("Redis error is 'redis: nil' string (non-retryable): %v", err)
		}
		return false
	}

	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset by peer",
		"i/o timeout",
		"timeout",
		"network is unreachable",
		"broken pipe",
		"connection aborted",
		"connection timed out",
		"temporary failure",
		"server overloaded",
		"too many connections",
		"connection lost",
		"connection closed",
		"no route to host",
		"operation timed out",
		"network unreachable",
		"connection reset",
		"host not found",
		"EOF",
	}

	for _, retryableErr := range retryableErrors {
		if errStr == retryableErr {
			if c.logger != nil {
				c.logger.Debugf("Redis error is retryable: %v", err)
			}
			return true
		}
	}

	if c.logger != nil {
		c.logger.Debugf("Redis error is non-retryable (default): %v (type: %T)", err, err)
	}
	return false
}
