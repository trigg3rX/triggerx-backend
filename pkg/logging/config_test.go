package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// Test LoggerConfig struct
func TestLoggerConfig_StructFields_AreAccessible(t *testing.T) {
	config := LoggerConfig{
		ProcessName:   AggregatorProcess,
		IsDevelopment: true,
	}

	assert.Equal(t, AggregatorProcess, config.ProcessName)
	assert.True(t, config.IsDevelopment)
}

func TestLoggerConfig_ZeroValue_IsValid(t *testing.T) {
	var config LoggerConfig

	assert.Equal(t, ProcessName(""), config.ProcessName)
	assert.False(t, config.IsDevelopment)
}

func TestLoggerConfig_StructLiteral_WorksCorrectly(t *testing.T) {
	tests := []struct {
		name           string
		processName    ProcessName
		isDevelopment  bool
		expectedConfig LoggerConfig
	}{
		{
			name:          "development config",
			processName:   KeeperProcess,
			isDevelopment: true,
			expectedConfig: LoggerConfig{
				ProcessName:   KeeperProcess,
				IsDevelopment: true,
			},
		},
		{
			name:          "production config",
			processName:   DatabaseProcess,
			isDevelopment: false,
			expectedConfig: LoggerConfig{
				ProcessName:   DatabaseProcess,
				IsDevelopment: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggerConfig{
				ProcessName:   tt.processName,
				IsDevelopment: tt.isDevelopment,
			}

			assert.Equal(t, tt.expectedConfig, config)
		})
	}
}

// Test getLogLevel function (from zap_logger.go but tested here since it uses config)
func TestGetLogLevel_DevelopmentMode_ReturnsDebugLevel(t *testing.T) {
	level := getLogLevel(true)
	assert.Equal(t, zapcore.DebugLevel, level)
}

func TestGetLogLevel_ProductionMode_ReturnsInfoLevel(t *testing.T) {
	level := getLogLevel(false)
	assert.Equal(t, zapcore.InfoLevel, level)
}

