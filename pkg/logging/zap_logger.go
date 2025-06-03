package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger      *zap.Logger
	useColors   bool
	config      LoggerConfig
	currentDay  string
	currentFile string
	fileSize    int64
	mu          sync.RWMutex
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
	var cores []zapcore.Core
	encoderConfig := z.getEncoderConfig()

	// Get the current log file path and create necessary directories
	logDir := filepath.Join(z.config.LogDir, LogsDir, string(z.config.ProcessName))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Find or create the current log file
	logFile, filePath, err := z.getCurrentLogFile(logDir)
	if err != nil {
		return err
	}
	z.currentFile = filePath

	// Get file info for size tracking
	fileInfo, err := logFile.Stat()
	if err != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		return fmt.Errorf("failed to get file info: %w", err)
	}
	z.fileSize = fileInfo.Size()

	// Create file core with appropriate level
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		z.getMinLevel(z.config.MinFileLogLevel),
	)
	cores = append(cores, fileCore)

	// Add stdout core in development or if explicitly configured
	if z.config.Environment == Development {
		stdoutCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			z.getMinLevel(z.config.MinStdoutLevel),
		)
		cores = append(cores, stdoutCore)
	}

	// Create the logger
	core := zapcore.NewTee(cores...)
	z.logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // Skip our wrapper methods
		zap.Development(),    // Include full caller path
	)
	return nil
}

func (z *ZapLogger) getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.UTC().Format(TimeFormat))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			_, file := filepath.Split(caller.File)
			enc.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
		},
	}
}

func (z *ZapLogger) getCurrentLogFile(logDir string) (*os.File, string, error) {
	// Find the latest file number for today
	pattern := filepath.Join(logDir, fmt.Sprintf("%s.*", z.currentDay))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to glob log files: %w", err)
	}

	var maxNum int
	for _, match := range matches {
		base := filepath.Base(match)
		parts := strings.Split(base, ".")
		if len(parts) >= 3 { // date.number.log
			if num, err := strconv.Atoi(parts[1]); err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	// Create the new file path
	fileName := fmt.Sprintf("%s.%d.log", z.currentDay, maxNum+1)
	filePath := filepath.Join(logDir, fileName)

	// Open the file in append mode
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open log file: %w", err)
	}

	return file, filePath, nil
}

func (z *ZapLogger) getMinLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (z *ZapLogger) compressLogFile(filePath string) error {
	if !z.config.CompressOldFiles {
		return nil
	}

	// Open the log file
	logFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open log file for compression: %w", err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("failed to close log file: %v\n", err)
		}
	}()

	// Create the gzip file
	gzipPath := filePath + ".gz"
	gzipFile, err := os.Create(gzipPath)
	if err != nil {
		return fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer func() {
		if err := gzipFile.Close(); err != nil {
			fmt.Printf("failed to close gzip file: %v\n", err)
		}
	}()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(gzipFile)
	defer func() {
		if err := gzipWriter.Close(); err != nil {
			fmt.Printf("failed to close gzip writer: %v\n", err)
		}
	}()

	// Copy the contents
	if _, err := io.Copy(gzipWriter, logFile); err != nil {
		return fmt.Errorf("failed to compress log file: %w", err)
	}

	// Remove the original file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove original log file: %w", err)
	}

	return nil
}

func (z *ZapLogger) checkAndRotateLog() error {
	z.mu.Lock()
	defer z.mu.Unlock()

	currentDay := time.Now().UTC().Format("2006-01-02")
	needsRotation := currentDay != z.currentDay

	// Check file size
	if !needsRotation {
		if z.fileSize >= MaxFileSize {
			needsRotation = true
		}
	}

	if needsRotation {
		// Close and compress the current log file
		if z.logger != nil {
			if err := z.logger.Sync(); err != nil {
				if !strings.Contains(err.Error(), "sync /dev/stdout") {
					return fmt.Errorf("failed to sync logger during rotation: %w", err)
				}
			}
		}

		// Compress the old file if it's a different day
		if currentDay != z.currentDay {
			if err := z.compressLogFile(z.currentFile); err != nil {
				// Log the error but continue with rotation
				fmt.Printf("failed to compress log file: %v\n", err)
			}
			z.currentDay = currentDay
		}

		// Create new logger
		if err := z.initLogger(); err != nil {
			return fmt.Errorf("failed to rotate logger: %w", err)
		}
	}

	return nil
}

func (z *ZapLogger) updateFileSize(n int64) {
	z.mu.Lock()
	z.fileSize += n
	z.mu.Unlock()
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

// Implement the Logger interface methods
func (z *ZapLogger) Debug(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("debug", msg)
	z.logger.Sugar().Debugw(coloredMsg, tags...)
	z.updateFileSize(int64(len(coloredMsg)))
}

func (z *ZapLogger) Info(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("info", msg)
	z.logger.Sugar().Infow(coloredMsg, tags...)
	z.updateFileSize(int64(len(coloredMsg)))
}

func (z *ZapLogger) Warn(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("warn", msg)
	z.logger.Sugar().Warnw(coloredMsg, tags...)
	z.updateFileSize(int64(len(coloredMsg)))
}

func (z *ZapLogger) Error(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("error", msg)
	z.logger.Sugar().Errorw(coloredMsg, tags...)
	z.updateFileSize(int64(len(coloredMsg)))
}

func (z *ZapLogger) Fatal(msg string, tags ...any) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredMsg := z.colorize("fatal", msg)
	z.logger.Sugar().Fatalw(coloredMsg, tags...)
	// No need to update file size as the program will exit
}

func (z *ZapLogger) Debugf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("debug", template)
	z.logger.Sugar().Debugf(coloredTemplate, args...)
	z.updateFileSize(int64(len(fmt.Sprintf(coloredTemplate, args...))))
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("info", template)
	z.logger.Sugar().Infof(coloredTemplate, args...)
	z.updateFileSize(int64(len(fmt.Sprintf(coloredTemplate, args...))))
}

func (z *ZapLogger) Warnf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("warn", template)
	z.logger.Sugar().Warnf(coloredTemplate, args...)
	z.updateFileSize(int64(len(fmt.Sprintf(coloredTemplate, args...))))
}

func (z *ZapLogger) Errorf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("error", template)
	z.logger.Sugar().Errorf(coloredTemplate, args...)
	z.updateFileSize(int64(len(fmt.Sprintf(coloredTemplate, args...))))
}

func (z *ZapLogger) Fatalf(template string, args ...interface{}) {
	if err := z.checkAndRotateLog(); err != nil {
		z.logger.Error("failed to rotate log", zap.Error(err))
	}
	coloredTemplate := z.colorize("fatal", template)
	z.logger.Sugar().Fatalf(coloredTemplate, args...)
	// No need to update file size as the program will exit
}

func (z *ZapLogger) With(tags ...any) Logger {
	return &ZapLogger{
		logger:      z.logger.Sugar().With(tags...).Desugar(),
		useColors:   z.useColors,
		config:      z.config,
		currentDay:  z.currentDay,
		currentFile: z.currentFile,
		fileSize:    z.fileSize,
	}
}
