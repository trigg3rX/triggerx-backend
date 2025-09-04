package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

func TestCodeExecutorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CodeExecutorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidConfig_ShouldPass",
			config:  NewDefaultMockConfigProvider(t).GetConfig(),
			wantErr: false,
		},
		{
			name: "InvalidFeesConfig_ShouldFail",
			config: func() CodeExecutorConfig {
				cfg := NewDefaultMockConfigProvider(t).GetConfig()
				cfg.Fees.PricePerTG = -1
				return cfg
			}(),
			wantErr: true,
			errMsg:  "fees config error",
		},
		{
			name: "InvalidCacheConfig_ShouldFail",
			config: func() CodeExecutorConfig {
				cfg := NewDefaultMockConfigProvider(t).GetConfig()
				cfg.Cache.CacheDir = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "cache config error",
		},
		{
			name: "InvalidValidationConfig_ShouldFail",
			config: func() CodeExecutorConfig {
				cfg := NewDefaultMockConfigProvider(t).GetConfig()
				cfg.Validation.MaxFileSize = -1
				return cfg
			}(),
			wantErr: true,
			errMsg:  "validation config error",
		},
		{
			name: "InvalidLanguageConfig_ShouldFail",
			config: func() CodeExecutorConfig {
				cfg := NewDefaultMockConfigProvider(t).GetConfig()
				if goConfig, exists := cfg.Languages["go"]; exists {
					goConfig.LanguageConfig.Language = "invalid"
					cfg.Languages["go"] = goConfig
				}
				return cfg
			}(),
			wantErr: true,
			errMsg:  "language config 'go' error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDockerContainerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DockerContainerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
				NetworkMode:    "bridge",
				Environment:    []string{"KEY=value"},
			},
			wantErr: false,
		},
		{
			name: "EmptyImage_ShouldFail",
			config: DockerContainerConfig{
				Image:          "",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
			},
			wantErr: true,
			errMsg:  "image cannot be empty",
		},
		{
			name: "InvalidImage_ShouldFail",
			config: DockerContainerConfig{
				Image:          "invalid@image",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
			},
			wantErr: true,
			errMsg:  "invalid docker image format",
		},
		{
			name: "ZeroTimeout_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 0,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
			},
			wantErr: true,
			errMsg:  "timeout_seconds must be positive",
		},
		{
			name: "NegativeTimeout_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: -1,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
			},
			wantErr: true,
			errMsg:  "timeout_seconds must be positive",
		},
		{
			name: "InvalidMemoryLimit_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "invalid",
				CPULimit:       1.0,
			},
			wantErr: true,
			errMsg:  "invalid memory_limit format",
		},
		{
			name: "ZeroCPULimit_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       0,
			},
			wantErr: true,
			errMsg:  "cpu_limit must be positive",
		},
		{
			name: "NegativeCPULimit_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       -1.0,
			},
			wantErr: true,
			errMsg:  "cpu_limit must be positive",
		},
		{
			name: "InvalidNetworkMode_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
				NetworkMode:    "invalid",
			},
			wantErr: true,
			errMsg:  "invalid network_mode",
		},
		{
			name: "InvalidEnvironmentVariable_ShouldFail",
			config: DockerContainerConfig{
				Image:          "golang:1.21-alpine",
				TimeoutSeconds: 300,
				MemoryLimit:    "1024m",
				CPULimit:       1.0,
				Environment:    []string{"INVALID_VAR"},
			},
			wantErr: true,
			errMsg:  "invalid environment variable format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExecutionFeeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ExecutionFeeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: ExecutionFeeConfig{
				PricePerTG:      0.0001,
				FixedCost:       1.0,
				TransactionCost: 1.0,
				OverheadCost:    0.1,
			},
			wantErr: false,
		},
		{
			name: "NegativePricePerTG_ShouldFail",
			config: ExecutionFeeConfig{
				PricePerTG:      -0.1,
				FixedCost:       1.0,
				TransactionCost: 1.0,
				OverheadCost:    0.1,
			},
			wantErr: true,
			errMsg:  "price_per_tg cannot be negative",
		},
		{
			name: "NegativeFixedCost_ShouldFail",
			config: ExecutionFeeConfig{
				PricePerTG:      0.0001,
				FixedCost:       -1.0,
				TransactionCost: 1.0,
				OverheadCost:    0.1,
			},
			wantErr: true,
			errMsg:  "fixed_cost cannot be negative",
		},
		{
			name: "NegativeTransactionCost_ShouldFail",
			config: ExecutionFeeConfig{
				PricePerTG:      0.0001,
				FixedCost:       1.0,
				TransactionCost: -1.0,
				OverheadCost:    0.1,
			},
			wantErr: true,
			errMsg:  "transaction_cost cannot be negative",
		},
		{
			name: "NegativeOverheadCost_ShouldFail",
			config: ExecutionFeeConfig{
				PricePerTG:      0.0001,
				FixedCost:       1.0,
				TransactionCost: 1.0,
				OverheadCost:    -0.1,
			},
			wantErr: true,
			errMsg:  "overhead_cost cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBasePoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BasePoolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: BasePoolConfig{
				MaxContainers:       5,
				MinContainers:       2,
				MaxWaitTime:         60 * time.Second,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "ZeroMaxContainers_ShouldFail",
			config: BasePoolConfig{
				MaxContainers:       0,
				MinContainers:       2,
				MaxWaitTime:         60 * time.Second,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "max_containers must be positive",
		},
		{
			name: "NegativeMinContainers_ShouldFail",
			config: BasePoolConfig{
				MaxContainers:       5,
				MinContainers:       -1,
				MaxWaitTime:         60 * time.Second,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "min_containers cannot be negative",
		},
		{
			name: "MinContainersExceedsMax_ShouldFail",
			config: BasePoolConfig{
				MaxContainers:       5,
				MinContainers:       6,
				MaxWaitTime:         60 * time.Second,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "min_containers cannot exceed max_containers",
		},

		{
			name: "ZeroMaxWaitTime_ShouldFail",
			config: BasePoolConfig{
				MaxContainers:       5,
				MinContainers:       2,
				MaxWaitTime:         0,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "max_wait_time must be positive",
		},

		{
			name: "ZeroHealthCheckInterval_ShouldFail",
			config: BasePoolConfig{
				MaxContainers:       5,
				MinContainers:       2,
				MaxWaitTime:         60 * time.Second,
				HealthCheckInterval: 0,
			},
			wantErr: true,
			errMsg:  "health_check_interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLanguageConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LanguageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: LanguageConfig{
				Language:   types.LanguageGo,
				ImageName:  "golang:1.21-alpine",
				RunCommand: "go run code.go",
				Extensions: []string{".go"},
			},
			wantErr: false,
		},
		{
			name: "InvalidLanguage_ShouldFail",
			config: LanguageConfig{
				Language:   "invalid",
				ImageName:  "golang:1.21-alpine",
				RunCommand: "go run code.go",
				Extensions: []string{".go"},
			},
			wantErr: true,
			errMsg:  "unsupported language",
		},
		{
			name: "EmptyImageName_ShouldFail",
			config: LanguageConfig{
				Language:   types.LanguageGo,
				ImageName:  "",
				RunCommand: "go run code.go",
				Extensions: []string{".go"},
			},
			wantErr: true,
			errMsg:  "image_name cannot be empty",
		},
		{
			name: "EmptyRunCommand_ShouldFail",
			config: LanguageConfig{
				Language:   types.LanguageGo,
				ImageName:  "golang:1.21-alpine",
				RunCommand: "",
				Extensions: []string{".go"},
			},
			wantErr: true,
			errMsg:  "run_command cannot be empty",
		},
		{
			name: "EmptyExtensions_ShouldFail",
			config: LanguageConfig{
				Language:   types.LanguageGo,
				ImageName:  "golang:1.21-alpine",
				RunCommand: "go run code.go",
				Extensions: []string{},
			},
			wantErr: true,
			errMsg:  "at least one file extension must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLanguagePoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LanguagePoolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: LanguagePoolConfig{
				BasePoolConfig: BasePoolConfig{
					MaxContainers:       5,
					MinContainers:       2,
					MaxWaitTime:         60 * time.Second,
					HealthCheckInterval: 30 * time.Second,
				},
				DockerConfig: DockerContainerConfig{
					Image:          "golang:1.21-alpine",
					TimeoutSeconds: 300,
					MemoryLimit:    "1024m",
					CPULimit:       1.0,
					NetworkMode:    "bridge",
				},
				LanguageConfig: LanguageConfig{
					Language:   types.LanguageGo,
					ImageName:  "golang:1.21-alpine",
					RunCommand: "go run code.go",
					Extensions: []string{".go"},
				},
			},
			wantErr: false,
		},
		{
			name: "InvalidBasePoolConfig_ShouldFail",
			config: LanguagePoolConfig{
				BasePoolConfig: BasePoolConfig{
					MaxContainers:       0, // Invalid
					MinContainers:       2,
					MaxWaitTime:         60 * time.Second,
					HealthCheckInterval: 30 * time.Second,
				},
				DockerConfig: DockerContainerConfig{
					Image:          "golang:1.21-alpine",
					TimeoutSeconds: 300,
					MemoryLimit:    "1024m",
					CPULimit:       1.0,
					NetworkMode:    "bridge",
				},
				LanguageConfig: LanguageConfig{
					Language:   types.LanguageGo,
					ImageName:  "golang:1.21-alpine",
					RunCommand: "go run code.go",
					Extensions: []string{".go"},
				},
			},
			wantErr: true,
			errMsg:  "base pool config is invalid",
		},
		{
			name: "InvalidLanguageConfig_ShouldFail",
			config: LanguagePoolConfig{
				BasePoolConfig: BasePoolConfig{
					MaxContainers:       5,
					MinContainers:       2,
					MaxWaitTime:         60 * time.Second,
					HealthCheckInterval: 30 * time.Second,
				},
				DockerConfig: DockerContainerConfig{
					Image:          "golang:1.21-alpine",
					TimeoutSeconds: 300,
					MemoryLimit:    "1024m",
					CPULimit:       1.0,
					NetworkMode:    "bridge",
				},
				LanguageConfig: LanguageConfig{
					Language:   types.LanguageGo,
					ImageName:  "", // Invalid
					RunCommand: "go run code.go",
					Extensions: []string{".go"},
				},
			},
			wantErr: true,
			errMsg:  "language-specific config is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFileCacheConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  FileCacheConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: FileCacheConfig{
				CacheDir:     "data/cache",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "ValidAbsolutePath_ShouldPass",
			config: FileCacheConfig{
				CacheDir:     "/tmp/cache",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "EmptyCacheDir_ShouldFail",
			config: FileCacheConfig{
				CacheDir:     "",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "cache_dir cannot be empty",
		},
		{
			name: "InvalidCacheDir_ShouldFail",
			config: FileCacheConfig{
				CacheDir:     "invalid/path",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "cache_dir should be an absolute path or relative to 'data/'",
		},
		{
			name: "ZeroMaxCacheSize_ShouldFail",
			config: FileCacheConfig{
				CacheDir:     "data/cache",
				MaxCacheSize: 0,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "max_cache_size must be positive",
		},
		{
			name: "EvictionSizeExceedsMax_ShouldFail",
			config: FileCacheConfig{
				CacheDir:     "data/cache",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 100 * 1024 * 1024,
				MaxFileSize:  1 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "eviction_size must be less than max_cache_size",
		},
		{
			name: "ZeroMaxFileSize_ShouldFail",
			config: FileCacheConfig{
				CacheDir:     "data/cache",
				MaxCacheSize: 100 * 1024 * 1024,
				EvictionSize: 25 * 1024 * 1024,
				MaxFileSize:  0,
			},
			wantErr: true,
			errMsg:  "max_file_size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ValidationConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidConfig_ShouldPass",
			config: ValidationConfig{
				MaxFileSize:       1 * 1024 * 1024,
				AllowedExtensions: []string{".go", ".py"},
				MaxComplexity:     10,
				TimeoutSeconds:    30,
			},
			wantErr: false,
		},
		{
			name: "ZeroMaxFileSize_ShouldFail",
			config: ValidationConfig{
				MaxFileSize:       0,
				AllowedExtensions: []string{".go", ".py"},
				MaxComplexity:     10,
				TimeoutSeconds:    30,
			},
			wantErr: true,
			errMsg:  "max_file_size must be positive",
		},
		{
			name: "InvalidExtension_ShouldFail",
			config: ValidationConfig{
				MaxFileSize:       1 * 1024 * 1024,
				AllowedExtensions: []string{"go", ".py"}, // Missing dot
				MaxComplexity:     10,
				TimeoutSeconds:    30,
			},
			wantErr: true,
			errMsg:  "extension must start with a dot",
		},
		{
			name: "InvalidRegex_ShouldFail",
			config: ValidationConfig{
				MaxFileSize:       1 * 1024 * 1024,
				AllowedExtensions: []string{".go", ".py"},
				MaxComplexity:     10,
				TimeoutSeconds:    30,
			},
			wantErr: true,
			errMsg:  "invalid regex",
		},
		{
			name: "ZeroTimeout_ShouldFail",
			config: ValidationConfig{
				MaxFileSize:       1 * 1024 * 1024,
				AllowedExtensions: []string{".go", ".py"},
				MaxComplexity:     10,
				TimeoutSeconds:    0,
			},
			wantErr: true,
			errMsg:  "timeout_seconds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsValidDockerImage(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected bool
	}{
		{"ValidImage_ShouldPass", "golang:1.21-alpine", true},
		{"ValidImageWithRegistry_ShouldPass", "dockerexecutor.io/golang:1.21-alpine", true},
		{"ValidImageWithTag_ShouldPass", "golang:latest", true},
		{"ValidImageWithoutTag_ShouldPass", "golang", true},
		{"EmptyImage_ShouldFail", "", false},
		{"InvalidImage_ShouldFail", "invalid@image", false},
		{"InvalidImageWithSpecialChars_ShouldFail", "golang:1.21@alpine", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDockerImage(tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidNetworkMode(t *testing.T) {
	validModes := []string{"bridge", "host", "none", "container:", "default"}

	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"ValidBridge_ShouldPass", "bridge", true},
		{"ValidHost_ShouldPass", "host", true},
		{"ValidNone_ShouldPass", "none", true},
		{"ValidContainer_ShouldPass", "container:abc123", true},
		{"ValidDefault_ShouldPass", "default", true},
		{"InvalidMode_ShouldFail", "invalid", false},
		{"EmptyMode_ShouldFail", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNetworkMode(tt.mode, validModes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language types.Language
		expected bool
	}{
		{"ValidGo_ShouldPass", types.LanguageGo, true},
		{"ValidPy_ShouldPass", types.LanguagePy, true},
		{"ValidJS_ShouldPass", types.LanguageJS, true},
		{"ValidTS_ShouldPass", types.LanguageTS, true},
		{"ValidNode_ShouldPass", types.LanguageNode, true},
		{"InvalidLanguage_ShouldFail", "invalid", false},
		{"EmptyLanguage_ShouldFail", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLanguage(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}
