package config

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

func DefaultConfig(lang string) ExecutorConfig {
	return ExecutorConfig{
		Docker:     DefaultDockerConfig(lang),
		Fees:       DefaultFeeConfig(),
		BasePool:   DefaultBasePoolConfig(),
		Cache:      DefaultCacheConfig(),
		Validation: DefaultValidationConfig(),
	}
}

func DefaultDockerConfig(lang string) DockerConfig {
	var imageName string
	switch lang {
	case "go":
		imageName = "golang:latest"
	case "py":
		imageName = "python:latest"
	case "js", "ts", "node":
		imageName = "node:latest"
	default:
		imageName = "golang:latest"
	}

	return DockerConfig{
		Image:          imageName,
		TimeoutSeconds: 600,
		AutoCleanup:    true,
		MemoryLimit:    "1024m",
		CPULimit:       1.0,
		NetworkMode:    "bridge",
		SecurityOpt:    []string{"no-new-privileges"},
		ReadOnlyRootFS: false,
		Environment:    []string{"GODEBUG=http2client=0"},
		Binds:          []string{"/var/run/docker.sock:/var/run/docker.sock"},
	}
}

// GetLanguageConfig returns language-specific configuration
func GetLanguageConfig(lang types.Language) LanguageConfig {
	switch lang {
	case types.LanguageGo:
		return LanguageConfig{
			Language:    types.LanguageGo,
			ImageName:   "golang:latest",
			SetupScript: getGoSetupScript(),
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
		}
	case types.LanguagePy:
		return LanguageConfig{
			Language:    types.LanguagePy,
			ImageName:   "python:latest",
			SetupScript: getPythonSetupScript(),
			RunCommand:  "python code.py",
			Extensions:  []string{".py"},
		}
	case types.LanguageJS:
		return LanguageConfig{
			Language:    types.LanguageJS,
			ImageName:   "node:latest",
			SetupScript: getJavaScriptSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".js"},
		}
	case types.LanguageTS:
		return LanguageConfig{
			Language:    types.LanguageTS,
			ImageName:   "node:latest",
			SetupScript: getTypeScriptSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".ts"},
		}
	case types.LanguageNode:
		return LanguageConfig{
			Language:    types.LanguageNode,
			ImageName:   "node:latest",
			SetupScript: getNodeSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".js", ".mjs", ".cjs"},
		}
	default:
		return LanguageConfig{
			Language:    types.LanguageGo,
			ImageName:   "golang:latest",
			SetupScript: getGoSetupScript(),
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
		}
	}
}

// GetLanguagePoolConfig returns pool configuration for a specific language
func GetLanguagePoolConfig(lang types.Language) LanguagePoolConfig {
	baseConfig := DefaultBasePoolConfig()
	return LanguagePoolConfig{
		Language:      lang,
		BasePoolConfig: baseConfig,
		Config:        GetLanguageConfig(lang),
	}
}

// GetSupportedLanguages returns all supported languages
func GetSupportedLanguages() []types.Language {
	return []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}
}

func getGoSetupScript() string {
	return `#!/bin/sh
cd /code
go mod init code
go mod tidy
echo "START_EXECUTION"
go run code.go 2>&1 || {
    echo "Error executing Go program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func getPythonSetupScript() string {
	return `#!/bin/sh
cd /code
echo "START_EXECUTION"
python code.py 2>&1 || {
    echo "Error executing Python program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func getJavaScriptSetupScript() string {
	return `#!/bin/sh
cd /code
echo "START_EXECUTION"
node code.js 2>&1 || {
    echo "Error executing JavaScript program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func getTypeScriptSetupScript() string {
	return `#!/bin/sh
cd /code
npm install -g typescript
echo "START_EXECUTION"
tsc code.ts && node code.js 2>&1 || {
    echo "Error executing TypeScript program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func getNodeSetupScript() string {
	return `#!/bin/sh
cd /code
echo "START_EXECUTION"
node code.js 2>&1 || {
    echo "Error executing Node.js program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		PricePerTG:            0.0001,
		FixedCost:             1.0,
		TransactionSimulation: 1.0,
		OverheadCost:          0.1,
	}
}

func DefaultBasePoolConfig() BasePoolConfig {
	return BasePoolConfig{
		MaxContainers:       2,
		MinContainers:       1,
		IdleTimeout:         50 * time.Minute,
		PreWarmCount:        2,
		MaxWaitTime:         100 * time.Second,
		CleanupInterval:     50 * time.Minute,
		HealthCheckInterval: 1 * time.Minute,
	}
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		CleanupInterval:   10 * time.Minute,
		EnableCompression: true,
		MaxFileSize:       1 * 1024 * 1024, // 1MB
	}
}

func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		EnableCodeValidation: true,
		MaxFileSize:          1 * 1024 * 1024, // 1MB
		AllowedExtensions:    []string{".go", ".py", ".js", ".ts"},
		BlockedPatterns: []string{
			"os.RemoveAll",
			"exec.Command",
			"syscall",
			"runtime.GC",
			"panic(",
		},
		TimeoutSeconds: 30,
	}
}

// OptimizedConfig returns a configuration optimized for high-performance execution
func OptimizedConfig(lang string) ExecutorConfig {
	cfg := DefaultConfig(lang)

	// Optimize pool settings for high throughput
	cfg.BasePool.MaxContainers = 10
	cfg.BasePool.MinContainers = 5
	cfg.BasePool.PreWarmCount = 8
	cfg.BasePool.IdleTimeout = 5 * time.Minute

	// Optimize cache settings
	cfg.Cache.MaxCacheSize = 500 * 1024 * 1024 // 500MB

	// Optimize Docker settings
	cfg.Docker.MemoryLimit = "2048m"
	cfg.Docker.CPULimit = 2.0

	return cfg
}

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig(lang string) ExecutorConfig {
	cfg := DefaultConfig(lang)

	// Reduce resource usage for development
	cfg.BasePool.MaxContainers = 2
	cfg.BasePool.MinContainers = 1
	cfg.BasePool.PreWarmCount = 1

	// Shorter timeouts for faster feedback
	cfg.BasePool.IdleTimeout = 2 * time.Minute

	return cfg
}

func (c *DockerConfig) MemoryLimitBytes() uint64 {
	memoryLimit, err := units.RAMInBytes(c.MemoryLimit)
	if err != nil {
		return 0
	}
	return uint64(memoryLimit)
}

func (c *DockerConfig) ToContainerResources() container.Resources {
	return container.Resources{
		Memory:   int64(c.MemoryLimitBytes()),
		NanoCPUs: int64(c.CPULimit * 1e9),
	}
}
