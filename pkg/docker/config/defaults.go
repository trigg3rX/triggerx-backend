package config

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

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
			"os.Exit",
			"panic(",
		},
		TimeoutSeconds: 30,
	}
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