func TestGetLogLevel_TableDriven_ReturnsCorrectLevels(t *testing.T) {
	tests := []struct {
		name          string
		isDevelopment bool
		expectedLevel zapcore.Level
	}{
		{"development mode", true, zapcore.DebugLevel},
		{"production mode", false, zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := getLogLevel(tt.isDevelopment)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

// Test customColorLevelEncoder function
func TestCustomColorLevelEncoder_AllLevels_ReturnCorrectColors(t *testing.T) {
	tests := []struct {
		name          string
		level         zapcore.Level
		expectedColor string
		expectedLevel string
	}{
		{"debug level", zapcore.DebugLevel, colorBlue, "DBG"},
		{"info level", zapcore.InfoLevel, colorGreen, "INF"},
		{"warn level", zapcore.WarnLevel, colorYellow, "WRN"},
		{"error level", zapcore.ErrorLevel, colorRed, "ERR"},
		{"fatal level", zapcore.FatalLevel, colorMagenta, "FTL"},
		{"unknown level", zapcore.Level(99), colorWhite, "???"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock encoder to capture the output
			var captured string
			mockEncoder := &mockPrimitiveArrayEncoder{
				appendStringFunc: func(s string) {
					captured = s
				},
			}

			customColorLevelEncoder(tt.level, mockEncoder)

			// Verify the output contains the expected color and level
			assert.Contains(t, captured, tt.expectedColor)
			assert.Contains(t, captured, tt.expectedLevel)
			assert.Contains(t, captured, colorReset)
		})
	}
}

// Mock encoder for testing customColorLevelEncoder
type mockPrimitiveArrayEncoder struct {
	appendStringFunc func(string)
}

func (m *mockPrimitiveArrayEncoder) AppendString(s string) {
	if m.appendStringFunc != nil {
		m.appendStringFunc(s)
	}
}

func (m *mockPrimitiveArrayEncoder) AppendBool(v bool)                   {}
func (m *mockPrimitiveArrayEncoder) AppendByteString(v []byte)           {}
func (m *mockPrimitiveArrayEncoder) AppendComplex128(v complex128)       {}
func (m *mockPrimitiveArrayEncoder) AppendComplex64(v complex64)         {}
func (m *mockPrimitiveArrayEncoder) AppendFloat64(v float64)             {}
func (m *mockPrimitiveArrayEncoder) AppendFloat32(v float32)             {}
func (m *mockPrimitiveArrayEncoder) AppendInt(v int)                     {}
func (m *mockPrimitiveArrayEncoder) AppendInt64(v int64)                 {}
func (m *mockPrimitiveArrayEncoder) AppendInt32(v int32)                 {}
func (m *mockPrimitiveArrayEncoder) AppendInt16(v int16)                 {}
func (m *mockPrimitiveArrayEncoder) AppendInt8(v int8)                   {}
func (m *mockPrimitiveArrayEncoder) AppendUint(v uint)                   {}
func (m *mockPrimitiveArrayEncoder) AppendUint64(v uint64)               {}
func (m *mockPrimitiveArrayEncoder) AppendUint32(v uint32)               {}
func (m *mockPrimitiveArrayEncoder) AppendUint16(v uint16)               {}
func (m *mockPrimitiveArrayEncoder) AppendUint8(v uint8)                 {}
func (m *mockPrimitiveArrayEncoder) AppendUintptr(v uintptr)             {}
func (m *mockPrimitiveArrayEncoder) AppendReflected(v interface{}) error { return nil }

// Test ProcessName validation scenarios
func TestProcessName_ValidationScenarios_HandleEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		process  ProcessName
		expected string
	}{
		{"empty string", ProcessName(""), ""},
		{"single character", ProcessName("a"), "a"},
		{"special characters", ProcessName("process-name_123"), "process-name_123"},
		{"unicode characters", ProcessName("process-测试"), "process-测试"},
		{"very long name", ProcessName("very-long-process-name-that-exceeds-normal-length"), "very-long-process-name-that-exceeds-normal-length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.process)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test LoggerConfig edge cases
func TestLoggerConfig_EdgeCases_HandleCorrectly(t *testing.T) {
	tests := []struct {
		name         string
		config       LoggerConfig
		expectedName ProcessName
		expectedDev  bool
	}{
		{
			name: "empty process name",
			config: LoggerConfig{
				ProcessName:   ProcessName(""),
				IsDevelopment: true,
			},
			expectedName: ProcessName(""),
			expectedDev:  true,
		},
		{
			name:         "zero value config",
			config:       LoggerConfig{},
			expectedName: ProcessName(""),
			expectedDev:  false,
		},
		{
			name: "custom process name",
			config: LoggerConfig{
				ProcessName:   ProcessName("custom-process"),
				IsDevelopment: false,
			},
			expectedName: ProcessName("custom-process"),
			expectedDev:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedName, tt.config.ProcessName)
			assert.Equal(t, tt.expectedDev, tt.config.IsDevelopment)
		})
	}
}

func TestGetBaseDataDir_WithEnvVar_ReturnsEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "custom path",
			envValue: "/custom/data/path",
			expected: "/custom/data/path",
		},
		{
			name:     "relative path",
			envValue: "./test-data",
			expected: "./test-data",
		},
		{
			name:     "empty string",
			envValue: "",
			expected: "", // Will fall through to other logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("TRIGGERX_DATA_DIR", tt.envValue)
			require.NoError(t, err)
			defer func() {
				if err := os.Unsetenv("TRIGGERX_DATA_DIR"); err != nil {
					t.Errorf("Failed to unset environment variable: %v", err)
				}
			}()

			result := getBaseDataDir()
			if tt.envValue != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// If env var is empty, should fall through to other logic
				assert.NotEqual(t, "", result)
			}
		})
	}
}

