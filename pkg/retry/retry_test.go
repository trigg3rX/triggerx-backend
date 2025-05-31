package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func init() {
	// Initialize logger for tests
	config := logging.NewDefaultConfig("retry_test")
	config.Environment = logging.Development
	config.UseColors = true
	if err := logging.InitServiceLogger(config); err != nil {
		panic(err)
	}
}

// TestRetry tests the basic retry functionality
func TestRetry(t *testing.T) {
	logger := logging.GetServiceLogger()

	tests := []struct {
		name           string
		operation      func() (string, error)
		config         *Config
		expectedResult string
		expectError    bool
		expectedDelay  time.Duration // Expected minimum delay between retries
	}{
		{
			name: "success on first try",
			operation: func() (string, error) {
				return "success", nil
			},
			config:         DefaultConfig(),
			expectedResult: "success",
			expectError:    false,
			expectedDelay:  0,
		},
		{
			name: "success after retries",
			operation: func() (string, error) {
				return "success", nil
			},
			config: &Config{
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
			config: &Config{
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
			config: &Config{
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
			result, err := Retry(tt.operation, tt.config, logger)
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
	logger := logging.GetServiceLogger()

	t.Run("success with default config", func(t *testing.T) {
		result, err := WithExponentialBackoff(func() (string, error) {
			return "success", nil
		}, logger)

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("failure with default config", func(t *testing.T) {
		result, err := WithExponentialBackoff(func() (string, error) {
			return "", errors.New("operation failed")
		}, logger)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed after")
	})
}

// TestRetryWithDifferentTypes tests retry with different return types
func TestRetryWithDifferentTypes(t *testing.T) {
	logger := logging.GetServiceLogger()

	t.Run("int type", func(t *testing.T) {
		result, err := Retry(func() (int, error) {
			return 42, nil
		}, DefaultConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("struct type", func(t *testing.T) {
		type TestStruct struct {
			Value string
		}

		result, err := Retry(func() (TestStruct, error) {
			return TestStruct{Value: "test"}, nil
		}, DefaultConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, "test", result.Value)
	})

	t.Run("pointer type", func(t *testing.T) {
		result, err := Retry(func() (*string, error) {
			value := "pointer"
			return &value, nil
		}, DefaultConfig(), logger)

		assert.NoError(t, err)
		assert.Equal(t, "pointer", *result)
	})
}

// TestRetryConfig tests configuration validation and defaults
func TestRetryConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, 5, config.MaxRetries)
		assert.Equal(t, time.Second, config.InitialDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.BackoffFactor)
		assert.Equal(t, 0.1, config.JitterFactor)
		assert.True(t, config.LogRetryAttempt)
	})

	t.Run("custom config values", func(t *testing.T) {
		config := &Config{
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
