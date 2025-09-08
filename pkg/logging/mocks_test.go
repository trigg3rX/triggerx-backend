package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test MockLogger
func TestMockLogger_Debug_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", "test message", []interface{}{"key", "value"}).Return(nil)

	mockLogger.Debug("test message", "key", "value")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Info_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "test message", []interface{}{"key", "value"}).Return(nil)

	mockLogger.Info("test message", "key", "value")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Warn_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Warn", "test message", []interface{}{"key", "value"}).Return(nil)

	mockLogger.Warn("test message", "key", "value")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Error_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Error", "test message", []interface{}{"key", "value"}).Return(nil)

	mockLogger.Error("test message", "key", "value")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Fatal_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Fatal", "test message", []interface{}{"key", "value"}).Return(nil)

	mockLogger.Fatal("test message", "key", "value")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Debugf_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Debugf", "test %s", []interface{}{"message"}).Return(nil)

	mockLogger.Debugf("test %s", "message")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Infof_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Infof", "test %d", []interface{}{42}).Return(nil)

	mockLogger.Infof("test %d", 42)

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Warnf_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Warnf", "test %v", []interface{}{[]string{"a", "b"}}).Return(nil)

	mockLogger.Warnf("test %v", []string{"a", "b"})

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Errorf_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Errorf", "test %s", []interface{}{"error"}).Return(nil)

	mockLogger.Errorf("test %s", "error")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_Fatalf_CallsMockMethod(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Fatalf", "test %s", []interface{}{"fatal"}).Return(nil)

	mockLogger.Fatalf("test %s", "fatal")

	mockLogger.AssertExpectations(t)
}

func TestMockLogger_With_ReturnsMockLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("With", []interface{}{"tag", "value"}).Return(mockLogger)

	result := mockLogger.With("tag", "value")

	assert.Equal(t, mockLogger, result)
	mockLogger.AssertExpectations(t)
}

// Test MockLoggerFactory
func TestMockLoggerFactory_NewMockLogger_ReturnsNewInstance(t *testing.T) {
	factory := &MockLoggerFactory{}

	logger := factory.NewMockLogger()

	assert.NotNil(t, logger)
	assert.IsType(t, &MockLogger{}, logger)
}

func TestMockLoggerFactory_CreateLogger_CallsMockMethod(t *testing.T) {
	factory := &MockLoggerFactory{}
	mockLogger := &MockLogger{}
	config := LoggerConfig{ProcessName: TestProcess, IsDevelopment: true}

	factory.On("CreateLogger", config).Return(mockLogger, nil)

	result, err := factory.CreateLogger(config)

	assert.NoError(t, err)
	assert.Equal(t, mockLogger, result)
	factory.AssertExpectations(t)
}

// Test MockSequentialRotator
func TestMockSequentialRotator_Write_CallsMockMethod(t *testing.T) {
	mockRotator := &MockSequentialRotator{}
	data := []byte("test data")

	mockRotator.On("Write", data).Return(len(data), nil)

	n, err := mockRotator.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	mockRotator.AssertExpectations(t)
}

func TestMockSequentialRotator_Close_CallsMockMethod(t *testing.T) {
	mockRotator := &MockSequentialRotator{}

	mockRotator.On("Close").Return(nil)

	err := mockRotator.Close()

	assert.NoError(t, err)
	mockRotator.AssertExpectations(t)
}

// Test NoOpLogger
func TestNoOpLogger_AllMethods_DoNothing(t *testing.T) {
	logger := NewNoOpLogger()

	// These should not panic
	assert.NotPanics(t, func() {
		logger.Debug("test")
		logger.Info("test")
		logger.Warn("test")
		logger.Error("test")
		logger.Fatal("test")
		logger.Debugf("test %s", "message")
		logger.Infof("test %d", 42)
		logger.Warnf("test %v", []string{"a"})
		logger.Errorf("test %s", "error")
		logger.Fatalf("test %s", "fatal")
	})
}

func TestNoOpLogger_With_ReturnsSameLogger(t *testing.T) {
	logger := NewNoOpLogger()

	result := logger.With("tag", "value")

	assert.Equal(t, logger, result)
}

