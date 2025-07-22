package config

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/scripts"
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
		imageName = "golang:1.21-alpine" // Use Alpine for faster startup
	case "py":
		imageName = "python:3.12-alpine" // Use Alpine for faster startup
	case "js", "ts", "node":
		imageName = "node:22-alpine" // Use Alpine for faster startup
	default:
		imageName = "golang:1.21-alpine"
	}

	return DockerConfig{
		Image:          imageName,
		TimeoutSeconds: 300, // Reduced timeout
		AutoCleanup:    true,
		MemoryLimit:    "1024m",
		CPULimit:       1.0,
		NetworkMode:    "bridge",
		SecurityOpt:    []string{"no-new-privileges"},
		ReadOnlyRootFS: false,
		Environment: []string{
			"GODEBUG=http2client=0",
			"GOCACHE=/tmp/go-cache", // Enable Go module cache
			"GOPROXY=direct",        // Faster dependency resolution
		},
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
			"/tmp/go-cache:/tmp/go-cache", // Persist Go cache
		},
	}
}

// GetLanguageConfig returns language-specific configuration
func GetLanguageConfig(lang types.Language) LanguageConfig {
	switch lang {
	case types.LanguageGo:
		return LanguageConfig{
			Language:    types.LanguageGo,
			ImageName:   "golang:1.21-alpine",
			SetupScript: scripts.GetGoSetupScript(),
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
		}
	case types.LanguagePy:
		return LanguageConfig{
			Language:    types.LanguagePy,
			ImageName:   "python:3.12-alpine",
			SetupScript: scripts.GetPythonSetupScript(),
			RunCommand:  "python code.py",
			Extensions:  []string{".py"},
		}
	case types.LanguageJS:
		return LanguageConfig{
			Language:    types.LanguageJS,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetJavaScriptSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".js"},
		}
	case types.LanguageTS:
		return LanguageConfig{
			Language:    types.LanguageTS,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetTypeScriptSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".ts"},
		}
	case types.LanguageNode:
		return LanguageConfig{
			Language:    types.LanguageNode,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetNodeSetupScript(),
			RunCommand:  "node code.js",
			Extensions:  []string{".js", ".mjs", ".cjs"},
		}
	default:
		return LanguageConfig{
			Language:    types.LanguageGo,
			ImageName:   "golang:1.21-alpine",
			SetupScript: scripts.GetGoSetupScript(),
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
		}
	}
}

// GetLanguagePoolConfig returns pool configuration for a specific language
func GetLanguagePoolConfig(lang types.Language) LanguagePoolConfig {
	baseConfig := DefaultBasePoolConfig()
	return LanguagePoolConfig{
		Language:       lang,
		BasePoolConfig: baseConfig,
		Config:         GetLanguageConfig(lang),
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
		MaxContainers:       5,
		MinContainers:       2,
		IdleTimeout:         30 * time.Minute,
		PreWarmCount:        3,
		MaxWaitTime:         60 * time.Second,
		CleanupInterval:     30 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
	}
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		CacheDir:          "/var/lib/triggerx/cache", // Persistent cache directory
		MaxCacheSize:      100 * 1024 * 1024,         // 100MB
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
	cfg.Cache.CacheDir = "/var/lib/triggerx/cache" // Persistent cache directory
	cfg.Cache.MaxCacheSize = 500 * 1024 * 1024     // 500MB

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
