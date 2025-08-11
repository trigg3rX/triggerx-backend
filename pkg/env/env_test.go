package env

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvString_ExistingVariable_ReturnsValue(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{"simple string", "TEST_STRING", "hello world", "default", "hello world"},
		{"empty string", "TEST_EMPTY", "", "default", ""},
		{"special characters", "TEST_SPECIAL", "!@#$%^&*()", "default", "!@#$%^&*()"},
		{"numbers as string", "TEST_NUMBERS", "12345", "default", "12345"},
		{"whitespace", "TEST_WHITESPACE", "  spaced  ", "default", "  spaced  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvString(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvString(%s, %s) = %s, want %s", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvString_MissingVariable_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		defaultValue string
		expected     string
	}{
		{"simple default", "NONEXISTENT_STRING", "default value", "default value"},
		{"empty default", "NONEXISTENT_EMPTY", "", ""},
		{"special default", "NONEXISTENT_SPECIAL", "!@#$%^&*()", "!@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure environment variable is not set
			if err := os.Unsetenv(tt.envKey); err != nil {
				t.Errorf("Failed to unset environment variable: %v", err)
			}

			result := GetEnvString(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvString(%s, %s) = %s, want %s", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool_ValidValues_ReturnsCorrectBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"true value", "TEST_BOOL_TRUE", "true", false, true},
		{"false value", "TEST_BOOL_FALSE", "false", true, false},
		{"1 as true", "TEST_BOOL_ONE", "1", false, true},
		{"0 as false", "TEST_BOOL_ZERO", "0", true, false},
		{"TRUE uppercase", "TEST_BOOL_UPPER", "TRUE", false, true},
		{"FALSE uppercase", "TEST_BOOL_FALSE_UPPER", "FALSE", true, false},
		{"True mixed case", "TEST_BOOL_MIXED", "True", false, true},
		{"False mixed case", "TEST_BOOL_FALSE_MIXED", "False", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvBool(%s, %t) = %t, want %t", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool_InvalidValues_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"invalid string", "TEST_BOOL_INVALID", "invalid", true, true},
		{"empty string", "TEST_BOOL_EMPTY", "", false, false},
		{"random text", "TEST_BOOL_RANDOM", "random text", true, true},
		{"number 2", "TEST_BOOL_TWO", "2", false, false},
		{"negative number", "TEST_BOOL_NEGATIVE", "-1", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvBool(%s, %t) = %t, want %t", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool_MissingVariable_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		defaultValue bool
		expected     bool
	}{
		{"true default", "NONEXISTENT_BOOL", true, true},
		{"false default", "NONEXISTENT_BOOL_FALSE", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure environment variable is not set
			if err := os.Unsetenv(tt.envKey); err != nil {
				t.Errorf("Failed to unset environment variable: %v", err)
			}

			result := GetEnvBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvBool(%s, %t) = %t, want %t", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt_ValidValues_ReturnsCorrectInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"positive number", "TEST_INT_POSITIVE", "123", 0, 123},
		{"zero", "TEST_INT_ZERO", "0", 999, 0},
		{"negative number", "TEST_INT_NEGATIVE", "-456", 0, -456},
		{"large number", "TEST_INT_LARGE", "2147483647", 0, 2147483647},
		{"small number", "TEST_INT_SMALL", "-2147483648", 0, -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvInt(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvInt(%s, %d) = %d, want %d", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt_InvalidValues_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"invalid string", "TEST_INT_INVALID", "not a number", 999, 999},
		{"empty string", "TEST_INT_EMPTY", "", 123, 123},
		{"float number", "TEST_INT_FLOAT", "123.45", 999, 999},
		{"hex number", "TEST_INT_HEX", "0xFF", 999, 999},
		{"boolean string", "TEST_INT_BOOL", "true", 999, 999},
		{"special characters", "TEST_INT_SPECIAL", "!@#$%", 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvInt(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvInt(%s, %d) = %d, want %d", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt_MissingVariable_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		defaultValue int
		expected     int
	}{
		{"positive default", "NONEXISTENT_INT", 123, 123},
		{"zero default", "NONEXISTENT_INT_ZERO", 0, 0},
		{"negative default", "NONEXISTENT_INT_NEGATIVE", -456, -456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure environment variable is not set
			if err := os.Unsetenv(tt.envKey); err != nil {
				t.Errorf("Failed to unset environment variable: %v", err)
			}

			result := GetEnvInt(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvInt(%s, %d) = %d, want %d", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration_ValidValues_ReturnsCorrectDuration(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"seconds", "TEST_DURATION_SECONDS", "30s", time.Minute, 30 * time.Second},
		{"minutes", "TEST_DURATION_MINUTES", "5m", time.Second, 5 * time.Minute},
		{"hours", "TEST_DURATION_HOURS", "2h", time.Second, 2 * time.Hour},
		{"milliseconds", "TEST_DURATION_MS", "500ms", time.Second, 500 * time.Millisecond},
		{"microseconds", "TEST_DURATION_US", "1000us", time.Second, 1000 * time.Microsecond},
		{"nanoseconds", "TEST_DURATION_NS", "1000000ns", time.Second, 1000000 * time.Nanosecond},
		{"zero duration", "TEST_DURATION_ZERO", "0s", time.Minute, 0},
		{"complex duration", "TEST_DURATION_COMPLEX", "1h30m45s", time.Second, 1*time.Hour + 30*time.Minute + 45*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvDuration(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvDuration(%s, %v) = %v, want %v", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration_InvalidValues_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"invalid string", "TEST_DURATION_INVALID", "not a duration", time.Minute, time.Minute},
		{"empty string", "TEST_DURATION_EMPTY", "", time.Second, time.Second},
		{"number without unit", "TEST_DURATION_NO_UNIT", "30", time.Minute, time.Minute},
		{"invalid unit", "TEST_DURATION_BAD_UNIT", "30x", time.Minute, time.Minute},
		{"boolean string", "TEST_DURATION_BOOL", "true", time.Minute, time.Minute},
		{"special characters", "TEST_DURATION_SPECIAL", "!@#$%", time.Minute, time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				t.Errorf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := GetEnvDuration(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvDuration(%s, %v) = %v, want %v", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration_MissingVariable_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"positive default", "NONEXISTENT_DURATION", 30 * time.Second, 30 * time.Second},
		{"zero default", "NONEXISTENT_DURATION_ZERO", 0, 0},
		{"complex default", "NONEXISTENT_DURATION_COMPLEX", 1*time.Hour + 30*time.Minute, 1*time.Hour + 30*time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure environment variable is not set
			if err := os.Unsetenv(tt.envKey); err != nil {
				t.Errorf("Failed to unset environment variable: %v", err)
			}

			result := GetEnvDuration(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvDuration(%s, %v) = %v, want %v", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
