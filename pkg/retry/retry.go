package retry

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Config holds the configuration for retry operations
type Config struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	LogRetryAttempt bool
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

// Retry executes the given operation with exponential backoff and retry logic
func Retry[T any](operation func() (T, error), config *Config, logger logging.Logger) (T, error) {
	var result T
	var err error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt < config.MaxRetries {
			if config.LogRetryAttempt {
				logger.Warnf("Attempt %d failed: %v. Retrying in %v...", attempt, err, delay)
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)

			// Add jitter to prevent thundering herd
			if config.JitterFactor > 0 {
				jitter := time.Duration(float64(delay) * config.JitterFactor)
				delay += time.Duration(float64(jitter) * (0.5 - rand.Float64()))
			}

			// Ensure delay doesn't exceed max delay
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			time.Sleep(delay)
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %v", config.MaxRetries, err)
}

// WithExponentialBackoff is a convenience function that uses default configuration
// with exponential backoff
func WithExponentialBackoff[T any](operation func() (T, error), logger logging.Logger) (T, error) {
	return Retry(operation, DefaultConfig(), logger)
}
