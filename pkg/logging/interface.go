package logging

import "context"

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	With(tags ...any) Logger

	DebugWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})
	InfoWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})
	WarnWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})
	ErrorWithTrace(ctx context.Context, msg string, keysAndValues ...interface{})
}
