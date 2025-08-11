package config

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

func TestDefaultConfig_ValidLanguage_ReturnsCompleteConfig(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected string
	}{
		{"go language", "go", "golang:1.21-alpine"},
		{"python language", "py", "python:3.12-alpine"},
		{"javascript language", "js", "node:22-alpine"},
		{"typescript language", "ts", "node:22-alpine"},
		{"node language", "node", "node:22-alpine"},
		{"unknown language", "unknown", "golang:1.21-alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig(tt.lang)

			// Test Docker config
			assert.Equal(t, tt.expected, config.Docker.Image)
			assert.Equal(t, 300, config.Docker.TimeoutSeconds)
			assert.True(t, config.Docker.AutoCleanup)
			assert.Equal(t, "1024m", config.Docker.MemoryLimit)
			assert.Equal(t, 1.0, config.Docker.CPULimit)
			assert.Equal(t, "bridge", config.Docker.NetworkMode)
			assert.Contains(t, config.Docker.SecurityOpt, "no-new-privileges")
			assert.False(t, config.Docker.ReadOnlyRootFS)

			// Test environment variables
			assert.Contains(t, config.Docker.Environment, "GODEBUG=http2client=0")
			assert.Contains(t, config.Docker.Environment, "GOCACHE=/tmp/go-cache")
			assert.Contains(t, config.Docker.Environment, "GOPROXY=direct")

			// Test binds
			assert.Contains(t, config.Docker.Binds, "/var/run/docker.sock:/var/run/docker.sock")
			assert.Contains(t, config.Docker.Binds, "/tmp/go-cache:/tmp/go-cache")

			// Test other configs are not empty
			assert.NotZero(t, config.Fees.PricePerTG)
			assert.NotZero(t, config.BasePool.MaxContainers)
			assert.NotZero(t, config.Cache.MaxCacheSize)
			assert.True(t, config.Validation.EnableCodeValidation)
		})
	}
}

func TestDefaultDockerConfig_ValidLanguage_ReturnsCorrectImage(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected string
	}{
		{"go language", "go", "golang:1.21-alpine"},
		{"python language", "py", "python:3.12-alpine"},
		{"javascript language", "js", "node:22-alpine"},
		{"typescript language", "ts", "node:22-alpine"},
		{"node language", "node", "node:22-alpine"},
		{"unknown language", "unknown", "golang:1.21-alpine"},
		{"empty language", "", "golang:1.21-alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultDockerConfig(tt.lang)

			assert.Equal(t, tt.expected, config.Image)
			assert.Equal(t, 300, config.TimeoutSeconds)
			assert.True(t, config.AutoCleanup)
			assert.Equal(t, "1024m", config.MemoryLimit)
			assert.Equal(t, 1.0, config.CPULimit)
			assert.Equal(t, "bridge", config.NetworkMode)
			assert.False(t, config.ReadOnlyRootFS)
		})
	}
}

func TestDefaultDockerConfig_EnvironmentVariables_AreCorrectlySet(t *testing.T) {
	config := DefaultDockerConfig("go")

	expectedEnvVars := []string{
		"GODEBUG=http2client=0",
		"GOCACHE=/tmp/go-cache",
		"GOPROXY=direct",
	}

	for _, expected := range expectedEnvVars {
		assert.Contains(t, config.Environment, expected)
	}
}

func TestDefaultDockerConfig_Binds_AreCorrectlySet(t *testing.T) {
	config := DefaultDockerConfig("go")

	expectedBinds := []string{
		"/var/run/docker.sock:/var/run/docker.sock",
		"/tmp/go-cache:/tmp/go-cache",
	}

	for _, expected := range expectedBinds {
		assert.Contains(t, config.Binds, expected)
	}
}

