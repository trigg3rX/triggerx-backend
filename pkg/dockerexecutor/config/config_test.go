package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

func TestDockerContainerConfig_JSONMarshaling(t *testing.T) {
	config := DockerContainerConfig{
		Image:          "golang:1.21-alpine",
		TimeoutSeconds: 300,
		AutoCleanup:    true,
		MemoryLimit:    "1024m",
		CPULimit:       1.0,
		NetworkMode:    "bridge",
		SecurityOpt:    []string{"no-new-privileges"},
		ReadOnlyRootFS: false,
		Environment:    []string{"KEY=value", "ANOTHER=value2"},
		Binds:          []string{"/host:/container"},
		Languages:      []types.Language{types.LanguageGo, types.LanguagePy},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled DockerContainerConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.Image, unmarshaled.Image)
	assert.Equal(t, config.TimeoutSeconds, unmarshaled.TimeoutSeconds)
	assert.Equal(t, config.AutoCleanup, unmarshaled.AutoCleanup)
	assert.Equal(t, config.MemoryLimit, unmarshaled.MemoryLimit)
	assert.Equal(t, config.CPULimit, unmarshaled.CPULimit)
	assert.Equal(t, config.NetworkMode, unmarshaled.NetworkMode)
	assert.Equal(t, config.SecurityOpt, unmarshaled.SecurityOpt)
	assert.Equal(t, config.ReadOnlyRootFS, unmarshaled.ReadOnlyRootFS)
	assert.Equal(t, config.Environment, unmarshaled.Environment)
	assert.Equal(t, config.Binds, unmarshaled.Binds)
	assert.Equal(t, config.Languages, unmarshaled.Languages)
}

func TestExecutionFeeConfig_JSONMarshaling(t *testing.T) {
	config := ExecutionFeeConfig{
		PricePerTG:      0.0001,
		FixedCost:       1.0,
		TransactionCost: 1.0,
		StaticComplexityFactor: 0.1,
		DynamicComplexityFactor: 0.1,
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled ExecutionFeeConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.PricePerTG, unmarshaled.PricePerTG)
	assert.Equal(t, config.FixedCost, unmarshaled.FixedCost)
	assert.Equal(t, config.TransactionCost, unmarshaled.TransactionCost)
	assert.Equal(t, config.StaticComplexityFactor, unmarshaled.StaticComplexityFactor)
	assert.Equal(t, config.DynamicComplexityFactor, unmarshaled.DynamicComplexityFactor)
}

func TestBasePoolConfig_JSONMarshaling(t *testing.T) {
	config := BasePoolConfig{
		MaxContainers:       5,
		MinContainers:       2,
		MaxWaitTime:         60 * time.Second,
		HealthCheckInterval: 30 * time.Second,
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled BasePoolConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.MaxContainers, unmarshaled.MaxContainers)
	assert.Equal(t, config.MinContainers, unmarshaled.MinContainers)
	assert.Equal(t, config.MaxWaitTime, unmarshaled.MaxWaitTime)
	assert.Equal(t, config.HealthCheckInterval, unmarshaled.HealthCheckInterval)
}

func TestLanguageConfig_JSONMarshaling(t *testing.T) {
	config := LanguageConfig{
		Language:    types.LanguageGo,
		ImageName:   "golang:1.21-alpine",
		SetupScript: "setup.sh",
		RunCommand:  "go run code.go",
		Extensions:  []string{".go", ".mod"},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LanguageConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.Language, unmarshaled.Language)
	assert.Equal(t, config.ImageName, unmarshaled.ImageName)
	assert.Equal(t, config.SetupScript, unmarshaled.SetupScript)
	assert.Equal(t, config.RunCommand, unmarshaled.RunCommand)
	assert.Equal(t, config.Extensions, unmarshaled.Extensions)
}

func TestLanguagePoolConfig_JSONMarshaling(t *testing.T) {
	config := LanguagePoolConfig{
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
			Language:    types.LanguageGo,
			ImageName:   "golang:1.21-alpine",
			SetupScript: "setup.sh",
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
		},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LanguagePoolConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.BasePoolConfig.MaxContainers, unmarshaled.BasePoolConfig.MaxContainers)
	assert.Equal(t, config.BasePoolConfig.MinContainers, unmarshaled.BasePoolConfig.MinContainers)
	assert.Equal(t, config.LanguageConfig.Language, unmarshaled.LanguageConfig.Language)
	assert.Equal(t, config.LanguageConfig.ImageName, unmarshaled.LanguageConfig.ImageName)
	assert.Equal(t, config.LanguageConfig.RunCommand, unmarshaled.LanguageConfig.RunCommand)
}

func TestFileCacheConfig_JSONMarshaling(t *testing.T) {
	config := FileCacheConfig{
		CacheDir:          "data/cache",
		MaxCacheSize:      100 * 1024 * 1024,
		EvictionSize:      25 * 1024 * 1024,
		EnableCompression: true,
		MaxFileSize:       1 * 1024 * 1024,
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled FileCacheConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.CacheDir, unmarshaled.CacheDir)
	assert.Equal(t, config.MaxCacheSize, unmarshaled.MaxCacheSize)
	assert.Equal(t, config.EvictionSize, unmarshaled.EvictionSize)
	assert.Equal(t, config.EnableCompression, unmarshaled.EnableCompression)
	assert.Equal(t, config.MaxFileSize, unmarshaled.MaxFileSize)
}

