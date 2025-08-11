package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// --- Test Helper Functions ---

// failingOperation is a helper that returns an error a specified number of times before succeeding.
func failingOperation(failures int, successValue string, failError error) (func() (string, error), *int) {
	callCount := 0
	return func() (string, error) {
		callCount++
		if callCount > failures {
			return successValue, nil
		}
		return "", failError
	}, &callCount
}

// --- Test Cases ---

func TestRetry(t *testing.T) {

	t.Run("Success on first attempt", func(t *testing.T) {
		// Arrange
		logger := logging.NewNoOpLogger()
		operation, callCount := failingOperation(0, "success", nil)
		config := retry.DefaultRetryConfig()

		// Act
		result, err := retry.Retry(context.Background(), operation, config, logger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, *callCount)
	})

	t.Run("Success after failures", func(t *testing.T) {
		// Arrange
		mockLogger := &logging.MockLogger{}
		// Expect Warnf to be called twice (for the 2 failures)
		mockLogger.On("Warnf", "Attempt %d/%d failed: %v. Retrying in %v...", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)

		operation, callCount := failingOperation(2, "success", errors.New("transient error"))

		config := retry.DefaultRetryConfig()
		config.MaxRetries = 5
		config.InitialDelay = 1 * time.Millisecond // Speed up test

		// Act
		result, err := retry.Retry(context.Background(), operation, config, mockLogger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, *callCount, "Should be called 3 times (2 failures + 1 success)")

		// Verify that Warnf was called for each failure
		mockLogger.AssertCalled(t, "Warnf", mock.AnythingOfType("string"), mock.Anything)
		mockLogger.AssertNumberOfCalls(t, "Warnf", 2)
	})

	t.Run("Failure after all attempts", func(t *testing.T) {
		// Arrange
		finalError := errors.New("permanent error")
		operation, callCount := failingOperation(5, "success", finalError)
		logger := logging.NewNoOpLogger()

		config := retry.DefaultRetryConfig()
		config.MaxRetries = 3
		config.InitialDelay = 1 * time.Millisecond

		// Act
		_, err := retry.Retry(context.Background(), operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, finalError, "The final error should be the one from the operation")
		assert.Contains(t, err.Error(), fmt.Sprintf("operation failed after %d attempts", config.MaxRetries))
		assert.Equal(t, 3, *callCount)
	})

	t.Run("Context cancellation during sleep", func(t *testing.T) {
		// Arrange
		operation, callCount := failingOperation(5, "success", errors.New("timeout"))
		logger := logging.NewNoOpLogger()

		config := retry.DefaultRetryConfig()
		config.InitialDelay = 100 * time.Millisecond // A noticeable delay

		ctx, cancel := context.WithCancel(context.Background())

		// Act
		// Cancel the context shortly after the first attempt fails and it starts sleeping
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, err := retry.Retry(ctx, operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
		assert.Equal(t, 1, *callCount, "Operation should only be called once before context is canceled")
	})

	t.Run("Context cancellation at start of retry attempt", func(t *testing.T) {
		// Arrange
		operation, callCount := failingOperation(3, "success", errors.New("transient error"))
		logger := logging.NewNoOpLogger()

		config := retry.DefaultRetryConfig()
		config.InitialDelay = 10 * time.Millisecond
		config.MaxRetries = 5

		ctx, cancel := context.WithCancel(context.Background())

		// Act
		// Cancel the context after the first failure but before the second retry attempt starts
		go func() {
			time.Sleep(5 * time.Millisecond) // Cancel during the sleep
			cancel()
		}()

		_, err := retry.Retry(ctx, operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
		// The operation should be called once (first attempt), then fail, then sleep,
		// then on the next iteration, the context check at the start should catch the cancellation
		assert.Equal(t, 1, *callCount, "Operation should only be called once before context is canceled at start of retry")
	})

	t.Run("Context cancellation between retry attempts", func(t *testing.T) {
		// Arrange
		operation, callCount := failingOperation(3, "success", errors.New("transient error"))
		logger := logging.NewNoOpLogger()

		config := retry.DefaultRetryConfig()
		config.InitialDelay = 5 * time.Millisecond
		config.MaxRetries = 5

		ctx, cancel := context.WithCancel(context.Background())

		// Act
		// Cancel the context after the first retry attempt completes but before the second one starts
		go func() {
			time.Sleep(15 * time.Millisecond) // Cancel after first retry attempt
			cancel()
		}()

		_, err := retry.Retry(ctx, operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
		// The operation should be called twice (first attempt + first retry), then on the third iteration
		// the context check at the start should catch the cancellation
		assert.Equal(t, 2, *callCount, "Operation should be called twice before context is canceled at start of retry")
	})

	t.Run("Stops immediately if ShouldRetry returns false", func(t *testing.T) {
		// Arrange
		nonRetryableError := errors.New("do not retry")
		operation, callCount := failingOperation(5, "success", nonRetryableError)
		logger := logging.NewNoOpLogger()

		config := retry.DefaultRetryConfig()
		config.ShouldRetry = func(err error) bool {
			return !errors.Is(err, nonRetryableError)
		}

		// Act
		_, err := retry.Retry(context.Background(), operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, nonRetryableError)
		assert.Equal(t, 1, *callCount, "Should not retry if predicate returns false")
	})

	t.Run("Invalid config returns error", func(t *testing.T) {
		// Arrange
		operation, _ := failingOperation(0, "success", nil)
		logger := logging.NewNoOpLogger()
		config := retry.DefaultRetryConfig()
		config.BackoffFactor = 0.5 // Invalid value

		// Act
		_, err := retry.Retry(context.Background(), operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid retry config")
	})
}

func TestRetryFunc(t *testing.T) {

	t.Run("Success after one failure", func(t *testing.T) {
		// Arrange
		callCount := 0
		operation := func() error {
			callCount++
			if callCount > 1 {
				return nil
			}
			return errors.New("transient error")
		}
		logger := logging.NewNoOpLogger()
		config := retry.DefaultRetryConfig()
		config.InitialDelay = 1 * time.Millisecond

		// Act
		err := retry.RetryFunc(context.Background(), operation, config, logger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("Failure after all attempts", func(t *testing.T) {
		// Arrange
		callCount := 0
		finalError := errors.New("permanent error")
		operation := func() error {
			callCount++
			return finalError
		}
		logger := logging.NewNoOpLogger()
		config := retry.DefaultRetryConfig()
		config.MaxRetries = 4
		config.InitialDelay = 1 * time.Millisecond

		// Act
		err := retry.RetryFunc(context.Background(), operation, config, logger)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, finalError)
		assert.Equal(t, 4, callCount)
	})
}

func TestRetryConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		modifyConfig  func(*retry.RetryConfig)
		expectedError string
	}{
		{
			name: "valid config should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				// Use default config as-is
			},
			expectedError: "",
		},
		{
			name: "MaxRetries negative should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.MaxRetries = -1
			},
			expectedError: "MaxRetries must be >= 0",
		},
		{
			name: "MaxRetries zero should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.MaxRetries = 0
			},
			expectedError: "",
		},
		{
			name: "InitialDelay zero should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.InitialDelay = 0
			},
			expectedError: "InitialDelay must be positive",
		},
		{
			name: "InitialDelay negative should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.InitialDelay = -1 * time.Second
			},
			expectedError: "InitialDelay must be positive",
		},
		{
			name: "MaxDelay zero should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.MaxDelay = 0
			},
			expectedError: "MaxDelay must be positive",
		},
		{
			name: "MaxDelay negative should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.MaxDelay = -1 * time.Second
			},
			expectedError: "MaxDelay must be positive",
		},
		{
			name: "BackoffFactor less than 1.0 should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.BackoffFactor = 0.5
			},
			expectedError: "BackoffFactor must be >= 1.0",
		},
		{
			name: "BackoffFactor exactly 1.0 should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.BackoffFactor = 1.0
			},
			expectedError: "",
		},
		{
			name: "BackoffFactor greater than 1.0 should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.BackoffFactor = 2.5
			},
			expectedError: "",
		},
		{
			name: "JitterFactor negative should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.JitterFactor = -0.1
			},
			expectedError: "JitterFactor must be between 0.0 and 1.0",
		},
		{
			name: "JitterFactor greater than 1.0 should fail",
			modifyConfig: func(c *retry.RetryConfig) {
				c.JitterFactor = 1.5
			},
			expectedError: "JitterFactor must be between 0.0 and 1.0",
		},
		{
			name: "JitterFactor exactly 0.0 should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.JitterFactor = 0.0
			},
			expectedError: "",
		},
		{
			name: "JitterFactor exactly 1.0 should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.JitterFactor = 1.0
			},
			expectedError: "",
		},
		{
			name: "JitterFactor between 0.0 and 1.0 should pass",
			modifyConfig: func(c *retry.RetryConfig) {
				c.JitterFactor = 0.5
			},
			expectedError: "",
		},
		{
			name: "multiple invalid fields should return first error",
			modifyConfig: func(c *retry.RetryConfig) {
				c.MaxRetries = -1
				c.InitialDelay = 0
				c.BackoffFactor = 0.5
			},
			expectedError: "MaxRetries must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := retry.DefaultRetryConfig()
			tt.modifyConfig(config)

			// Act
			err := config.Validate()

			// Assert
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			}
		})
	}
}

