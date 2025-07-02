package handlers

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type MockLogger struct{}

func (l *MockLogger) Debug(msg string, tags ...any)               {}
func (l *MockLogger) Info(msg string, tags ...any)                {}
func (l *MockLogger) Warn(msg string, tags ...any)                {}
func (l *MockLogger) Error(msg string, tags ...any)               {}
func (l *MockLogger) Fatal(msg string, tags ...any)               {}
func (l *MockLogger) Debugf(template string, args ...interface{}) {}
func (l *MockLogger) Infof(template string, args ...interface{})  {}
func (l *MockLogger) Warnf(template string, args ...interface{})  {}
func (l *MockLogger) Errorf(template string, args ...interface{}) {}
func (l *MockLogger) Fatalf(template string, args ...interface{}) {}
func (l *MockLogger) With(tags ...any) logging.Logger             { return l }

// Trace correlation methods (no-op implementations for testing)
func (l *MockLogger) DebugWithTrace(ctx context.Context, msg string, keysAndValues ...interface{}) {}
func (l *MockLogger) InfoWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})  {}
func (l *MockLogger) WarnWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})  {}
func (l *MockLogger) ErrorWithTrace(ctx context.Context, msg string, keysAndValues ...interface{}) {}
