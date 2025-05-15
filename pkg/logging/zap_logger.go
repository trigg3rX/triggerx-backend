package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger     *zap.Logger
	useColors  bool
	config     LoggerConfig
	currentDay string
	mu         sync.RWMutex
}

var _ Logger = (*ZapLogger)(nil)

func NewZapLogger(config LoggerConfig) (Logger, error) {
	logger := &ZapLogger{
		config:     config,
		useColors:  config.UseColors,
		currentDay: time.Now().UTC().Format("2006-01-02"),
	}

	if err := logger.initLogger(); err != nil {
		return nil, err
	}

	return logger, nil
}

func (z *ZapLogger) initLogger() error {
	var zapConfig zap.Config

	logDir := filepath.Join(z.config.LogDir, LogsDir, string(z.config.ProcessName))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", z.currentDay))

	if z.config.Environment == Production {
		zapConfig = zap.NewProductionConfig()
		zapConfig.OutputPaths = []string{logPath}
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.OutputPaths = []string{"stdout", logPath}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.UTC().Format(TimeFormat))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
			_, file := filepath.Split(caller.File)
			encoder.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
		},
	}

	zapConfig.EncoderConfig = encoderConfig

	// Create a new logger
	newLogger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	// If we have an existing logger, sync and close it first
	if z.logger != nil {
		// Sync but ignore stdout sync errors
		if err := z.logger.Sync(); err != nil {
			// Check if error is related to stdout
			if !strings.Contains(err.Error(), "sync /dev/stdout") {
				return fmt.Errorf("failed to sync logger: %w", err)
			}
		}
	}

	// Set the new logger
	z.logger = newLogger
	return nil
}

func (z *ZapLogger) checkAndRotateLog() error {
	z.mu.Lock()
	defer z.mu.Unlock()

	currentDay := z.currentDay
	if z.config.Environment == Production {
		currentDay = time.Now().UTC().Format("2006-01-02")
	}

	if currentDay != z.currentDay {
		// Close the current log file
		if z.logger != nil {
			// Sync but ignore stdout sync errors
			if err := z.logger.Sync(); err != nil {
				// Check if error is related to stdout
				if !strings.Contains(err.Error(), "sync /dev/stdout") {
					return fmt.Errorf("failed to sync logger during rotation: %w", err)
				}
			}
		}

		// Update current day and create new log file
		z.currentDay = currentDay
		if err := z.initLogger(); err != nil {
			return fmt.Errorf("failed to rotate logger: %w", err)
		}
	}
	return nil
}

func (z *ZapLogger) colorize(level, msg string) string {
	if !z.useColors {
		return msg
	}

	switch level {
	case "debug":
		return fmt.Sprintf("[%sdebug%s] %s", colorBlue, colorReset, msg)
	case "info":
		return fmt.Sprintf("[%sinfo%s] %s", colorGreen, colorReset, msg)
	case "warn":
		return fmt.Sprintf("[%swarn%s] %s", colorYellow, colorReset, msg)
	case "error":
		return fmt.Sprintf("[%serror%s] %s", colorRed, colorReset, msg)
	case "fatal":
		return fmt.Sprintf("[%sfatal%s] %s", colorPurple, colorReset, msg)
	default:
		return msg
	}
}

func (z *ZapLogger) Debug(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("debug", msg)
	z.logger.Sugar().Debugw(coloredMsg, tags...)
}

func (z *ZapLogger) Info(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("info", msg)
	z.logger.Sugar().Infow(coloredMsg, tags...)
}

func (z *ZapLogger) Warn(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("warn", msg)
	z.logger.Sugar().Warnw(coloredMsg, tags...)
}

func (z *ZapLogger) Error(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("error", msg)
	z.logger.Sugar().Errorw(coloredMsg, tags...)
}

func (z *ZapLogger) Fatal(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("fatal", msg)
	z.logger.Sugar().Fatalw(coloredMsg, tags...)
}

func (z *ZapLogger) Debugf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("debug", template)
	z.logger.Sugar().Debugf(coloredTemplate, args...)
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("info", template)
	z.logger.Sugar().Infof(coloredTemplate, args...)
}

func (z *ZapLogger) Warnf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("warn", template)
	z.logger.Sugar().Warnf(coloredTemplate, args...)
}

func (z *ZapLogger) Errorf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("error", template)
	z.logger.Sugar().Errorf(coloredTemplate, args...)
}

func (z *ZapLogger) Fatalf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("fatal", template)
	z.logger.Sugar().Fatalf(coloredTemplate, args...)
}

func (z *ZapLogger) With(tags ...any) Logger {
	return &ZapLogger{
		logger:     z.logger.Sugar().With(tags...).Desugar(),
		useColors:  z.useColors,
		config:     z.config,
		currentDay: z.currentDay,
	}
}
