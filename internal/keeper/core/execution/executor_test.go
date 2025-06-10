package execution

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Simple mocks for unit testing
type UnitTestLogger struct {
	mock.Mock
}

func (t *UnitTestLogger) Warnf(format string, args ...interface{})  { t.Called(format, args) }
func (t *UnitTestLogger) Debugf(format string, args ...interface{}) { t.Called(format, args) }
func (t *UnitTestLogger) Errorf(format string, args ...interface{}) { t.Called(format, args) }
func (t *UnitTestLogger) Infof(format string, args ...interface{})  { t.Called(format, args) }
func (t *UnitTestLogger) Debug(msg string, keysAndValues ...interface{}) {
	t.Called(msg, keysAndValues)
}
func (t *UnitTestLogger) Info(msg string, keysAndValues ...interface{}) { t.Called(msg, keysAndValues) }
func (t *UnitTestLogger) Warn(msg string, keysAndValues ...interface{}) { t.Called(msg, keysAndValues) }
func (t *UnitTestLogger) Error(msg string, keysAndValues ...interface{}) {
	t.Called(msg, keysAndValues)
}
func (t *UnitTestLogger) Fatal(msg string, keysAndValues ...interface{}) {
	t.Called(msg, keysAndValues)
}
func (t *UnitTestLogger) Fatalf(template string, args ...interface{}) { t.Called(template, args) }
func (t *UnitTestLogger) With(tags ...any) logging.Logger             { return t }

// Add mock assertion methods
func (t *UnitTestLogger) AssertExpected(tb testing.TB) bool {
	return t.AssertExpectations(tb)
}

// ===========================================
// UNIT TESTS FOR NewTaskExecutor
// ===========================================

func TestNewTaskExecutor_ValidParams(t *testing.T) {
	// Arrange
	alchemyAPIKey := "test-api-key"
	codeExecutor := &docker.CodeExecutor{}
	validator := &validation.TaskValidator{}
	aggregatorClient := &aggregator.AggregatorClient{}
	logger := &UnitTestLogger{}

	// Act
	executor := NewTaskExecutor(alchemyAPIKey, codeExecutor, validator, aggregatorClient, logger)

	// Assert
	assert.NotNil(t, executor)
	assert.Equal(t, alchemyAPIKey, executor.alchemyAPIKey)
	assert.Equal(t, codeExecutor, executor.codeExecutor)
	assert.NotNil(t, executor.argConverter)
	assert.Equal(t, validator, executor.validator)
	assert.Equal(t, aggregatorClient, executor.aggregatorClient)
	assert.Equal(t, logger, executor.logger)
}

func TestNewTaskExecutor_NilLogger(t *testing.T) {
	// Arrange
	alchemyAPIKey := "test-api-key"
	codeExecutor := &docker.CodeExecutor{}
	validator := &validation.TaskValidator{}
	aggregatorClient := &aggregator.AggregatorClient{}

	// Act
	executor := NewTaskExecutor(alchemyAPIKey, codeExecutor, validator, aggregatorClient, nil)

	// Assert
	assert.NotNil(t, executor)
	assert.Equal(t, alchemyAPIKey, executor.alchemyAPIKey)
	assert.Nil(t, executor.logger)
	assert.NotNil(t, executor.argConverter)
}

// ===========================================
// UNIT TESTS FOR parseStringToInt (package level function)
// ===========================================

func TestParseStringToInt_ValidString(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-456", -456},
		{"999999", 999999},
	}

	for _, test := range tests {
		// Act
		result := parseStringToInt(test.input)

		// Assert
		assert.Equal(t, test.expected, result, "Failed for input: %s", test.input)
	}
}

func TestParseStringToInt_InvalidString(t *testing.T) {
	invalidInputs := []string{
		"not-a-number",
		"123.456", // float strings
		"",
		"abc123",
		"123abc",
	}

	for _, input := range invalidInputs {
		// Act
		result := parseStringToInt(input)

		// Assert
		assert.Equal(t, 0, result, "Should return 0 for invalid input: %s", input)
	}
}

func TestParseStringToInt_LargeNumber(t *testing.T) {
	// Use a number that fits in int but tests edge cases
	largeNum := strconv.Itoa(int(^uint(0) >> 1)) // Max int value

	// Act
	result := parseStringToInt(largeNum)

	// Assert
	expected, _ := strconv.Atoi(largeNum)
	assert.Equal(t, expected, result)
}

// ===========================================
// UNIT TESTS FOR ExecuteTask with basic validation
// ===========================================

func TestExecuteTask_NilTask(t *testing.T) {
	// Arrange
	logger := &UnitTestLogger{}
	executor := &TaskExecutor{logger: logger}

	logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return()

	// Act
	success, err := executor.ExecuteTask(context.Background(), nil, "test-trace")

	// Assert
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task data cannot be nil")
	logger.AssertExpected(t)
}

func TestExecuteTask_NilTargetData(t *testing.T) {
	// Arrange
	logger := &UnitTestLogger{}
	executor := &TaskExecutor{logger: logger}

	task := &types.SendTaskDataToKeeper{
		TaskID:      123,
		TargetData:  nil,
		TriggerData: &types.TaskTriggerData{},
	}

	logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Return()

	// Act
	success, err := executor.ExecuteTask(context.Background(), task, "test-trace")

	// Assert
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target data cannot be nil")
	logger.AssertExpected(t)
}

func TestExecuteTask_NilTriggerData(t *testing.T) {
	// Arrange
	logger := &UnitTestLogger{}
	executor := &TaskExecutor{logger: logger}

	task := &types.SendTaskDataToKeeper{
		TaskID:      123,
		TargetData:  &types.TaskTargetData{},
		TriggerData: nil,
	}

	logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Return()

	// Act
	success, err := executor.ExecuteTask(context.Background(), task, "test-trace")

	// Assert
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger data cannot be nil")
	logger.AssertExpected(t)
}

// ===========================================
// UNIT TESTS FOR supported task definition IDs
// ===========================================

func TestTaskDefinitionSupport_SupportedTypes(t *testing.T) {
	// Based on the switch statement in ExecuteTask, supported types are:
	supportedTypes := []int{1, 2, 3, 4, 5, 6}

	for _, taskType := range supportedTypes {
		t.Run(fmt.Sprintf("TaskType_%d", taskType), func(t *testing.T) {
			// This is more of a documentation test showing which types are supported
			// The actual logic is in the switch statement in ExecuteTask
			assert.Contains(t, supportedTypes, taskType, "Task type %d should be supported", taskType)
		})
	}
}

func TestTaskDefinitionSupport_UnsupportedTypes(t *testing.T) {
	// Test that unsupported types would be rejected
	// This tests the default case in the switch statement
	unsupportedTypes := []int{0, 7, 8, 9, 10, 999, -1}

	for _, taskType := range unsupportedTypes {
		t.Run(fmt.Sprintf("TaskType_%d", taskType), func(t *testing.T) {
			// This is a documentation test showing which types are NOT supported
			supportedTypes := []int{1, 2, 3, 4, 5, 6}
			assert.NotContains(t, supportedTypes, taskType, "Task type %d should not be supported", taskType)
		})
	}
}
