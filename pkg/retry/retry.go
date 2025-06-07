package retry

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Config holds the configuration for retry operations
type Config struct {
	MaxRetries      int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Multiplier for exponential backoff
	JitterFactor    float64       // Factor for adding jitter to delays
	LogRetryAttempt bool          // Whether to log retry attempts
}

// DefaultConfig returns a default configuration for retry operations
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:      5,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
	}
}

// validate checks the configuration for reasonable values
func (c *Config) validate() error {
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
		// In a retry context, falling back to a default value is better than panicking
		return 0.5
	}
	// Convert to uint64 and divide by max uint64 to get [0,1)
	return float64(binary.BigEndian.Uint64(buf[:])) / float64(^uint64(0))
}

// Retry executes the given operation with exponential backoff and retry logic.
// Returns the result of the operation if successful, or an error if all attempts fail.
func Retry[T any](ctx context.Context, operation func() (T, error), config *Config, logger logging.Logger) (T, error) {
	var result T
	var err error

	if config == nil {
		config = DefaultConfig()
	}

	if err := config.validate(); err != nil {
		return result, fmt.Errorf("invalid retry config: %w", err)
	}

	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt == config.MaxRetries {
			break
		}

		if config.LogRetryAttempt {
			logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt, config.MaxRetries, err, delay)
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * config.BackoffFactor)

		// Add jitter to prevent thundering herd
		if config.JitterFactor > 0 {
			jitter := time.Duration(float64(nextDelay) * config.JitterFactor)
			nextDelay += time.Duration(float64(jitter) * (0.5 - secureFloat64()))
		}

		// Ensure delay doesn't exceed max delay
		if nextDelay > config.MaxDelay {
			nextDelay = config.MaxDelay
		}

		select {
		case <-time.After(delay):
			delay = nextDelay
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries, err)
}

// WithExponentialBackoff is a convenience function that uses default configuration
// with exponential backoff.
func WithExponentialBackoff[T any](ctx context.Context, operation func() (T, error), logger logging.Logger) (T, error) {
	return Retry(ctx, operation, DefaultConfig(), logger)
}
