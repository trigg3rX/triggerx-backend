package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	BaseDataDir = "data"
	LogsDir     = "logs"
)

type LogLevel string

const (
	Development LogLevel = "development" // prints debug and above
	Production  LogLevel = "production"  // prints info and above
)

// ProcessName type to ensure valid process names
type ProcessName string

const (
	ManagerProcess   ProcessName = "manager"
	DatabaseProcess  ProcessName = "database"
	KeeperProcess    ProcessName = "keeper"
)

type ZapLogger struct {
	logger *zap.Logger
}

var _ Logger = (*ZapLogger)(nil)

var (
	loggers map[ProcessName]Logger = make(map[ProcessName]Logger)
)

// NewZapLogger creates a new logger wrapped the zap.Logger
func NewZapLogger(env LogLevel, processName string) (Logger, error) {
	var config zap.Config

	// Create timestamp for log file
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")

	// Create service-specific log directory
	logDir := filepath.Join(BaseDataDir, LogsDir, processName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", timestamp))

	if env == Production {
		config = zap.NewProductionConfig()
		config.OutputPaths = []string{logPath}
	} else if env == Development {
		config = zap.NewDevelopmentConfig()
		// In development, write to both console and file
		config.OutputPaths = []string{"stdout", logPath}
	}

	return NewZapLoggerByConfig(config, zap.AddCallerSkip(1))
}

// NewZapLoggerByConfig creates a logger wrapped the zap.Logger
// Note if the logger need to show the caller, need use `zap.AddCallerSkip(1)` ad options
func NewZapLoggerByConfig(config zap.Config, options ...zap.Option) (Logger, error) {
	logger, err := config.Build(options...)
	if err != nil {
		panic(err)
	}

	return &ZapLogger{
		logger: logger,
	}, nil
}

func (z *ZapLogger) Debug(msg string, tags ...any) {
	z.logger.Sugar().Debugw(msg, tags...)
}

func (z *ZapLogger) Info(msg string, tags ...any) {
	z.logger.Sugar().Infow(msg, tags...)
}

func (z *ZapLogger) Warn(msg string, tags ...any) {
	z.logger.Sugar().Warnw(msg, tags...)
}

func (z *ZapLogger) Error(msg string, tags ...any) {
	z.logger.Sugar().Errorw(msg, tags...)
}

func (z *ZapLogger) Fatal(msg string, tags ...any) {
	z.logger.Sugar().Fatalw(msg, tags...)
}

func (z *ZapLogger) Debugf(template string, args ...interface{}) {
	z.logger.Sugar().Debugf(template, args...)
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	z.logger.Sugar().Infof(template, args...)
}

func (z *ZapLogger) Warnf(template string, args ...interface{}) {
	z.logger.Sugar().Warnf(template, args...)
}

func (z *ZapLogger) Errorf(template string, args ...interface{}) {
	z.logger.Sugar().Errorf(template, args...)
}

func (z *ZapLogger) Fatalf(template string, args ...interface{}) {
	z.logger.Sugar().Fatalf(template, args...)
}

func (z *ZapLogger) With(tags ...any) Logger {
	return &ZapLogger{
		logger: z.logger.Sugar().With(tags...).Desugar(),
	}
}

// InitLogger initializes a logger for a specific process with production environment by default
func InitLogger(env LogLevel, processName ProcessName) error {
	logger, err := NewZapLogger(env, string(processName))
	if err != nil {
		return err
	}
	loggers[processName] = logger
	return nil
}

// GetLogger returns the logger for a specific process
func GetLogger(env LogLevel, processName ProcessName) Logger {
	if logger, exists := loggers[processName]; exists {
		return logger
	}
	// Initialize with production environment if not found
	logger, _ := NewZapLogger(env, string(processName))
	loggers[processName] = logger
	return logger
}
