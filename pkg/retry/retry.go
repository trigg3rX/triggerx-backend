package retry

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Config holds the configuration for retry operations
type RetryConfig struct {
	MaxRetries      int              // Maximum number of retry attempts
	InitialDelay    time.Duration    // Initial delay between retries
	MaxDelay        time.Duration    // Maximum delay between retries
	BackoffFactor   float64          // Multiplier for exponential backoff
	JitterFactor    float64          // Factor for adding jitter to delays
	LogRetryAttempt bool             // Whether to log retry attempts
	StatusCodes     []int            // Status codes that should trigger a retry
	ShouldRetry     func(error) bool // Custom function to determine if error should be retried
}

// DefaultConfig returns a default configuration for retry operations
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      5,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
		StatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		ShouldRetry: nil,
	}
}

// validate checks the configuration for reasonable values
func (c *RetryConfig) validate() error {
	if c.MaxRetries < 0 {
		return errors.New("MaxRetries must be >= 0")
	}
	if c.InitialDelay <= 0 {
		return errors.New("InitialDelay must be positive")
	}
	if c.MaxDelay <= 0 {
		return errors.New("MaxDelay must be positive")
	}
	if c.BackoffFactor < 1.0 {
		return errors.New("BackoffFactor must be >= 1.0")
	}
	if c.JitterFactor < 0 || c.JitterFactor > 1.0 {
		return errors.New("JitterFactor must be between 0.0 and 1.0")
	}
	return nil
}

// secureFloat64 returns a secure random float64 in [0.0,1.0)
func secureFloat64() float64 {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// Fallback to time-based randomness if crypto fails
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(binary.BigEndian.Uint64(buf[:])) / float64(^uint64(0))
}

// Retry executes the given operation with exponential backoff and retry logic.
// Returns the result of the operation if successful, or an error if all attempts fail.
func Retry[T any](ctx context.Context, operation func() (T, error), retryConfig *RetryConfig, logger logging.Logger) (T, error) {
	var result T
	var err error

	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	} else {
		if err := retryConfig.validate(); err != nil {
			return result, fmt.Errorf("invalid retry config: %w", err)
		}
	}

	delay := retryConfig.InitialDelay

	for attempt := 1; attempt <= retryConfig.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		result, err = operation()
		if err == nil {
			return result, nil
		}

		// Check if we should retry based on custom predicate
		if retryConfig.ShouldRetry != nil && !retryConfig.ShouldRetry(err) {
			return result, err
		}

		if attempt == retryConfig.MaxRetries {
			break
		}

		if retryConfig.LogRetryAttempt {
			logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt, retryConfig.MaxRetries, err, delay)
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * retryConfig.BackoffFactor)

		// Add jitter to prevent thundering herd
		if retryConfig.JitterFactor > 0 {
			jitter := time.Duration(float64(nextDelay) * retryConfig.JitterFactor)
			nextDelay += time.Duration(float64(jitter) * (0.5 - secureFloat64()))
		}

		// Ensure delay doesn't exceed max delay
		if nextDelay > retryConfig.MaxDelay {
			nextDelay = retryConfig.MaxDelay
		}

		select {
		case <-time.After(delay):
			delay = nextDelay
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %w", retryConfig.MaxRetries, err)
}

// WithExponentialBackoff is a convenience function that uses default configuration
// with exponential backoff.
func WithExponentialBackoff[T any](ctx context.Context, operation func() (T, error), logger logging.Logger) (T, error) {
	return Retry(ctx, operation, DefaultRetryConfig(), logger)
}