func TestValidationConfig_JSONMarshaling(t *testing.T) {
	config := ValidationConfig{
		MaxFileSize:       1 * 1024 * 1024,
		AllowedExtensions: []string{".go", ".py", ".js"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled ValidationConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.MaxFileSize, unmarshaled.MaxFileSize)
	assert.Equal(t, config.AllowedExtensions, unmarshaled.AllowedExtensions)
	assert.Equal(t, config.MaxComplexity, unmarshaled.MaxComplexity)
	assert.Equal(t, config.TimeoutSeconds, unmarshaled.TimeoutSeconds)
}

func TestCodeExecutorConfig_JSONMarshaling(t *testing.T) {
	config := CodeExecutorConfig{
		Fees: ExecutionFeeConfig{
			PricePerTG:      0.0001,
			FixedCost:       1.0,
			TransactionCost: 1.0,
			StaticComplexityFactor: 0.1,
			DynamicComplexityFactor: 0.1,
		},
		Languages: map[string]LanguagePoolConfig{
			"go": {
				BasePoolConfig: BasePoolConfig{
					MaxContainers: 5,
					MinContainers: 2,
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
		},
		Cache: FileCacheConfig{
			CacheDir:          "data/cache",
			MaxCacheSize:      100 * 1024 * 1024,
			EvictionSize:      25 * 1024 * 1024,
			EnableCompression: true,
			MaxFileSize:       1 * 1024 * 1024,
		},
		Validation: ValidationConfig{
			MaxFileSize:       1 * 1024 * 1024,
			AllowedExtensions: []string{".go", ".py"},
			MaxComplexity:     10,
			TimeoutSeconds:    30,
		},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled CodeExecutorConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, config.Fees.PricePerTG, unmarshaled.Fees.PricePerTG)
	assert.Equal(t, config.Cache.CacheDir, unmarshaled.Cache.CacheDir)
	assert.Equal(t, config.Validation.MaxFileSize, unmarshaled.Validation.MaxFileSize)
	assert.Len(t, unmarshaled.Languages, 1)

	// Verify language config is preserved
	goConfig, exists := unmarshaled.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "golang:1.21-alpine", goConfig.LanguageConfig.ImageName)
}

func TestCodeExecutorConfig_EmptyConfig(t *testing.T) {
	config := CodeExecutorConfig{}

	// Test that empty config can be marshaled/unmarshaled
	data, err := json.Marshal(config)
	require.NoError(t, err)

	var unmarshaled CodeExecutorConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify empty config is preserved
	assert.Equal(t, config.Fees, unmarshaled.Fees)
	assert.Equal(t, config.Cache, unmarshaled.Cache)
	assert.Equal(t, config.Validation, unmarshaled.Validation)
	assert.Nil(t, unmarshaled.Languages)
}

func TestConfigStructs_ZeroValues(t *testing.T) {
	// Test that zero values are handled correctly
	dockerConfig := DockerContainerConfig{}
	feeConfig := ExecutionFeeConfig{}
	poolConfig := BasePoolConfig{}
	langConfig := LanguageConfig{}
	cacheConfig := FileCacheConfig{}
	validationConfig := ValidationConfig{}

	// Test that zero values can be marshaled
	_, err := json.Marshal(dockerConfig)
	require.NoError(t, err)

	_, err = json.Marshal(feeConfig)
	require.NoError(t, err)

	_, err = json.Marshal(poolConfig)
	require.NoError(t, err)

	_, err = json.Marshal(langConfig)
	require.NoError(t, err)

	_, err = json.Marshal(cacheConfig)
	require.NoError(t, err)

	_, err = json.Marshal(validationConfig)
	require.NoError(t, err)
}

func TestConfigStructs_FieldAccess(t *testing.T) {
	// Test that all fields can be accessed and modified
	dockerConfig := DockerContainerConfig{}
	dockerConfig.Image = "test-image"
	dockerConfig.TimeoutSeconds = 100
	dockerConfig.Environment = []string{"TEST=value"}

	assert.Equal(t, "test-image", dockerConfig.Image)
	assert.Equal(t, 100, dockerConfig.TimeoutSeconds)
	assert.Contains(t, dockerConfig.Environment, "TEST=value")

	feeConfig := ExecutionFeeConfig{}
	feeConfig.PricePerTG = 0.5
	feeConfig.FixedCost = 10.0

	assert.Equal(t, 0.5, feeConfig.PricePerTG)
	assert.Equal(t, 10.0, feeConfig.FixedCost)

	poolConfig := BasePoolConfig{}
	poolConfig.MaxContainers = 20
	poolConfig.MinContainers = 5

	assert.Equal(t, 20, poolConfig.MaxContainers)
	assert.Equal(t, 5, poolConfig.MinContainers)
}

func TestLanguageTypes_Compatibility(t *testing.T) {
	// Test that language types work correctly with the config
	config := LanguageConfig{
		Language: types.LanguageGo,
	}

	// Test type conversion
	langStr := string(config.Language)
	assert.Equal(t, "go", langStr)

	// Test comparison
	assert.Equal(t, types.LanguageGo, config.Language)
	assert.NotEqual(t, types.LanguagePy, config.Language)

	// Test in slice
	languages := []types.Language{types.LanguageGo, types.LanguagePy}
	assert.Contains(t, languages, types.LanguageGo)
	assert.Contains(t, languages, types.LanguagePy)
	assert.NotContains(t, languages, types.LanguageJS)
}