// Test MockLoggerBuilder
func TestMockLoggerBuilder_ExpectDebug_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectDebug("test message", "key", "value").Build()

	mockLogger.Debug("test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectInfo_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectInfo("test message", "key", "value").Build()

	mockLogger.Info("test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectWarn_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectWarn("test message", "key", "value").Build()

	mockLogger.Warn("test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectError_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectError("test message", "key", "value").Build()

	mockLogger.Error("test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectFatal_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectFatal("test message", "key", "value").Build()

	mockLogger.Fatal("test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectDebugf_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectDebugf("test %s", "message").Build()

	mockLogger.Debugf("test %s", "message")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectInfof_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectInfof("test %d", 42).Build()

	mockLogger.Infof("test %d", 42)
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectWarnf_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectWarnf("test %v", []string{"a", "b"}).Build()

	mockLogger.Warnf("test %v", []string{"a", "b"})
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectErrorf_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectErrorf("test %s", "error").Build()

	mockLogger.Errorf("test %s", "error")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectFatalf_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectFatalf("test %s", "fatal").Build()

	mockLogger.Fatalf("test %s", "fatal")
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_ExpectWith_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectWith("tag", "value").Build()

	result := mockLogger.With("tag", "value")
	assert.Equal(t, mockLogger, result)
	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_MultipleExpectations_BuildsMockLogger(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.
		ExpectInfo("start", "phase", "init").
		ExpectDebug("processing", "step", 1).
		ExpectWarn("warning", "issue", "minor").
		ExpectError("error", "code", 500).
		Build()

	mockLogger.Info("start", "phase", "init")
	mockLogger.Debug("processing", "step", 1)
	mockLogger.Warn("warning", "issue", "minor")
	mockLogger.Error("error", "code", 500)

	mockLogger.AssertExpectations(t)
}

func TestMockLoggerBuilder_AssertNumberOfCalls_WorksCorrectly(t *testing.T) {
	builder := NewMockLoggerBuilder()

	mockLogger := builder.ExpectInfo("test", "key", "value").Build()

	mockLogger.Info("test", "key", "value")
	mockLogger.Info("test", "key", "value")

	assert.True(t, builder.AssertNumberOfCalls(t, "Info", 2))
}

// Test MockLoggerConfig
func TestNewMockLoggerConfig_ReturnsValidConfig(t *testing.T) {
	config := NewMockLoggerConfig()

	assert.Equal(t, TestProcess, config.ProcessName)
	assert.True(t, config.IsDevelopment)
}

// Test NewTestLogger functions
func TestNewTestLogger_CreatesLogger(t *testing.T) {
	logger, err := NewTestLogger()

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewTestLoggerWithConfig_CreatesLogger(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   AggregatorProcess,
		IsDevelopment: false,
	}

	logger, err := NewTestLoggerWithConfig(config)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

// Integration test showing how other packages would use the mocks
func TestMockLogger_IntegrationExample_ShowsUsage(t *testing.T) {
	// This demonstrates how other packages would use the mock logger

	// Create a mock logger with expectations
	mockLogger := NewMockLoggerBuilder().
		ExpectInfo("User logged in", "user_id", "123", "ip", "192.168.1.1").
		ExpectDebug("Processing request", "method", "GET", "path", "/api/users").
		ExpectWarn("Rate limit approaching", "current", 95, "limit", 100).
		ExpectError("Database connection failed", "error", "connection timeout").
		Build()

	// Simulate application behavior
	mockLogger.Info("User logged in", "user_id", "123", "ip", "192.168.1.1")
	mockLogger.Debug("Processing request", "method", "GET", "path", "/api/users")
	mockLogger.Warn("Rate limit approaching", "current", 95, "limit", 100)
	mockLogger.Error("Database connection failed", "error", "connection timeout")

	// Verify all expectations were met
	mockLogger.AssertExpectations(t)
}

// Test showing how to use NoOpLogger for tests that don't care about logging
func TestNoOpLogger_IntegrationExample_ShowsUsage(t *testing.T) {
	// This shows how to use NoOpLogger when logging is not important for the test

	logger := NewNoOpLogger()

	// These calls should not affect the test
	logger.Info("This won't be logged")
	logger.Error("This won't be logged either")
	logger.With("tag", "value").Info("Still won't be logged")

	// Test passes if no panic occurs
	assert.True(t, true)
}

// Test showing how to use MockSequentialRotator
func TestMockSequentialRotator_IntegrationExample_ShowsUsage(t *testing.T) {
	// This demonstrates how to use the mock rotator

	mockRotator := &MockSequentialRotator{}

	// Set up expectations
	mockRotator.On("Write", []byte("log entry 1\n")).Return(12, nil)
	mockRotator.On("Write", []byte("log entry 2\n")).Return(12, nil)
	mockRotator.On("Close").Return(nil)

	// Simulate logging behavior
	data1 := []byte("log entry 1\n")
	data2 := []byte("log entry 2\n")

	n1, err1 := mockRotator.Write(data1)
	assert.NoError(t, err1)
	assert.Equal(t, 12, n1)

	n2, err2 := mockRotator.Write(data2)
	assert.NoError(t, err2)
	assert.Equal(t, 12, n2)

	err3 := mockRotator.Close()
	assert.NoError(t, err3)

	// Verify all expectations were met
	mockRotator.AssertExpectations(t)
}
