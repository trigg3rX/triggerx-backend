package config

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

func DefaultConfig(lang string) ExecutorConfig {
	return ExecutorConfig{
		Docker:     DefaultDockerConfig(lang),
		Fees:       DefaultFeeConfig(),
		Pool:       DefaultPoolConfig(),
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
	case "node":
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

func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		PricePerTG:            0.0001,
		FixedCost:             1.0,
		TransactionSimulation: 1.0,
		OverheadCost:          0.1,
	}
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxContainers:       5,
		MinContainers:       2,
		IdleTimeout:         50 * time.Minute,
		PreWarmCount:        3,
		MaxWaitTime:         100 * time.Second,
		CleanupInterval:     50 * time.Minute,
		HealthCheckInterval: 1 * time.Minute,
	}
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		CacheTTL:          1 * time.Hour,
		CleanupInterval:   10 * time.Minute,
		EnableCompression: true,
		MaxFileSize:       10 * 1024 * 1024, // 10MB
	}
}

func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		EnableCodeValidation: true,
		MaxFileSize:          10 * 1024 * 1024, // 10MB
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
	cfg.Pool.MaxContainers = 10
	cfg.Pool.MinContainers = 5
	cfg.Pool.PreWarmCount = 8
	cfg.Pool.IdleTimeout = 5 * time.Minute

	// Optimize cache settings
	cfg.Cache.MaxCacheSize = 500 * 1024 * 1024 // 500MB
	cfg.Cache.CacheTTL = 2 * time.Hour

	// Optimize Docker settings
	cfg.Docker.MemoryLimit = "2048m"
	cfg.Docker.CPULimit = 2.0

	return cfg
}

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig(lang string) ExecutorConfig {
	cfg := DefaultConfig(lang)

	// Reduce resource usage for development
	cfg.Pool.MaxContainers = 2
	cfg.Pool.MinContainers = 1
	cfg.Pool.PreWarmCount = 1

	// Shorter timeouts for faster feedback
	cfg.Pool.IdleTimeout = 2 * time.Minute
	cfg.Cache.CacheTTL = 30 * time.Minute

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