func TestGetLanguageConfig_ValidLanguages_ReturnsCorrectConfig(t *testing.T) {
	tests := []struct {
		name     string
		lang     types.Language
		expected LanguageConfig
	}{
		{
			name: "go language",
			lang: types.LanguageGo,
			expected: LanguageConfig{
				Language:   types.LanguageGo,
				ImageName:  "golang:1.21-alpine",
				RunCommand: "go run code.go",
				Extensions: []string{".go"},
			},
		},
		{
			name: "python language",
			lang: types.LanguagePy,
			expected: LanguageConfig{
				Language:   types.LanguagePy,
				ImageName:  "python:3.12-alpine",
				RunCommand: "python code.py",
				Extensions: []string{".py"},
			},
		},
		{
			name: "javascript language",
			lang: types.LanguageJS,
			expected: LanguageConfig{
				Language:   types.LanguageJS,
				ImageName:  "node:22-alpine",
				RunCommand: "node code.js",
				Extensions: []string{".js"},
			},
		},
		{
			name: "typescript language",
			lang: types.LanguageTS,
			expected: LanguageConfig{
				Language:   types.LanguageTS,
				ImageName:  "node:22-alpine",
				RunCommand: "node code.js",
				Extensions: []string{".ts"},
			},
		},
		{
			name: "node language",
			lang: types.LanguageNode,
			expected: LanguageConfig{
				Language:   types.LanguageNode,
				ImageName:  "node:22-alpine",
				RunCommand: "node code.js",
				Extensions: []string{".js", ".mjs", ".cjs"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetLanguageConfig(tt.lang)

			assert.Equal(t, tt.expected.Language, config.Language)
			assert.Equal(t, tt.expected.ImageName, config.ImageName)
			assert.Equal(t, tt.expected.RunCommand, config.RunCommand)
			assert.Equal(t, tt.expected.Extensions, config.Extensions)
			assert.NotEmpty(t, config.SetupScript)
		})
	}
}

func TestGetLanguageConfig_InvalidLanguage_ReturnsGoDefault(t *testing.T) {
	config := GetLanguageConfig("invalid")

	expected := LanguageConfig{
		Language:   types.LanguageGo,
		ImageName:  "golang:1.21-alpine",
		RunCommand: "go run code.go",
		Extensions: []string{".go"},
	}

	assert.Equal(t, expected.Language, config.Language)
	assert.Equal(t, expected.ImageName, config.ImageName)
	assert.Equal(t, expected.RunCommand, config.RunCommand)
	assert.Equal(t, expected.Extensions, config.Extensions)
	assert.NotEmpty(t, config.SetupScript)
}

func TestGetLanguagePoolConfig_ValidLanguage_ReturnsCorrectConfig(t *testing.T) {
	lang := types.LanguageGo
	config := GetLanguagePoolConfig(lang)

	assert.Equal(t, lang, config.Language)
	assert.Equal(t, lang, config.Config.Language)
	assert.Equal(t, "golang:1.21-alpine", config.Config.ImageName)
	assert.Equal(t, "go run code.go", config.Config.RunCommand)
	assert.Equal(t, []string{".go"}, config.Config.Extensions)

	// Test that base pool config is set
	assert.Equal(t, 5, config.MaxContainers)
	assert.Equal(t, 2, config.MinContainers)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Equal(t, 3, config.PreWarmCount)
	assert.Equal(t, 60*time.Second, config.MaxWaitTime)
	assert.Equal(t, 30*time.Minute, config.CleanupInterval)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
}

func TestGetSupportedLanguages_ReturnsAllLanguages(t *testing.T) {
	languages := GetSupportedLanguages()

	expected := []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}

	assert.Equal(t, expected, languages)
	assert.Len(t, languages, 5)
}

func TestDefaultFeeConfig_ReturnsCorrectValues(t *testing.T) {
	config := DefaultFeeConfig()

	assert.Equal(t, 0.0001, config.PricePerTG)
	assert.Equal(t, 1.0, config.FixedCost)
	assert.Equal(t, 1.0, config.TransactionSimulation)
	assert.Equal(t, 0.1, config.OverheadCost)
}

func TestDefaultBasePoolConfig_ReturnsCorrectValues(t *testing.T) {
	config := DefaultBasePoolConfig()

	assert.Equal(t, 5, config.MaxContainers)
	assert.Equal(t, 2, config.MinContainers)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Equal(t, 3, config.PreWarmCount)
	assert.Equal(t, 60*time.Second, config.MaxWaitTime)
	assert.Equal(t, 30*time.Minute, config.CleanupInterval)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
}

func TestDefaultCacheConfig_ReturnsCorrectValues(t *testing.T) {
	config := DefaultCacheConfig()

	assert.Equal(t, "data/cache", config.CacheDir)
	assert.Equal(t, int64(100*1024*1024), config.MaxCacheSize) // 100MB
	assert.Equal(t, 10*time.Minute, config.CleanupInterval)
	assert.True(t, config.EnableCompression)
	assert.Equal(t, int64(1*1024*1024), config.MaxFileSize) // 1MB
}

func TestDefaultValidationConfig_ReturnsCorrectValues(t *testing.T) {
	config := DefaultValidationConfig()

	assert.True(t, config.EnableCodeValidation)
	assert.Equal(t, int64(1*1024*1024), config.MaxFileSize) // 1MB
	assert.Equal(t, []string{".go", ".py", ".js", ".ts"}, config.AllowedExtensions)
	assert.Contains(t, config.BlockedPatterns, "os.RemoveAll")
	assert.Contains(t, config.BlockedPatterns, "exec.Command")
	assert.Contains(t, config.BlockedPatterns, "syscall")
	assert.Contains(t, config.BlockedPatterns, "runtime.GC")
	assert.Contains(t, config.BlockedPatterns, "panic(")
	assert.Equal(t, 30, config.TimeoutSeconds)
}

func TestOptimizedConfig_ReturnsOptimizedValues(t *testing.T) {
	config := OptimizedConfig("go")

	// Test optimized pool settings
	assert.Equal(t, 10, config.BasePool.MaxContainers)
	assert.Equal(t, 5, config.BasePool.MinContainers)
	assert.Equal(t, 8, config.BasePool.PreWarmCount)
	assert.Equal(t, 5*time.Minute, config.BasePool.IdleTimeout)

	// Test optimized cache settings
	assert.Equal(t, "data/cache", config.Cache.CacheDir)
	assert.Equal(t, int64(500*1024*1024), config.Cache.MaxCacheSize) // 500MB

	// Test optimized Docker settings
	assert.Equal(t, "2048m", config.Docker.MemoryLimit)
	assert.Equal(t, 2.0, config.Docker.CPULimit)

	// Test that other settings remain from default
	assert.Equal(t, "golang:1.21-alpine", config.Docker.Image)
	assert.Equal(t, 300, config.Docker.TimeoutSeconds)
	assert.True(t, config.Docker.AutoCleanup)
}

