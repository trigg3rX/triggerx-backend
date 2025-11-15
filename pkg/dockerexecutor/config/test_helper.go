package config

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"go.uber.org/mock/gomock"
)

// NewDefaultMockConfigProvider creates a mock config provider with default test values
func NewDefaultMockConfigProvider(t gomock.TestReporter) *MockConfigProviderInterface {
	ctrl := gomock.NewController(t)
	mock := NewMockConfigProviderInterface(ctrl)

	// Set up default test configuration
	defaultConfig := CodeExecutorConfig{
		Fees: ExecutionFeeConfig{
			PricePerTG:      0.001,
			FixedCost:       0.0001,
			TransactionCost: 0.001,
			StaticComplexityFactor: 0.001,
			DynamicComplexityFactor: 0.01,
		},
		Languages: map[string]LanguagePoolConfig{
			"go": {
				BasePoolConfig: BasePoolConfig{
					MaxContainers:       5,
					MinContainers:       1,
					MaxWaitTime:         30 * time.Second,
					HealthCheckInterval: 60 * time.Second,
				},
				DockerConfig: DockerContainerConfig{
					Image:          "golang:1.24-alpine",
					TimeoutSeconds: 30,
					AutoCleanup:    true,
					MemoryLimit:    "512m",
					CPULimit:       1.0,
					NetworkMode:    "bridge",
					SecurityOpt:    []string{"no-new-privileges"},
					ReadOnlyRootFS: true,
					Environment:    []string{"GOPATH=/go"},
					Binds:          []string{"/tmp:/tmp:ro"},
				},
				LanguageConfig: LanguageConfig{
					Language:    types.LanguageGo,
					ImageName:   "golang:1.24-alpine",
					SetupScript: "go mod init test",
					RunCommand:  "go run",
					Extensions:  []string{".go"},
				},
			},
			"python": {
				BasePoolConfig: BasePoolConfig{
					MaxContainers:       5,
					MinContainers:       1,
					MaxWaitTime:         30 * time.Second,
					HealthCheckInterval: 60 * time.Second,
				},
				DockerConfig: DockerContainerConfig{
					Image:          "python:3.11-alpine",
					TimeoutSeconds: 30,
					AutoCleanup:    true,
					MemoryLimit:    "512m",
					CPULimit:       1.0,
					NetworkMode:    "bridge",
					SecurityOpt:    []string{"no-new-privileges"},
					ReadOnlyRootFS: true,
					Environment:    []string{"PYTHONPATH=/app"},
					Binds:          []string{"/tmp:/tmp:ro"},
				},
				LanguageConfig: LanguageConfig{
					Language:    types.LanguagePy,
					ImageName:   "python:3.11-alpine",
					SetupScript: "pip install --user .",
					RunCommand:  "python",
					Extensions:  []string{".py"},
				},
			},
		},
		Cache: FileCacheConfig{
			CacheDir:          "/tmp/cache",
			MaxCacheSize:      1024 * 1024 * 100, // 100MB
			EvictionSize:      1024 * 1024 * 10,  // 10MB
			EnableCompression: true,
			MaxFileSize:       1024 * 1024 * 10, // 10MB
		},
		Validation: ValidationConfig{
			MaxFileSize:       1024 * 1024 * 10, // 10MB
			AllowedExtensions: []string{".go", ".py", ".js", ".ts"},
			MaxComplexity:     10,
			TimeoutSeconds:    30,
		},
	}

	// Set up mock expectations
	mock.EXPECT().GetConfig().Return(defaultConfig).AnyTimes()
	mock.EXPECT().GetFeesConfig().Return(defaultConfig.Fees).AnyTimes()
	mock.EXPECT().GetCacheConfig().Return(defaultConfig.Cache).AnyTimes()
	mock.EXPECT().GetValidationConfig().Return(defaultConfig.Validation).AnyTimes()
	mock.EXPECT().GetSupportedLanguages().Return([]types.Language{types.LanguageGo, types.LanguagePy, types.LanguageJS}).AnyTimes()

	// Set up language-specific config expectations
	mock.EXPECT().GetLanguagePoolConfig(types.LanguageGo).Return(defaultConfig.Languages["go"], true).AnyTimes()
	mock.EXPECT().GetLanguagePoolConfig(types.LanguagePy).Return(defaultConfig.Languages["python"], true).AnyTimes()
	mock.EXPECT().GetLanguagePoolConfig(types.LanguageJS).Return(LanguagePoolConfig{}, false).AnyTimes()

	return mock
}
