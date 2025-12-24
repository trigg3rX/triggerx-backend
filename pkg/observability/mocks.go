package observability

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new instance of MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Debug mocks the Debug method
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

// Info mocks the Info method
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

// Warn mocks the Warn method
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

// Error mocks the Error method
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

// Fatal mocks the Fatal method
func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

// With mocks the With method
func (m *MockLogger) With(fields ...Field) Logger {
	args := m.Called(fields)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(Logger)
}

// noopLogger is a no-op implementation of Logger that discards all log calls
type noopLogger struct{}

// NewNoOpLogger creates a new no-op logger that discards all log calls
// This is useful for testing when you don't want actual logging to occur
func NewNoOpLogger() Logger {
	return &noopLogger{}
}

// Debug is a no-op
func (l *noopLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	// No-op
}

// Info is a no-op
func (l *noopLogger) Info(ctx context.Context, msg string, fields ...Field) {
	// No-op
}

// Warn is a no-op
func (l *noopLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	// No-op
}

// Error is a no-op
func (l *noopLogger) Error(ctx context.Context, msg string, fields ...Field) {
	// No-op
}

// Fatal is a no-op (does not exit in test context)
func (l *noopLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	// No-op - in tests, we don't want to exit
}

// With returns itself as no-op logger ignores fields
func (l *noopLogger) With(fields ...Field) Logger {
	return l
}