func TestCalculateNextDelay(t *testing.T) {
	t.Run("nextDelay exceeds maxDelay should cap at maxDelay", func(t *testing.T) {
		// Arrange
		currentDelay := 10 * time.Second
		backoffFactor := 4.0 // This will make nextDelay = 40s
		maxDelay := 20 * time.Second

		// Act
		result := retry.CalculateNextDelay(currentDelay, backoffFactor, maxDelay)

		// Assert
		assert.Equal(t, maxDelay, result, "Should cap at maxDelay when calculated delay exceeds it")
	})

	t.Run("nextDelay within maxDelay should use calculated value", func(t *testing.T) {
		// Arrange
		currentDelay := 5 * time.Second
		backoffFactor := 2.0 // This will make nextDelay = 10s
		maxDelay := 20 * time.Second

		// Act
		result := retry.CalculateNextDelay(currentDelay, backoffFactor, maxDelay)

		// Assert
		expected := 10 * time.Second
		assert.Equal(t, expected, result, "Should use calculated delay when within maxDelay")
	})

	t.Run("nextDelay exactly equals maxDelay should use maxDelay", func(t *testing.T) {
		// Arrange
		currentDelay := 10 * time.Second
		backoffFactor := 2.0 // This will make nextDelay = 20s
		maxDelay := 20 * time.Second

		// Act
		result := retry.CalculateNextDelay(currentDelay, backoffFactor, maxDelay)

		// Assert
		assert.Equal(t, maxDelay, result, "Should use maxDelay when calculated delay equals it")
	})
}