func TestGetBaseDataDir_WithGoMod_ReturnsProjectDataDir(t *testing.T) {
	// Ensure no env var is set
	err := os.Unsetenv("TRIGGERX_DATA_DIR")
	require.NoError(t, err)

	// Create a temporary directory structure with go.mod
	tempDir, err := os.MkdirTemp("", "test-project-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create go.mod file
	goModPath := filepath.Join(tempDir, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	result := getBaseDataDir()
	expected := filepath.Join(tempDir, "data")
	assert.Equal(t, expected, result)
}

func TestGetBaseDataDir_NoGoMod_ReturnsDefault(t *testing.T) {
	// Ensure no env var is set
	err := os.Unsetenv("TRIGGERX_DATA_DIR")
	require.NoError(t, err)

	// Create a temporary directory without go.mod
	tempDir, err := os.MkdirTemp("", "test-no-gomod-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create nested directory structure without go.mod
	nestedDir := filepath.Join(tempDir, "nested", "deep", "path")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Change to nested directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change directory: %v", err)
		}
	}()

	err = os.Chdir(nestedDir)
	require.NoError(t, err)

	result := getBaseDataDir()
	assert.Equal(t, "data", result)
}

func TestGetBaseDataDir_GoModInParentDirectory_ReturnsParentDataDir(t *testing.T) {
	// Ensure no env var is set
	err := os.Unsetenv("TRIGGERX_DATA_DIR")
	require.NoError(t, err)

	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-parent-gomod-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create go.mod in root
	goModPath := filepath.Join(tempDir, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	require.NoError(t, err)

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "pkg", "logging")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Change to nested directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change directory: %v", err)
		}
	}()

	err = os.Chdir(nestedDir)
	require.NoError(t, err)

	result := getBaseDataDir()
	expected := filepath.Join(tempDir, "data")
	assert.Equal(t, expected, result)
}

func TestGetBaseDataDir_TableDriven_HandlesAllScenarios(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change directory: %v", err)
		}
	}()

	tests := []struct {
		name           string
		envVar         string
		hasGoMod       bool
		goModLevel     int    // 0 = current dir, 1 = parent, 2 = grandparent
		expectedSuffix string // suffix to check since full path varies
	}{
		{
			name:           "env var takes precedence",
			envVar:         "/priority/path",
			hasGoMod:       true,
			goModLevel:     0,
			expectedSuffix: "/priority/path",
		},
		{
			name:           "go.mod in current directory",
			envVar:         "",
			hasGoMod:       true,
			goModLevel:     0,
			expectedSuffix: "/data",
		},
		{
			name:           "go.mod in parent directory",
			envVar:         "",
			hasGoMod:       true,
			goModLevel:     1,
			expectedSuffix: "/data",
		},
		{
			name:           "no go.mod found",
			envVar:         "",
			hasGoMod:       false,
			goModLevel:     0,
			expectedSuffix: "data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envVar != "" {
				err := os.Setenv("TRIGGERX_DATA_DIR", tt.envVar)
				require.NoError(t, err)
				defer func() {
					if err := os.Unsetenv("TRIGGERX_DATA_DIR"); err != nil {
						t.Errorf("Failed to unset environment variable: %v", err)
					}
				}()
			} else {
				err := os.Unsetenv("TRIGGERX_DATA_DIR")
				require.NoError(t, err)
			}

			// Create temp directory
			tempDir, err := os.MkdirTemp("", "test-table-*")
			require.NoError(t, err)
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Errorf("Failed to remove temporary directory: %v", err)
				}
			}()

			// Set up go.mod if needed
			if tt.hasGoMod {
				var goModDir string
				switch tt.goModLevel {
				case 0:
					goModDir = tempDir
				case 1:
					goModDir = filepath.Dir(tempDir)
				case 2:
					goModDir = filepath.Dir(filepath.Dir(tempDir))
				}

				if tt.goModLevel > 0 {
					// For parent directories, create a proper structure
					projectRoot, err := os.MkdirTemp("", "test-project-*")
					require.NoError(t, err)
					defer func() {
						if err := os.RemoveAll(projectRoot); err != nil {
							t.Errorf("Failed to remove temporary directory: %v", err)
						}
					}()

					goModPath := filepath.Join(projectRoot, "go.mod")
					err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
					require.NoError(t, err)

					// Create nested directory
					workDir := filepath.Join(projectRoot, "nested")
					if tt.goModLevel == 2 {
						workDir = filepath.Join(projectRoot, "nested", "deep")
					}
					err = os.MkdirAll(workDir, 0755)
					require.NoError(t, err)

					err = os.Chdir(workDir)
					require.NoError(t, err)
				} else {
					goModPath := filepath.Join(goModDir, "go.mod")
					err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
					require.NoError(t, err)

					err = os.Chdir(tempDir)
					require.NoError(t, err)
				}
			} else {
				err = os.Chdir(tempDir)
				require.NoError(t, err)
			}

			result := getBaseDataDir()

			if tt.envVar != "" {
				assert.Equal(t, tt.envVar, result)
			} else {
				assert.True(t, strings.HasSuffix(result, tt.expectedSuffix),
					"Expected result '%s' to end with '%s'", result, tt.expectedSuffix)
			}
		})
	}
}
