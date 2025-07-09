package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	sugarLogger *zap.SugaredLogger
}

func NewZapLogger(config LoggerConfig) (*zapLogger, error) {
	fileName := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02"))

	// File writer with rotation, compression and no colors
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename: filepath.Join(BaseDataDir, LogsDir, string(config.ProcessName), fileName),
		MaxSize:  30,   // MB
		MaxAge:   30,   // Days
		Compress: false, // Compress old logs (disabled for now for Loki compatibility)
	})

	// Console output (with colors)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Create cores
	fileCore := zapcore.NewCore(
		plainFileEncoder(),
		fileWriter,
		zapcore.DebugLevel,
	)
	consoleCore := zapcore.NewCore(
		coloredConsoleEncoder(),
		consoleWriter,
		getLogLevel(config.IsDevelopment),
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Build the logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	return &zapLogger{
		sugarLogger: logger.Sugar(),
	}, nil
}

func getLogLevel(isDevelopment bool) zapcore.Level {
	if isDevelopment {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

// Custom console encoder with colors
func coloredConsoleEncoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    customColorLevelEncoder,
		// TODO: Add short file name, lowest priority, for aesthetics
	}
	return zapcore.NewConsoleEncoder(config)
}

// Plain JSON encoder for files (no colors)
func plainFileEncoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
	}
	return zapcore.NewJSONEncoder(config)
}

// Custom color encoder (stdout only)
func customColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		color = colorBlue
		levelStr = "DBG"
	case zapcore.InfoLevel:
		color = colorGreen
		levelStr = "INF"
	case zapcore.WarnLevel:
		color = colorYellow
		levelStr = "WRN"
	case zapcore.ErrorLevel:
		color = colorRed
		levelStr = "ERR"
	case zapcore.FatalLevel:
		color = colorMagenta
		levelStr = "FTL"
	default:
		color = colorWhite
		levelStr = "???"
	}

	// Format: [COLOR][LEVEL][RESET] (e.g., "\x1b[32mINF\x1b[0m")
	enc.AppendString(fmt.Sprintf("%s%-3s%s", color, levelStr, colorReset))
}

// Implement the Logger interface methods
func (z *zapLogger) Debug(msg string, keysAndValues ...any) {
	z.sugarLogger.Debugw(msg, keysAndValues...)
}

func (z *zapLogger) Info(msg string, keysAndValues ...any) {
	z.sugarLogger.Infow(msg, keysAndValues...)
}

func (z *zapLogger) Warn(msg string, keysAndValues ...any) {
	z.sugarLogger.Warnw(msg, keysAndValues...)
}

func (z *zapLogger) Error(msg string, keysAndValues ...any) {
	z.sugarLogger.Errorw(msg, keysAndValues...)
}

func (z *zapLogger) Fatal(msg string, keysAndValues ...any) {
	z.sugarLogger.Fatalw(msg, keysAndValues...)
}

func (z *zapLogger) Debugf(template string, args ...interface{}) {
	z.sugarLogger.Debugf(template, args...)
}

func (z *zapLogger) Infof(template string, args ...interface{}) {
	z.sugarLogger.Infof(template, args...)
}

func (z *zapLogger) Warnf(template string, args ...interface{}) {
	z.sugarLogger.Warnf(template, args...)
}

func (z *zapLogger) Errorf(template string, args ...interface{}) {
	z.sugarLogger.Errorf(template, args...)
}

func (z *zapLogger) Fatalf(template string, args ...interface{}) {
	z.sugarLogger.Fatalf(template, args...)
}

func (z *zapLogger) With(tags ...any) Logger {
	return &zapLogger{
		sugarLogger: z.sugarLogger.With(tags...),
	}
}
