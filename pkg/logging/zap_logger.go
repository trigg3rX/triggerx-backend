package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	BaseDataDir = "data"
	LogsDir     = "logs"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

type LogLevel string

const (
	Development LogLevel = "development"
	Production  LogLevel = "production"
)

type ProcessName string

const (
	ManagerProcess   ProcessName = "manager"
	DatabaseProcess  ProcessName = "database"
	KeeperProcess    ProcessName = "keeper"
	RegistrarProcess ProcessName = "registrar"
	HealthProcess    ProcessName = "health"
)

type ZapLogger struct {
	logger    *zap.Logger
	useColors bool
}

var _ Logger = (*ZapLogger)(nil)

var (
	loggers map[ProcessName]Logger = make(map[ProcessName]Logger)
)

func NewZapLogger(env LogLevel, processName string) (Logger, error) {
	var config zap.Config

	timestamp := time.Now().UTC().Format(time.RFC3339)

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
		config.OutputPaths = []string{"stdout", logPath}
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
			encoder.AppendString(t.UTC().Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
			_, file := filepath.Split(caller.File)
			encoder.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
		},
	}

	config.EncoderConfig = encoderConfig

	return NewZapLoggerByConfig(config, zap.AddCallerSkip(1))
}

func NewZapLoggerByConfig(config zap.Config, options ...zap.Option) (Logger, error) {
	logger, err := config.Build(options...)
	if err != nil {
		panic(err)
	}

	return &ZapLogger{
		logger:    logger,
		useColors: true,
	}, nil
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
	coloredMsg := z.colorize("debug", msg)
	z.logger.Sugar().Debugw(coloredMsg, tags...)
}

func (z *ZapLogger) Info(msg string, tags ...any) {
	coloredMsg := z.colorize("info", msg)
	z.logger.Sugar().Infow(coloredMsg, tags...)
}

func (z *ZapLogger) Warn(msg string, tags ...any) {
	coloredMsg := z.colorize("warn", msg)
	z.logger.Sugar().Warnw(coloredMsg, tags...)
}

func (z *ZapLogger) Error(msg string, tags ...any) {
	coloredMsg := z.colorize("error", msg)
	z.logger.Sugar().Errorw(coloredMsg, tags...)
}

func (z *ZapLogger) Fatal(msg string, tags ...any) {
	coloredMsg := z.colorize("fatal", msg)
	z.logger.Sugar().Fatalw(coloredMsg, tags...)
}

func (z *ZapLogger) Debugf(template string, args ...interface{}) {
	coloredTemplate := z.colorize("debug", template)
	z.logger.Sugar().Debugf(coloredTemplate, args...)
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	coloredTemplate := z.colorize("info", template)
	z.logger.Sugar().Infof(coloredTemplate, args...)
}

func (z *ZapLogger) Warnf(template string, args ...interface{}) {
	coloredTemplate := z.colorize("warn", template)
	z.logger.Sugar().Warnf(coloredTemplate, args...)
}

func (z *ZapLogger) Errorf(template string, args ...interface{}) {
	coloredTemplate := z.colorize("error", template)
	z.logger.Sugar().Errorf(coloredTemplate, args...)
}

func (z *ZapLogger) Fatalf(template string, args ...interface{}) {
	coloredTemplate := z.colorize("fatal", template)
	z.logger.Sugar().Fatalf(coloredTemplate, args...)
}

func (z *ZapLogger) With(tags ...any) Logger {
	return &ZapLogger{
		logger:    z.logger.Sugar().With(tags...).Desugar(),
		useColors: z.useColors,
	}
}

func InitLogger(env LogLevel, processName ProcessName) error {
	logger, err := NewZapLogger(env, string(processName))
	if err != nil {
		return err
	}
	loggers[processName] = logger
	return nil
}

func GetLogger(env LogLevel, processName ProcessName) Logger {
	if logger, exists := loggers[processName]; exists {
		return logger
	}
	logger, _ := NewZapLogger(env, string(processName))
	loggers[processName] = logger
	return logger
}

func SetUseColors(processName ProcessName, useColors bool) {
	if logger, exists := loggers[processName]; exists {
		if zapLogger, ok := logger.(*ZapLogger); ok {
			zapLogger.useColors = useColors
		}
	}
}

func init() {
	for _, process := range []ProcessName{ManagerProcess, DatabaseProcess, KeeperProcess, RegistrarProcess} {
		logger, _ := NewZapLogger(Development, string(process))
		if zapLogger, ok := logger.(*ZapLogger); ok {
			zapLogger.useColors = true
		}
		loggers[process] = logger
	}
}