func TestDevelopmentConfig_ReturnsDevelopmentValues(t *testing.T) {
	config := DevelopmentConfig("go")

	// Test development pool settings
	assert.Equal(t, 2, config.BasePool.MaxContainers)
	assert.Equal(t, 1, config.BasePool.MinContainers)
	assert.Equal(t, 1, config.BasePool.PreWarmCount)
	assert.Equal(t, 2*time.Minute, config.BasePool.IdleTimeout)

	// Test that other settings remain from default
	assert.Equal(t, "golang:1.21-alpine", config.Docker.Image)
	assert.Equal(t, 300, config.Docker.TimeoutSeconds)
	assert.True(t, config.Docker.AutoCleanup)
	assert.Equal(t, "1024m", config.Docker.MemoryLimit)
	assert.Equal(t, 1.0, config.Docker.CPULimit)
}

func TestDockerConfig_MemoryLimitBytes_ValidMemoryLimit_ReturnsCorrectBytes(t *testing.T) {
	tests := []struct {
		name          string
		memoryLimit   string
		expectedBytes uint64
		expectError   bool
	}{
		{"1024m", "1024m", 1024 * 1024 * 1024, false},
		{"512m", "512m", 512 * 1024 * 1024, false},
		{"2g", "2g", 2 * 1024 * 1024 * 1024, false},
		{"1g", "1g", 1024 * 1024 * 1024, false},
		{"100mb", "100mb", 100 * 1024 * 1024, false},
		{"invalid", "invalid", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerConfig := DockerConfig{MemoryLimit: tt.memoryLimit}
			result := dockerConfig.MemoryLimitBytes()

			if tt.expectError {
				assert.Equal(t, uint64(0), result)
			} else {
				assert.Equal(t, tt.expectedBytes, result)
			}
		})
	}
}

func TestDockerConfig_ToContainerResources_ReturnsCorrectResources(t *testing.T) {
	tests := []struct {
		name           string
		memoryLimit    string
		cpuLimit       float64
		expectedMemory int64
		expectedCPU    int64
	}{
		{
			name:           "1024m memory, 1.0 CPU",
			memoryLimit:    "1024m",
			cpuLimit:       1.0,
			expectedMemory: 1024 * 1024 * 1024,
			expectedCPU:    1e9,
		},
		{
			name:           "2048m memory, 2.0 CPU",
			memoryLimit:    "2048m",
			cpuLimit:       2.0,
			expectedMemory: 2048 * 1024 * 1024,
			expectedCPU:    2e9,
		},
		{
			name:           "512m memory, 0.5 CPU",
			memoryLimit:    "512m",
			cpuLimit:       0.5,
			expectedMemory: 512 * 1024 * 1024,
			expectedCPU:    5e8,
		},
		{
			name:           "invalid memory, 1.0 CPU",
			memoryLimit:    "invalid",
			cpuLimit:       1.0,
			expectedMemory: 0,
			expectedCPU:    1e9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerConfig := DockerConfig{
				MemoryLimit: tt.memoryLimit,
				CPULimit:    tt.cpuLimit,
			}

			resources := dockerConfig.ToContainerResources()

			assert.Equal(t, tt.expectedMemory, resources.Memory)
			assert.Equal(t, tt.expectedCPU, resources.NanoCPUs)
		})
	}
}

func TestDockerConfig_ToContainerResources_MatchesDockerAPI(t *testing.T) {
	dockerConfig := DockerConfig{
		MemoryLimit: "1024m",
		CPULimit:    1.5,
	}

	resources := dockerConfig.ToContainerResources()

	// Verify the returned type matches Docker API
	assert.IsType(t, container.Resources{}, resources)
	assert.Equal(t, int64(1024*1024*1024), resources.Memory)
	assert.Equal(t, int64(1.5e9), resources.NanoCPUs)
}

// Benchmark tests for performance-critical functions
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultConfig("go")
	}
}

func BenchmarkGetLanguageConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetLanguageConfig(types.LanguageGo)
	}
}

func BenchmarkDockerConfig_MemoryLimitBytes(b *testing.B) {
	dockerConfig := DockerConfig{MemoryLimit: "1024m"}
	for i := 0; i < b.N; i++ {
		dockerConfig.MemoryLimitBytes()
	}
}

func BenchmarkDockerConfig_ToContainerResources(b *testing.B) {
	dockerConfig := DockerConfig{
		MemoryLimit: "1024m",
		CPULimit:    1.0,
	}
	for i := 0; i < b.N; i++ {
		dockerConfig.ToContainerResources()
	}
}
