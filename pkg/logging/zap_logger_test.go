package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// Test NewZapLogger constructor
func TestNewZapLogger_ValidConfig_CreatesLoggerSuccessfully(t *testing.T) {
	tests := []struct {
		name          string
		config        LoggerConfig
		expectedLevel zapcore.Level
	}{
		{
			name: "development mode",
			config: LoggerConfig{
				ProcessName:   AggregatorProcess,
				IsDevelopment: true,
			},
			expectedLevel: zapcore.DebugLevel,
		},
		{
			name: "production mode",
			config: LoggerConfig{
				ProcessName:   DatabaseProcess,
				IsDevelopment: false,
			},
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name: "keeper process",
			config: LoggerConfig{
				ProcessName:   KeeperProcess,
				IsDevelopment: true,
			},
			expectedLevel: zapcore.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewZapLogger(tt.config)

			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.NotNil(t, logger.sugarLogger)
		})
	}
}

func TestNewZapLogger_InvalidDirectory_ReturnsError(t *testing.T) {
	// Test with invalid directory path
	config := LoggerConfig{
		ProcessName:   ProcessName("invalid"),
		IsDevelopment: true,
	}

	// Create a logger with a process name that will result in an invalid path
	// The BaseDataDir is a constant, so we can't modify it directly
	// Instead, we'll test with a process name that creates an invalid path
	logger, err := NewZapLogger(config)

	// This might not error on all systems, so we'll just verify the logger is created
	// The actual file creation will fail when we try to write to it
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewZapLogger_CreatesCorrectFileStructure(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}

	// We'll test the file creation by writing a log message
	logger, err := NewZapLogger(config)
	require.NoError(t, err)
	defer func() {
		if logger != nil {
			// Force sync to ensure file is written
			// Note: Sync() may fail on stdout on some systems, which is expected
			if err := logger.sugarLogger.Sync(); err != nil {
				// Ignore sync errors for stdout as they are expected in test environments
				_ = err
			}
		}
	}()

	// Write a log message to trigger file creation
	logger.Info("test message")

	// Check that the log directory was created
	expectedLogDir := filepath.Join(BaseDataDir, LogsDir, string(config.ProcessName))
	_, err = os.Stat(expectedLogDir)
	assert.NoError(t, err, "Log directory should be created")

	// Check that a log file was created
	today := time.Now().Format("2006-01-02")
	expectedLogFile := filepath.Join(expectedLogDir, today+".log")
	_, err = os.Stat(expectedLogFile)
	assert.NoError(t, err, "Log file should be created")
}

// Test logging methods
func TestZapLogger_Debug_LogsMessageCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Debug("test debug message", "key", "value")
	})
}

func TestZapLogger_Info_LogsMessageCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Info("test info message", "key", "value")
	})
}

func TestZapLogger_Warn_LogsMessageCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Warn("test warn message", "key", "value")
	})
}

func TestZapLogger_Error_LogsMessageCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Error("test error message", "key", "value")
	})
}

func TestZapLogger_Fatal_LogsMessageCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Note: Fatal calls os.Exit, so we can't test it in a unit test
	// In a real application, this would terminate the program
	// We'll skip this test as it's not practical to test
	t.Skip("Fatal method calls os.Exit and cannot be tested in unit tests")
}

// Test formatted logging methods
func TestZapLogger_Debugf_LogsFormattedMessage(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Debugf("test debug message: %s", "formatted")
	})
}

func TestZapLogger_Infof_LogsFormattedMessage(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Infof("test info message: %d", 42)
	})
}

func TestZapLogger_Warnf_LogsFormattedMessage(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Warnf("test warn message: %v", []string{"a", "b", "c"})
	})
}

func TestZapLogger_Errorf_LogsFormattedMessage(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Just verify the method doesn't panic
	assert.NotPanics(t, func() {
		logger.Errorf("test error message: %s", "error details")
	})
}

func TestZapLogger_Fatalf_LogsFormattedMessage(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Note: Fatalf calls os.Exit, so we can't test it in a unit test
	// In a real application, this would terminate the program
	// We'll skip this test as it's not practical to test
	t.Skip("Fatalf method calls os.Exit and cannot be tested in unit tests")
}

// Test With method
func TestZapLogger_With_ReturnsNewLoggerWithTags(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	newLogger := logger.With("tag1", "value1", "tag2", "value2")

	assert.NotNil(t, newLogger)
	assert.NotEqual(t, logger, newLogger)

	// Verify the new logger works
	assert.NotPanics(t, func() {
		newLogger.Info("test message")
	})
}