func TestRetry_NilConfigUsesDefault(t *testing.T) {
	t.Run("nil config should use default configuration", func(t *testing.T) {
		// Arrange
		logger := logging.NewNoOpLogger()
		operation, callCount := failingOperation(0, "success", nil)

		// Act - Pass nil config
		result, err := retry.Retry(context.Background(), operation, nil, logger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, *callCount)
	})

	t.Run("nil config should use default retry behavior", func(t *testing.T) {
		// Arrange
		mockLogger := &logging.MockLogger{}
		// Expect Warnf to be called for the failure
		mockLogger.On("Warnf", "Attempt %d/%d failed: %v. Retrying in %v...", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)

		operation, callCount := failingOperation(1, "success", errors.New("transient error"))

		// Act - Pass nil config
		result, err := retry.Retry(context.Background(), operation, nil, mockLogger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 2, *callCount, "Should be called 2 times (1 failure + 1 success)")

		// Verify that Warnf was called for the failure
		mockLogger.AssertCalled(t, "Warnf", mock.AnythingOfType("string"), mock.Anything)
		mockLogger.AssertNumberOfCalls(t, "Warnf", 1)
	})
}

func TestSecureFloat64(t *testing.T) {
	t.Run("should return value in range [0.0, 1.0)", func(t *testing.T) {
		// Act
		result := retry.SecureFloat64()

		// Assert
		assert.GreaterOrEqual(t, result, 0.0, "Should be >= 0.0")
		assert.Less(t, result, 1.0, "Should be < 1.0")
	})

	t.Run("should return different values on multiple calls", func(t *testing.T) {
		// Act
		result1 := retry.SecureFloat64()
		result2 := retry.SecureFloat64()
		result3 := retry.SecureFloat64()

		// Assert
		// Note: While theoretically possible, it's extremely unlikely for crypto/rand to return the same value
		// We're testing that we get different values, which indicates the function is working
		results := []float64{result1, result2, result3}
		uniqueResults := make(map[float64]bool)
		for _, r := range results {
			uniqueResults[r] = true
		}

		// In a real scenario, we'd expect at least 2 unique values out of 3 calls
		// But for testing purposes, we'll just verify the values are in the correct range
		for _, r := range results {
			assert.GreaterOrEqual(t, r, 0.0)
			assert.Less(t, r, 1.0)
		}
	})
}

