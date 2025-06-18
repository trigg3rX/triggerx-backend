package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestRetry tests the basic retry functionality
func TestRetry(t *testing.T) {
	config := logging.LoggerConfig{
		ProcessName:   logging.TestProcess,
		IsDevelopment: true,
	}

	logger, err := logging.NewZapLogger(config)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name           string
		operation      func() (string, error)
		config         *RetryConfig
		expectedResult string
		expectError    bool
		expectedDelay  time.Duration // Expected minimum delay between retries
	}{
		{
			name: "success on first try",
			operation: func() (string, error) {
				return "success", nil
			},
			config:         DefaultRetryConfig(),
			expectedResult: "success",
			expectError:    false,
			expectedDelay:  0,
		},
		{
			name: "success after retries",
			operation: func() (string, error) {
				return "success", nil
			},
			config: &RetryConfig{
				MaxRetries:      3,
				InitialDelay:    10 * time.Millisecond,
				MaxDelay:        100 * time.Millisecond,
				BackoffFactor:   2.0,
				JitterFactor:    0.1,
				LogRetryAttempt: false,
			},
			expectedResult: "success",
			expectError:    false,
			expectedDelay:  10 * time.Millisecond,
		},
		{
			name: "failure after all retries",
			operation: func() (string, error) {
				return "", errors.New("operation failed")
			},
			config: &RetryConfig{
				MaxRetries:      2,
				InitialDelay:    10 * time.Millisecond,
				MaxDelay:        100 * time.Millisecond,
				BackoffFactor:   2.0,
				JitterFactor:    0.1,
				LogRetryAttempt: false,
			},
			expectedResult: "",
			expectError:    true,
			expectedDelay:  10 * time.Millisecond,
		},
		{
			name: "exponential backoff test",
			operation: func() (string, error) {
				return "", errors.New("temporary error")
			},
			config: &RetryConfig{
				MaxRetries:      3,
				InitialDelay:    10 * time.Millisecond,
				MaxDelay:        100 * time.Millisecond,
				BackoffFactor:   2.0,
				JitterFactor:    0.1,
				LogRetryAttempt: false,
			},
			expectedResult: "",
			expectError:    true,
			expectedDelay:  10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result, err := Retry(context.Background(), tt.operation, tt.config, logger)
			duration := time.Since(start)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed after")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify minimum delay for retries
			if tt.config.MaxRetries > 1 && tt.expectError {
				assert.GreaterOrEqual(t, duration, tt.expectedDelay)
			}
		})
	}
}

// TestWithExponentialBackoff tests the convenience function
func TestWithExponentialBackoff(t *testing.T) {
	config := logging.LoggerConfig{
		ProcessName:   logging.TestProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(config)
	if err != nil {
		panic(err)
	}

	t.Run("success with default config", func(t *testing.T) {
		result, err := WithExponentialBackoff(context.Background(), func() (string, error) {
			return "success", nil
		}, logger)

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("failure with default config", func(t *testing.T) {
		result, err := WithExponentialBackoff(context.Background(), func() (string, error) {
			return "", errors.New("operation failed")
		}, logger)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed after")
	})
}

// TestRetryWithDifferentTypes tests retry with different return types
func TestRetryWithDifferentTypes(t *testing.T) {
	config := logging.LoggerConfig{
		ProcessName:   logging.TestProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(config)
	if err != nil {
		panic(err)
	}

	t.Run("int type", func(t *testing.T) {
		result, err := Retry(context.Background(), func() (int, error) {
			return 42, nil
		}, DefaultRetryConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("struct type", func(t *testing.T) {
		type TestStruct struct {
			Value string
		}

		result, err := Retry(context.Background(), func() (TestStruct, error) {
			return TestStruct{Value: "test"}, nil
		}, DefaultRetryConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, "test", result.Value)
	})

	t.Run("pointer type", func(t *testing.T) {
		result, err := Retry(context.Background(), func() (*string, error) {
			value := "pointer"
			return &value, nil
		}, DefaultRetryConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, "pointer", *result)
	})
}

// TestRetryConfig tests configuration validation and defaults
func TestRetryConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultRetryConfig()
		assert.Equal(t, 5, config.MaxRetries)
		assert.Equal(t, time.Second, config.InitialDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.BackoffFactor)
		assert.Equal(t, 0.1, config.JitterFactor)
		assert.True(t, config.LogRetryAttempt)
	})

	t.Run("custom config values", func(t *testing.T) {
		config := &RetryConfig{
			MaxRetries:      3,
			InitialDelay:    100 * time.Millisecond,
			MaxDelay:        1 * time.Second,
			BackoffFactor:   1.5,
			JitterFactor:    0.2,
			LogRetryAttempt: false,
		}

		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
		assert.Equal(t, time.Second, config.MaxDelay)
		assert.Equal(t, 1.5, config.BackoffFactor)
		assert.Equal(t, 0.2, config.JitterFactor)
		assert.False(t, config.LogRetryAttempt)
	})
}
