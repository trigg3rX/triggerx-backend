package retry

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Config holds the configuration for retry operations
type RetryConfig struct {
	MaxRetries      int                   // Maximum number of retry attempts
	InitialDelay    time.Duration         // Initial delay between retries
	MaxDelay        time.Duration         // Maximum delay between retries
	BackoffFactor   float64               // Multiplier for exponential backoff
	JitterFactor    float64               // Factor for adding jitter to delays (% of delay)
	LogRetryAttempt bool                  // Whether to log retry attempts
	ShouldRetry     func(error, int) bool // Custom function to determine if error should be retried (error, attempt number)
}

// DefaultConfig returns a default configuration for retry operations
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      5,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.2,
		LogRetryAttempt: true,
		ShouldRetry:     nil,
	}
}

// validate checks the configuration for reasonable values
func (c *RetryConfig) Validate() error {
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

// SecureFloat64 returns a secure random float64 in [0.0,1.0)
func SecureFloat64() float64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		// Fallback to math/rand if crypto/rand fails
		return mathrand.Float64()
	}
	return float64(binary.BigEndian.Uint64(b[:])) / (1 << 64)
}

// CalculateDelayWithJitter calculates the sleep duration for the given base delay with jitter applied
func CalculateDelayWithJitter(baseDelay time.Duration, jitterFactor float64) time.Duration {
	sleepDuration := baseDelay
	if jitterFactor > 0 {
		jitter := time.Duration(jitterFactor * float64(baseDelay) * SecureFloat64())
		sleepDuration += jitter
	}
	return sleepDuration
}

// CalculateNextDelay calculates the next delay value using exponential backoff
func CalculateNextDelay(currentDelay time.Duration, backoffFactor float64, maxDelay time.Duration) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * backoffFactor)
	if nextDelay > maxDelay {
		nextDelay = maxDelay
	}
	return nextDelay
}

// Retry executes the given operation with exponential backoff and retry logic.
// Returns the result of the operation if successful, or an error if all attempts fail.
func Retry[T any](ctx context.Context, operation func() (T, error), retryConfig *RetryConfig, logger logging.Logger) (T, error) {
	var zero T
	var err error

	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	} else if err := retryConfig.Validate(); err != nil {
		return zero, fmt.Errorf("invalid retry config: %w", err)
	}

	delay := retryConfig.InitialDelay

	for attempt := 0; attempt < retryConfig.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, opErr := operation()
		if opErr == nil {
			return result, nil
		}
		err = opErr

		// Check if we should retry based on custom predicate
		if retryConfig.ShouldRetry != nil && !retryConfig.ShouldRetry(err, attempt+1) {
			return zero, err
		}

		sleepDuration := CalculateDelayWithJitter(delay, retryConfig.JitterFactor)

		if retryConfig.LogRetryAttempt {
			logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt+1, retryConfig.MaxRetries, err, sleepDuration)
		}

		select {
		case <-time.After(sleepDuration):
			// Calculate next delay
			delay = CalculateNextDelay(delay, retryConfig.BackoffFactor, retryConfig.MaxDelay)
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts: %w", retryConfig.MaxRetries, err)
}

// RetryFunc executes an operation that only returns an error, with exponential backoff.
// This is a convenience wrapper around Retry.
func RetryFunc(ctx context.Context, operation func() error, config *RetryConfig, logger logging.Logger) error {
	opWithValue := func() (struct{}, error) {
		return struct{}{}, operation()
	}
	_, err := Retry(ctx, opWithValue, config, logger)
	return err
}