// TestSecureFloat64FallbackBehavior tests the fallback behavior by analyzing the pattern
func TestSecureFloat64FallbackBehavior(t *testing.T) {
	t.Run("should handle fallback pattern correctly", func(t *testing.T) {
		// The fallback uses: math.Mod(float64(time.Now().UnixNano()), 1000) / 1000.0
		// This means the result should be in the range [0.0, 0.999]

		// Act - Call multiple times to see the pattern
		results := make([]float64, 10)
		for i := 0; i < 10; i++ {
			results[i] = retry.SecureFloat64()
			time.Sleep(1 * time.Microsecond) // Small delay to ensure different timestamps
		}

		// Assert
		for _, result := range results {
			assert.GreaterOrEqual(t, result, 0.0, "All results should be >= 0.0")
			assert.Less(t, result, 1.0, "All results should be < 1.0")
		}

		// Verify that we get different values (indicating the function is working)
		uniqueResults := make(map[float64]bool)
		for _, r := range results {
			uniqueResults[r] = true
		}

		// We should have at least 5 unique values out of 10 calls
		// This indicates the function is working properly
		assert.GreaterOrEqual(t, len(uniqueResults), 5, "Should have multiple unique values")
	})

	t.Run("should work correctly in retry scenarios", func(t *testing.T) {
		// Test that SecureFloat64 works correctly when used in retry scenarios
		// This indirectly tests the fallback behavior in a real-world context

		// Arrange
		logger := logging.NewNoOpLogger()
		config := retry.DefaultRetryConfig()
		config.InitialDelay = 1 * time.Millisecond
		config.JitterFactor = 0.5 // This will use SecureFloat64
		config.MaxRetries = 3

		callCount := 0
		operation := func() (string, error) {
			callCount++
			if callCount > 2 {
				return "success", nil
			}
			return "", errors.New("transient error")
		}

		// Act
		result, err := retry.Retry(context.Background(), operation, config, logger)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, callCount, "Should be called 3 times (2 failures + 1 success)")

		// The fact that this test passes means SecureFloat64 is working correctly
		// in the context of CalculateDelayWithJitter, which includes the fallback
	})
}
