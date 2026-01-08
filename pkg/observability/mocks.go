package observability

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

// noopTracer is a no-op implementation of Tracer that discards all trace calls
type noopTracer struct{}

// NewNoOpTracer creates a new no-op tracer that discards all trace calls
// This is useful for testing when you don't want actual tracing to occur
func NewNoOpTracer() Tracer {
	return &noopTracer{}
}

// Start returns the context unchanged and a no-op span
func (t *noopTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

// StartSpan returns a no-op span
func (t *noopTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) Span {
	return &noopSpan{}
}

// noopSpan is a no-op implementation of Span that discards all span operations
type noopSpan struct{}

// End is a no-op
func (s *noopSpan) End() {
	// No-op
}

// SetAttributes is a no-op
func (s *noopSpan) SetAttributes(attrs ...attribute.KeyValue) {
	// No-op
}

// AddEvent is a no-op
func (s *noopSpan) AddEvent(name string, opts ...EventOption) {
	// No-op
}

// RecordError is a no-op
func (s *noopSpan) RecordError(err error, opts ...RecordErrorOption) {
	// No-op
}

// SetStatus is a no-op
func (s *noopSpan) SetStatus(code codes.Code, description string) {
	// No-op
}

// SpanContext returns an empty span context
func (s *noopSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}
