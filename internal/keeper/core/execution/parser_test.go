package execution

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestLogger is a mock implementation of the logger interface for parser tests
type TestLogger struct {
	mock.Mock
}

func (m *TestLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestLogger) Debugf(format string, args ...interface{}) {
	// Just ignore all debug calls for testing
}

func (m *TestLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *TestLogger) Warnf(format string, args ...interface{}) {
	// Just ignore all warning calls for testing
}

func (m *TestLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *TestLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *TestLogger) With(tags ...any) logging.Logger {
	return m
}

func TestParserDynamicArgs(t *testing.T) {
	t.Run("JSON parsing", func(t *testing.T) {
		logger := new(TestLogger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with valid JSON array
		output := `[123.45, "test", true]`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 3)
		assert.Equal(t, 123.45, args[0])
		assert.Equal(t, "test", args[1])
		assert.Equal(t, true, args[2])
	})

	t.Run("Response pattern", func(t *testing.T) {
		logger := new(TestLogger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output containing Response: lines
		output := `
Container Log: START_EXECUTION
Container Log: Condition satisfied: true
Container Log: Timestamp: 2025-07-18T12:31:08Z
Container Log: Response: 3603.66
Container Log: Response: 3603.66
Container Log: Ethereum price is greater than 0
Code execution completed in: 42.284769713s
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 2)
		assert.Equal(t, 3603.66, args[0])
		assert.Equal(t, 3603.66, args[1])
	})

	t.Run("Condition pattern", func(t *testing.T) {
		logger := new(TestLogger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output containing only condition satisfied
		output := `
Container Log: START_EXECUTION
Container Log: Condition satisfied: true
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 1)
		assert.Equal(t, true, args[0])
	})

	t.Run("Fallback", func(t *testing.T) {
		logger := new(TestLogger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output that doesn't match any pattern
		output := `
Container Log: START_EXECUTION
Container Log: Some random output
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 1)
		assert.Equal(t, "0", args[0])
	})
}

func TestRealWorldExample(t *testing.T) {
	logger := new(TestLogger)

	executor := &TaskExecutor{
		logger: logger,
	}

	// This is the exact output from the error logs
	output := `Container Log: go: creating new go.mod: module code
Container Log: go: to add module requirements and sums:
Container Log: 	go mod tidy
Code execution started
Container Log: START_EXECUTION
Container Log: Condition satisfied: true
Container Log: Timestamp: 2025-07-18T12:31:08Z
Container Log: Response: 3603.66
Container Log: Response: 3603.66
Container Log: Ethereum price is greater than 0
Code execution completed in: 42.284769713s
Container Log: END_EXECUTION`

	args := executor.parseDynamicArgs(output)

	assert.Len(t, args, 2)
	assert.Equal(t, 3603.66, args[0])
	assert.Equal(t, 3603.66, args[1])
}