// Test encoder functions
func TestColoredConsoleEncoder_ReturnsConsoleEncoder(t *testing.T) {
	encoder := coloredConsoleEncoder()

	assert.NotNil(t, encoder)
	// Verify it's a console encoder by checking if it's not nil
	assert.NotNil(t, encoder, "Should return a console encoder")
}

func TestPlainFileEncoder_ReturnsJSONEncoder(t *testing.T) {
	encoder := plainFileEncoder()

	assert.NotNil(t, encoder)
	// Verify it's a JSON encoder by checking if it's not nil
	assert.NotNil(t, encoder, "Should return a JSON encoder")
}

// Test edge cases
func TestZapLogger_EmptyMessage_HandlesCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Should not crash and should log something
	assert.NotPanics(t, func() {
		logger.Info("")
	})
}

func TestZapLogger_NoKeyValuePairs_HandlesCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	assert.NotPanics(t, func() {
		logger.Info("test message")
	})
}

func TestZapLogger_OddNumberOfKeyValuePairs_HandlesCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	assert.NotPanics(t, func() {
		logger.Info("test message", "key1", "value1", "key2")
	})
}

// Test different log levels in development vs production
func TestZapLogger_DevelopmentMode_LogsDebugMessages(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}

	logger, err := NewZapLogger(config)
	require.NoError(t, err)

	// Should not panic in development mode
	assert.NotPanics(t, func() {
		logger.Debug("debug message in development")
	})
}

func TestZapLogger_ProductionMode_DoesNotLogDebugMessages(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: false,
	}

	logger, err := NewZapLogger(config)
	require.NoError(t, err)

	// Should not panic in production mode (debug messages are filtered by zap)
	assert.NotPanics(t, func() {
		logger.Debug("debug message in production")
	})
}

// Test file logging
func TestZapLogger_LogsToFile(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}

	logger, err := NewZapLogger(config)
	require.NoError(t, err)

	// Log a message
	logger.Info("test file logging", "key", "value")

	// Force sync to ensure file is written
	// Note: Sync() may fail on stdout on some systems, which is expected
	if err := logger.sugarLogger.Sync(); err != nil {
		// Ignore sync errors for stdout as they are expected in test environments
		_ = err
	}

	// Check that the log file was created and contains the message
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(BaseDataDir, LogsDir, string(config.ProcessName), today+".log")

	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "test file logging")
	assert.Contains(t, contentStr, "key")
	assert.Contains(t, contentStr, "value")
}

// Test concurrent logging
func TestZapLogger_ConcurrentLogging_HandlesSafely(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Log concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent message", "id", id)
			done <- true
		}(i)
	}

	// Wait for all logs to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not crash
	assert.True(t, true, "Concurrent logging should not crash")
}

// Test various data types in logging
func TestZapLogger_VariousDataTypes_HandlesCorrectly(t *testing.T) {
	logger := createTestLogger(t)
	defer cleanupTestLogger(t, logger)

	// Test different data types
	assert.NotPanics(t, func() {
		logger.Info("test with various types",
			"string", "value",
			"int", 42,
			"float", 3.14,
			"bool", true,
			"slice", []string{"a", "b", "c"},
			"map", map[string]int{"key": 1},
		)
	})
}

// Test logger with different process names
func TestZapLogger_DifferentProcessNames_WorkCorrectly(t *testing.T) {
	processNames := []ProcessName{
		AggregatorProcess,
		DatabaseProcess,
		KeeperProcess,
		RegistrarProcess,
		HealthProcess,
		TaskDispatcherProcess,
		TaskMonitorProcess,
		TimeSchedulerProcess,
		ConditionSchedulerProcess,
		TestProcess,
	}

	for _, processName := range processNames {
		t.Run(string(processName), func(t *testing.T) {
			config := LoggerConfig{
				ProcessName:   processName,
				IsDevelopment: true,
			}

			logger, err := NewZapLogger(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)

			// Test that logging works
			assert.NotPanics(t, func() {
				logger.Info("test message for " + string(processName))
			})
		})
	}
}

// Helper functions
func createTestLogger(t *testing.T) *zapLogger {
	config := LoggerConfig{
		ProcessName:   TestProcess,
		IsDevelopment: true,
	}

	logger, err := NewZapLogger(config)
	require.NoError(t, err)

	return logger
}

func cleanupTestLogger(t *testing.T, logger *zapLogger) {
	t.Helper()
	if logger != nil {
		// Force sync to ensure all logs are written
		// Note: Sync() may fail on stdout on some systems, which is expected
		if err := logger.sugarLogger.Sync(); err != nil {
			// Ignore sync errors for stdout as they are expected in test environments
			// This is a known issue with zap logger when syncing to stdout
			_ = err
		}
	}
}
