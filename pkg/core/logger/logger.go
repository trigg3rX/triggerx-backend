package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

func Info(format string, v ...interface{}) {
	log.Sugar().Infof(format, v...)
}

func Event(format string, v ...interface{}) {
	log.Sugar().Infof(format, v...)
}

func Error(format string, v ...interface{}) {
	log.Sugar().Errorf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	log.Sugar().Fatalf(format, v...)
}

func InfoWithFields(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func ErrorWithFields(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func EventWithFields(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func FatalWithFields(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}