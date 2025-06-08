package logging

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"testing"
// )

// func TestLogRotationAndShutdown(t *testing.T) {
// 	// Create a temporary directory for test logs
// 	tempDir := "data/logs"
// 	logDir := filepath.Join(tempDir, "logs", "test")
// 	// os.RemoveAll(logDir) // Clean up before test

// 	// Initialize logger with test config
// 	config := LoggerConfig{
// 		LogDir:      tempDir,
// 		ProcessName: "test",
// 		Environment: Development,
// 		UseColors:   true,
// 	}

// 	if err := InitServiceLogger(config); err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}

// 	logger := GetServiceLogger()

// 	// Get the underlying ZapLogger
// 	zl, ok := logger.(*ZapLogger)
// 	if !ok {
// 		t.Fatal("Failed to get ZapLogger instance")
// 	}

// 	// Day 1: Write some logs
// 	zl.mu.Lock()
// 	zl.currentDay = "2025-05-13"
// 	zl.mu.Unlock()

// 	if err := zl.initLogger(); err != nil {
// 		t.Fatalf("Failed to initialize logger for day 1: %v", err)
// 	}
// 	logger.Info("Log entry for day 1")

// 	// Verify day 1 file exists
// 	if !fileExists(filepath.Join(logDir, "2025-05-13.log")) {
// 		t.Error("Day 1 log file was not created")
// 	}

// 	// Day 2: Write some logs
// 	zl.mu.Lock()
// 	zl.currentDay = "2025-05-14"
// 	zl.mu.Unlock()

// 	if err := zl.initLogger(); err != nil {
// 		t.Fatalf("Failed to initialize logger for day 2: %v", err)
// 	}
// 	logger.Info("Log entry for day 2")

// 	// Verify both files exist
// 	if !fileExists(filepath.Join(logDir, "2025-05-14.log")) {
// 		t.Error("Day 2 log file was not created")
// 	}

// 	// Day 3: Write some logs
// 	zl.mu.Lock()
// 	zl.currentDay = "2025-05-15"
// 	zl.mu.Unlock()

// 	if err := zl.initLogger(); err != nil {
// 		t.Fatalf("Failed to initialize logger for day 3: %v", err)
// 	}
// 	logger.Info("Log entry for day 3")

// 	// Verify all files exist
// 	if !fileExists(filepath.Join(logDir, "2025-05-15.log")) {
// 		t.Error("Day 3 log file was not created")
// 	}

// 	// Test shutdown
// 	if err := Shutdown(); err != nil {
// 		t.Fatalf("Failed to shutdown logger: %v", err)
// 	}

// 	// Verify all log files exist and have content
// 	expectedFiles := []string{
// 		"2025-05-13.log",
// 		"2025-05-14.log",
// 		"2025-05-15.log",
// 	}

// 	for _, fileName := range expectedFiles {
// 		filePath := filepath.Join(logDir, fileName)
// 		if !fileExists(filePath) {
// 			t.Errorf("Expected log file %s was not created", fileName)
// 			continue
// 		}

// 		info, err := os.Stat(filePath)
// 		if err != nil {
// 			t.Errorf("Failed to stat file %s: %v", fileName, err)
// 			continue
// 		}

// 		if info.Size() == 0 {
// 			t.Errorf("Log file %s is empty", fileName)
// 		}
// 	}
// }

// func TestLogLevelColors(t *testing.T) {
// 	// Initialize logger with test config and colors enabled
// 	config := LoggerConfig{
// 		LogDir:      "data/logs",
// 		ProcessName: "test",
// 		Environment: Development,
// 		UseColors:   true,
// 	}

// 	if err := InitServiceLogger(config); err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}

// 	logger := GetServiceLogger()
// 	zl, ok := logger.(*ZapLogger)
// 	if !ok {
// 		t.Fatal("Failed to get ZapLogger instance")
// 	}

// 	// Test cases for each log level
// 	testCases := []struct {
// 		level    string
// 		message  string
// 		expected string
// 	}{
// 		{
// 			level:    "debug",
// 			message:  "test debug",
// 			expected: fmt.Sprintf("[%sdebug%s] test debug", colorBlue, colorReset),
// 		},
// 		{
// 			level:    "info",
// 			message:  "test info",
// 			expected: fmt.Sprintf("[%sinfo%s] test info", colorGreen, colorReset),
// 		},
// 		{
// 			level:    "warn",
// 			message:  "test warn",
// 			expected: fmt.Sprintf("[%swarn%s] test warn", colorYellow, colorReset),
// 		},
// 		{
// 			level:    "error",
// 			message:  "test error",
// 			expected: fmt.Sprintf("[%serror%s] test error", colorRed, colorReset),
// 		},
// 		{
// 			level:    "fatal",
// 			message:  "test fatal",
// 			expected: fmt.Sprintf("[%sfatal%s] test fatal", colorPurple, colorReset),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.level, func(t *testing.T) {
// 			result := zl.colorize(tc.level, tc.message)
// 			if result != tc.expected {
// 				t.Errorf("colorize(%q, %q) = %q, want %q", tc.level, tc.message, result, tc.expected)
// 			}
// 		})
// 	}

// 	// Test with colors disabled
// 	zl.useColors = false
// 	for _, tc := range testCases {
// 		t.Run(tc.level+"_no_color", func(t *testing.T) {
// 			result := zl.colorize(tc.level, tc.message)
// 			if result != tc.message {
// 				t.Errorf("colorize(%q, %q) with colors disabled = %q, want %q", tc.level, tc.message, result, tc.message)
// 			}
// 		})
// 	}
// }

// func fileExists(path string) bool {
// 	_, err := os.Stat(path)
// 	return err == nil
// }
