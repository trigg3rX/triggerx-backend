package logging

import (
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

// SetupDefaultExpectations sets up common logger mock expectations that accept any arguments.
// This is useful for tests where you don't care about specific logging calls.
// It allows all logger methods to be called with any arguments without causing test failures.
func (m *MockLogger) SetupDefaultExpectations() {
	// Debug methods with various argument counts
	m.On("Debug", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Debugf", mock.Anything).Maybe().Return()
	m.On("Debugf", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	m.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	m.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Info methods with various argument counts
	m.On("Info", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Infof", mock.Anything).Maybe().Return()
	m.On("Infof", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Infof", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	m.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	m.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Warn methods with various argument counts
	m.On("Warn", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Warnf", mock.Anything).Maybe().Return()
	m.On("Warnf", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Error methods with various argument counts
	m.On("Error", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Errorf", mock.Anything).Maybe().Return()
	m.On("Errorf", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	m.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Fatal methods with various argument counts
	m.On("Fatal", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Fatalf", mock.Anything).Maybe().Return()
	m.On("Fatalf", mock.Anything, mock.Anything).Maybe().Return()
	m.On("Fatalf", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
}

// Debug mocks the Debug method
func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Info mocks the Info method
func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Warn mocks the Warn method
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Error mocks the Error method
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Fatal mocks the Fatal method
func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Debugf mocks the Debugf method
func (m *MockLogger) Debugf(template string, args ...interface{}) {
	m.Called(template, args)
}

// Infof mocks the Infof method
func (m *MockLogger) Infof(template string, args ...interface{}) {
	m.Called(template, args)
}

// Warnf mocks the Warnf method
func (m *MockLogger) Warnf(template string, args ...interface{}) {
	m.Called(template, args)
}

// Errorf mocks the Errorf method
func (m *MockLogger) Errorf(template string, args ...interface{}) {
	m.Called(template, args)
}

// Fatalf mocks the Fatalf method
func (m *MockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

// With mocks the With method
func (m *MockLogger) With(tags ...any) Logger {
	args := m.Called(tags)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(Logger)
}

// WithTraceID mocks the WithTraceID method
func (m *MockLogger) WithTraceID(traceID string) Logger {
	args := m.Called(traceID)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(Logger)
}

// MockLoggerFactory is a factory for creating mock loggers
// This can be used to inject mock loggers into components under test
type MockLoggerFactory struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger instance
func (f *MockLoggerFactory) NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// CreateLogger mocks the logger creation process
func (f *MockLoggerFactory) CreateLogger(config LoggerConfig) (Logger, error) {
	args := f.Called(config)
	return args.Get(0).(Logger), args.Error(1)
}

// MockSequentialRotator is a mock implementation of the SequentialRotator
// This can be used to test components that depend on log rotation
type MockSequentialRotator struct {
	mock.Mock
}

// Write mocks the Write method
func (m *MockSequentialRotator) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

// Close mocks the Close method
func (m *MockSequentialRotator) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockLoggerConfig is a mock configuration for testing
type MockLoggerConfig struct {
	ProcessName   ProcessName
	IsDevelopment bool
}

// NewMockLoggerConfig creates a new mock logger config
func NewMockLoggerConfig() MockLoggerConfig {
	return MockLoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}
}

// Test utilities for common mock setups

// NewNoOpLogger creates a logger that does nothing (useful for tests that don't care about logging)
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// NoOpLogger is a logger implementation that does nothing
// Useful for tests where logging is not important
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (n *NoOpLogger) Info(msg string, keysAndValues ...interface{})  {}
func (n *NoOpLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (n *NoOpLogger) Error(msg string, keysAndValues ...interface{}) {}
func (n *NoOpLogger) Fatal(msg string, keysAndValues ...interface{}) {}
func (n *NoOpLogger) Debugf(template string, args ...interface{})    {}
func (n *NoOpLogger) Infof(template string, args ...interface{})     {}
func (n *NoOpLogger) Warnf(template string, args ...interface{})     {}
func (n *NoOpLogger) Errorf(template string, args ...interface{})    {}
func (n *NoOpLogger) Fatalf(template string, args ...interface{})    {}
func (n *NoOpLogger) With(tags ...interface{}) Logger                { return n }
func (n *NoOpLogger) WithTraceID(traceID string) Logger              { return n }

// NewTestLogger creates a logger suitable for testing
// It uses a mock rotator to avoid file system dependencies
func NewTestLogger() (Logger, error) {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}
	return NewZapLogger(config)
}

// NewTestLoggerWithConfig creates a test logger with custom configuration
func NewTestLoggerWithConfig(config LoggerConfig) (Logger, error) {
	return NewZapLogger(config)
}

// MockLoggerBuilder provides a fluent interface for building mock loggers
type MockLoggerBuilder struct {
	logger *MockLogger
}

// NewMockLoggerBuilder creates a new mock logger builder
func NewMockLoggerBuilder() *MockLoggerBuilder {
	return &MockLoggerBuilder{
		logger: &MockLogger{},
	}
}

// ExpectDebug sets up an expectation for a Debug call
func (b *MockLoggerBuilder) ExpectDebug(msg string, keysAndValues ...interface{}) *MockLoggerBuilder {
	b.logger.On("Debug", msg, keysAndValues).Return(nil)
	return b
}

// ExpectInfo sets up an expectation for an Info call
func (b *MockLoggerBuilder) ExpectInfo(msg string, keysAndValues ...interface{}) *MockLoggerBuilder {
	b.logger.On("Info", msg, keysAndValues).Return(nil)
	return b
}

// ExpectWarn sets up an expectation for a Warn call
func (b *MockLoggerBuilder) ExpectWarn(msg string, keysAndValues ...interface{}) *MockLoggerBuilder {
	b.logger.On("Warn", msg, keysAndValues).Return(nil)
	return b
}

// ExpectError sets up an expectation for an Error call
func (b *MockLoggerBuilder) ExpectError(msg string, keysAndValues ...interface{}) *MockLoggerBuilder {
	b.logger.On("Error", msg, keysAndValues).Return(nil)
	return b
}

// ExpectFatal sets up an expectation for a Fatal call
func (b *MockLoggerBuilder) ExpectFatal(msg string, keysAndValues ...interface{}) *MockLoggerBuilder {
	b.logger.On("Fatal", msg, keysAndValues).Return(nil)
	return b
}

// ExpectDebugf sets up an expectation for a Debugf call
func (b *MockLoggerBuilder) ExpectDebugf(template string, args ...interface{}) *MockLoggerBuilder {
	b.logger.On("Debugf", template, args).Return(nil)
	return b
}

// ExpectInfof sets up an expectation for an Infof call
func (b *MockLoggerBuilder) ExpectInfof(template string, args ...interface{}) *MockLoggerBuilder {
	b.logger.On("Infof", template, args).Return(nil)
	return b
}

// ExpectWarnf sets up an expectation for a Warnf call
func (b *MockLoggerBuilder) ExpectWarnf(template string, args ...interface{}) *MockLoggerBuilder {
	b.logger.On("Warnf", template, args).Return(nil)
	return b
}

// ExpectErrorf sets up an expectation for an Errorf call
func (b *MockLoggerBuilder) ExpectErrorf(template string, args ...interface{}) *MockLoggerBuilder {
	b.logger.On("Errorf", template, args).Return(nil)
	return b
}

// ExpectFatalf sets up an expectation for a Fatalf call
func (b *MockLoggerBuilder) ExpectFatalf(template string, args ...interface{}) *MockLoggerBuilder {
	b.logger.On("Fatalf", template, args).Return(nil)
	return b
}

// ExpectWith sets up an expectation for a With call
func (b *MockLoggerBuilder) ExpectWith(tags ...interface{}) *MockLoggerBuilder {
	b.logger.On("With", tags).Return(b.logger)
	return b
}

// ExpectWithTraceID sets up an expectation for a WithTraceID call
func (b *MockLoggerBuilder) ExpectWithTraceID(traceID string) *MockLoggerBuilder {
	b.logger.On("WithTraceID", traceID).Return(b.logger)
	return b
}

// Build returns the configured mock logger
func (b *MockLoggerBuilder) Build() *MockLogger {
	return b.logger
}

// AssertExpectations asserts that all expected calls were made
func (b *MockLoggerBuilder) AssertExpectations(t mock.TestingT) bool {
	return b.logger.AssertExpectations(t)
}

// AssertNumberOfCalls asserts the number of calls to a specific method
func (b *MockLoggerBuilder) AssertNumberOfCalls(t mock.TestingT, methodName string, expectedCalls int) bool {
	return b.logger.AssertNumberOfCalls(t, methodName, expectedCalls)
}
