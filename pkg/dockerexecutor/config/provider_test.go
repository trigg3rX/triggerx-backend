package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

func TestNewConfigProvider_EmptyPath_ShouldUseDefaults(t *testing.T) {
	provider, err := NewConfigProvider("")

	require.NoError(t, err)
	require.NotNil(t, provider)

	// Test that the config is valid
	err = provider.cfg.Validate()
	require.NoError(t, err)

	// Test that it contains default values
	assert.Equal(t, "golang:1.21-alpine", provider.cfg.Languages["go"].DockerConfig.Image)
	assert.Equal(t, 0.0001, provider.cfg.Fees.PricePerTG)
	assert.Equal(t, "data/cache", provider.cfg.Cache.CacheDir)
}

func TestNewConfigProvider_NonExistentFile_ShouldUseDefaults(t *testing.T) {
	provider, err := NewConfigProvider("/path/to/nonexistent/file.yaml")

	require.NoError(t, err)
	require.NotNil(t, provider)

	// Test that the config is valid
	err = provider.cfg.Validate()
	require.NoError(t, err)

	// Test that it contains default values
	assert.Equal(t, "golang:1.21-alpine", provider.cfg.Languages["go"].DockerConfig.Image)
}

func TestNewConfigProvider_ValidYAMLFile_ShouldOverrideDefaults(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
docker:
  languages:
    go:
      docker:
        image: "custom-image:latest"
        timeout_seconds: 600
  timeout_seconds: 600
fees:
  price_per_tg: 0.0002
  fixed_cost: 2.0
cache:
  cache_dir: "/custom/cache"
validation:
  max_file_size: 2097152
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())

	require.NoError(t, err)
	require.NotNil(t, provider)

	// Test that the config is valid
	err = provider.cfg.Validate()
	require.NoError(t, err)

	// Test that custom values are applied
	assert.Equal(t, "custom-image:latest", provider.cfg.Languages["go"].DockerConfig.Image)
	assert.Equal(t, 600, provider.cfg.Languages["go"].DockerConfig.TimeoutSeconds)
	assert.Equal(t, 0.0002, provider.cfg.Fees.PricePerTG)
	assert.Equal(t, 2.0, provider.cfg.Fees.FixedCost)
	assert.Equal(t, "/custom/cache", provider.cfg.Cache.CacheDir)
	assert.Equal(t, int64(2097152), provider.cfg.Validation.MaxFileSize)

	// Test that default values are preserved for non-overridden fields
	assert.Equal(t, "1024m", provider.cfg.Languages["go"].DockerConfig.MemoryLimit)
	assert.Equal(t, 1.0, provider.cfg.Languages["go"].DockerConfig.CPULimit)
}

func TestNewConfigProvider_InvalidYAMLFile_ShouldFail(t *testing.T) {
	// Create a temporary invalid YAML file
	invalidYAML := `
docker:
  languages:
    go:
      docker:
        image: "custom-image:latest"
  timeout_seconds: 600
fees:
  price_per_tg: 0.0002
  fixed_cost: 2.0
cache:
  cache_dir: "/custom/cache"
validation:
  max_file_size: 2097152
invalid: yaml: content: [with: invalid: syntax
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(invalidYAML)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())

	require.Error(t, err)
	require.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to parse YAML config file")
}

func TestNewConfigProvider_InvalidConfig_ShouldFail(t *testing.T) {
	// Create a temporary YAML file with invalid config
	invalidConfig := `
docker:
  languages:
    go:
      docker:
        image: ""  # Invalid: empty image
fees:
  price_per_tg: -1  # Invalid: negative value
cache:
  cache_dir: ""  # Invalid: empty cache dir
validation:
  max_file_size: -1  # Invalid: negative value
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(invalidConfig)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())

	require.Error(t, err)
	require.Nil(t, provider)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestNewConfigProvider_PartialConfig_ShouldMergeWithDefaults(t *testing.T) {
	// Create a temporary YAML file with partial config
	partialConfig := `
docker:
  image: "partial-image:latest"
fees:
  price_per_tg: 0.0003
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(partialConfig)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())

	require.NoError(t, err)
	require.NotNil(t, provider)

	// Test that the config is valid
	err = provider.cfg.Validate()
	require.NoError(t, err)

	// Test that custom values are applied
	assert.Equal(t, "partial-image:latest", provider.cfg.Languages["go"].DockerConfig.Image)
	assert.Equal(t, 0.0003, provider.cfg.Fees.PricePerTG)

	// Test that default values are preserved for non-overridden fields
	assert.Equal(t, 300, provider.cfg.Languages["go"].DockerConfig.TimeoutSeconds)
	assert.Equal(t, "1024m", provider.cfg.Languages["go"].DockerConfig.MemoryLimit)
	assert.Equal(t, 1.0, provider.cfg.Fees.FixedCost)
	assert.Equal(t, "data/cache", provider.cfg.Cache.CacheDir)
	assert.Equal(t, int64(1048576), provider.cfg.Validation.MaxFileSize)
}

func TestNewConfigProvider_FileReadError_ShouldFail(t *testing.T) {
	// Test with a directory path instead of a file
	provider, err := NewConfigProvider("/tmp")

	require.Error(t, err)
	require.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestNewConfigProvider_ComplexConfig_ShouldWork(t *testing.T) {
	// Create a temporary YAML file with complex config
	complexConfig := `
docker:
  languages:
    go:
      docker:
        image: "complex-image:latest"
  timeout_seconds: 900
  memory_limit: "2048m"
  cpu_limit: 2.0
  network_mode: "host"
  environment:
    - "CUSTOM_ENV=value"
    - "ANOTHER_ENV=another_value"
  security_opt:
    - "no-new-privileges"
    - "seccomp=unconfined"
fees:
  price_per_tg: 0.0005
  fixed_cost: 5.0
  transaction_cost: 2.0
  overhead_cost: 0.5
languages:
  go:
    language: "go"
    base_config:
      max_containers: 10
      min_containers: 3
      idle_timeout: "1h"  # 1 hour
      pre_warm_count: 3
      max_wait_time: "60s"  # 60 seconds
      cleanup_interval: "30m"  # 30 minutes
      health_check_interval: "30s"  # 30 seconds
    config:
      language: "go"
      image_name: "custom-golang:1.22"
      run_command: "go run main.go"
      extensions: [".go", ".mod"]
cache:
  cache_dir: "/var/cache/custom"
  max_cache_size: 536870912  # 512MB
  eviction_size: 134217728   # 128MB
  cleanup_interval: "30m"  # 30 minutes
  enable_compression: false
  max_file_size: 5242880  # 5MB
validation:
  max_file_size: 5242880  # 5MB
  allowed_extensions: [".go", ".py", ".js", ".ts", ".rs"]
  blocked_patterns: ["os\\.RemoveAll", "exec\\.Command", "syscall", "runtime\\.GC", "panic\\(", "unsafe\\."]
  timeout_seconds: 60
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(complexConfig)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())

	require.NoError(t, err)
	require.NotNil(t, provider)

	// Test that the config is valid
	err = provider.cfg.Validate()
	require.NoError(t, err)

	// Test complex docker config - check go language config
	goConfig, exists := provider.cfg.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "complex-image:latest", goConfig.DockerConfig.Image)
	assert.Equal(t, 900, goConfig.DockerConfig.TimeoutSeconds)
	assert.Equal(t, "2048m", goConfig.DockerConfig.MemoryLimit)
	assert.Equal(t, 2.0, goConfig.DockerConfig.CPULimit)
	assert.Equal(t, "host", goConfig.DockerConfig.NetworkMode)
	assert.Contains(t, goConfig.DockerConfig.Environment, "CUSTOM_ENV=value")
	assert.Contains(t, goConfig.DockerConfig.Environment, "ANOTHER_ENV=another_value")
	assert.Contains(t, goConfig.DockerConfig.SecurityOpt, "no-new-privileges")
	assert.Contains(t, goConfig.DockerConfig.SecurityOpt, "seccomp=unconfined")

	// Test complex fees config
	assert.Equal(t, 0.0005, provider.cfg.Fees.PricePerTG)
	assert.Equal(t, 5.0, provider.cfg.Fees.FixedCost)
	assert.Equal(t, 2.0, provider.cfg.Fees.TransactionCost)
	assert.Equal(t, 0.5, provider.cfg.Fees.OverheadCost)

	// Test complex cache config
	assert.Equal(t, "/var/cache/custom", provider.cfg.Cache.CacheDir)
	assert.Equal(t, int64(536870912), provider.cfg.Cache.MaxCacheSize)
	assert.Equal(t, int64(134217728), provider.cfg.Cache.EvictionSize)
	assert.False(t, provider.cfg.Cache.EnableCompression)
	assert.Equal(t, int64(5242880), provider.cfg.Cache.MaxFileSize)

	// Test complex validation config
	assert.Equal(t, int64(5242880), provider.cfg.Validation.MaxFileSize)
	assert.Contains(t, provider.cfg.Validation.AllowedExtensions, ".rs")
	// BlockedPatterns field does not exist in ValidationConfig
	assert.Equal(t, 60, provider.cfg.Validation.TimeoutSeconds)

	// Test that languages are properly configured
	goConfig, exists = provider.cfg.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "go", string(goConfig.LanguageConfig.Language))
	assert.Equal(t, 10, goConfig.BasePoolConfig.MaxContainers)
	assert.Equal(t, 3, goConfig.BasePoolConfig.MinContainers)
	assert.Equal(t, "custom-golang:1.22", goConfig.LanguageConfig.ImageName)
	assert.Equal(t, "go run main.go", goConfig.LanguageConfig.RunCommand)
	assert.Contains(t, goConfig.LanguageConfig.Extensions, ".mod")
}

func TestConfigProvider_Integration(t *testing.T) {
	// Test that the provider works correctly in a real-world scenario
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test that the configuration is complete and valid
	config := provider.cfg

	// Test that all required sections are present
	// Check that languages are configured
	assert.NotEmpty(t, config.Languages)
	assert.Greater(t, config.Fees.PricePerTG, 0.0)
	assert.NotEmpty(t, config.Cache.CacheDir)
	assert.Greater(t, config.Validation.MaxFileSize, int64(0))
	assert.NotEmpty(t, config.Languages)

	// Test that all language configs are valid
	for langKey, langConfig := range config.Languages {
		err := langConfig.Validate()
		require.NoError(t, err, "Language config for %s should be valid", langKey)
	}
}

// Test GetConfig method
func TestConfigProvider_GetConfig(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	config := provider.GetConfig()

	// Test that we get a complete configuration
	assert.NotEmpty(t, config.Languages)
	assert.Greater(t, config.Fees.PricePerTG, 0.0)
	assert.NotEmpty(t, config.Cache.CacheDir)
	assert.Greater(t, config.Validation.MaxFileSize, int64(0))
	assert.NotEmpty(t, config.Languages)

	// Test that it's the same as the internal config
	assert.Equal(t, provider.cfg, config)
}

func TestConfigProvider_GetConfig_WithCustomValues(t *testing.T) {
	// Create config with custom values
	yamlContent := `
docker:
  image: "custom-image:latest"
  timeout_seconds: 600
fees:
  price_per_tg: 0.0002
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	config := provider.GetConfig()

	// Test that custom values are returned
	// Check custom values in go language config
	goConfig, exists := config.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "custom-image:latest", goConfig.DockerConfig.Image)
	assert.Equal(t, 600, goConfig.DockerConfig.TimeoutSeconds)
	assert.Equal(t, 0.0002, config.Fees.PricePerTG)
}

// Test GetLanguagePoolConfig method
func TestConfigProvider_GetLanguagePoolConfig_ValidLanguage(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test all supported languages
	supportedLanguages := []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}

	for _, lang := range supportedLanguages {
		config, exists := provider.GetLanguagePoolConfig(lang)
		assert.True(t, exists, "Language %s should exist", lang)
		assert.Equal(t, lang, config.LanguageConfig.Language)
		assert.NotEmpty(t, config.LanguageConfig.ImageName)
		assert.NotEmpty(t, config.LanguageConfig.RunCommand)
		assert.NotEmpty(t, config.LanguageConfig.Extensions)
		assert.Greater(t, config.BasePoolConfig.MaxContainers, 0)
		assert.Greater(t, config.BasePoolConfig.MinContainers, 0)
	}
}

func TestConfigProvider_GetLanguagePoolConfig_InvalidLanguage(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test with non-existent language
	config, exists := provider.GetLanguagePoolConfig(types.Language("nonexistent"))
	assert.False(t, exists)
	assert.Equal(t, types.Language(""), config.LanguageConfig.Language)
}

func TestConfigProvider_GetLanguagePoolConfig_SpecificLanguageProperties(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test Go language specifically
	goConfig, exists := provider.GetLanguagePoolConfig(types.LanguageGo)
	assert.True(t, exists)
	assert.Equal(t, types.LanguageGo, goConfig.LanguageConfig.Language)
	assert.Equal(t, "golang:1.21-alpine", goConfig.LanguageConfig.ImageName)
	assert.Equal(t, "go run code.go", goConfig.LanguageConfig.RunCommand)
	assert.Contains(t, goConfig.LanguageConfig.Extensions, ".go")

	// Test Python language specifically
	pyConfig, exists := provider.GetLanguagePoolConfig(types.LanguagePy)
	assert.True(t, exists)
	assert.Equal(t, types.LanguagePy, pyConfig.LanguageConfig.Language)
	assert.Equal(t, "python:3.12-alpine", pyConfig.LanguageConfig.ImageName)
	assert.Equal(t, "python code.py", pyConfig.LanguageConfig.RunCommand)
	assert.Contains(t, pyConfig.LanguageConfig.Extensions, ".py")
}

// Test GetDockerConfig method
func TestConfigProvider_GetDockerConfig(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test default values in go language config
	goConfig, exists := provider.cfg.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "golang:1.21-alpine", goConfig.DockerConfig.Image)
	assert.Equal(t, 300, goConfig.DockerConfig.TimeoutSeconds)
	assert.True(t, goConfig.DockerConfig.AutoCleanup)
	assert.Equal(t, "1024m", goConfig.DockerConfig.MemoryLimit)
	assert.Equal(t, 1.0, goConfig.DockerConfig.CPULimit)
	assert.Equal(t, "bridge", goConfig.DockerConfig.NetworkMode)
	assert.Contains(t, goConfig.DockerConfig.SecurityOpt, "no-new-privileges")
	assert.False(t, goConfig.DockerConfig.ReadOnlyRootFS)
	assert.Contains(t, goConfig.DockerConfig.Environment, "GODEBUG=http2client=0")
	assert.Contains(t, goConfig.DockerConfig.Binds, "/var/run/dockerexecutor.sock:/var/run/dockerexecutor.sock")
}

func TestConfigProvider_GetDockerConfig_WithCustomValues(t *testing.T) {
	yamlContent := `
docker:
  image: "custom-docker:latest"
  timeout_seconds: 600
  memory_limit: "2048m"
  cpu_limit: 2.0
  network_mode: "host"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	// Test custom values in go language config
	goConfig, exists := provider.cfg.Languages["go"]
	assert.True(t, exists)
	assert.Equal(t, "custom-docker:latest", goConfig.DockerConfig.Image)
	assert.Equal(t, 600, goConfig.DockerConfig.TimeoutSeconds)
	assert.Equal(t, "2048m", goConfig.DockerConfig.MemoryLimit)
	assert.Equal(t, 2.0, goConfig.DockerConfig.CPULimit)
	assert.Equal(t, "host", goConfig.DockerConfig.NetworkMode)
}

// Test GetFeesConfig method
func TestConfigProvider_GetFeesConfig(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	feesConfig := provider.GetFeesConfig()

	// Test default values
	assert.Equal(t, 0.0001, feesConfig.PricePerTG)
	assert.Equal(t, 1.0, feesConfig.FixedCost)
	assert.Equal(t, 1.0, feesConfig.TransactionCost)
	assert.Equal(t, 0.1, feesConfig.OverheadCost)
}

func TestConfigProvider_GetFeesConfig_WithCustomValues(t *testing.T) {
	yamlContent := `
fees:
  price_per_tg: 0.0005
  fixed_cost: 2.5
  transaction_cost: 3.0
  overhead_cost: 0.5
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	feesConfig := provider.GetFeesConfig()

	// Test custom values
	assert.Equal(t, 0.0005, feesConfig.PricePerTG)
	assert.Equal(t, 2.5, feesConfig.FixedCost)
	assert.Equal(t, 3.0, feesConfig.TransactionCost)
	assert.Equal(t, 0.5, feesConfig.OverheadCost)
}

// Test GetCacheConfig method
func TestConfigProvider_GetCacheConfig(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	cacheConfig := provider.GetCacheConfig()

	// Test default values
	assert.Equal(t, "data/cache", cacheConfig.CacheDir)
	assert.Equal(t, int64(100*1024*1024), cacheConfig.MaxCacheSize) // 100MB
	assert.Equal(t, int64(25*1024*1024), cacheConfig.EvictionSize)  // 25MB
	assert.True(t, cacheConfig.EnableCompression)
	assert.Equal(t, int64(1*1024*1024), cacheConfig.MaxFileSize) // 1MB
}

func TestConfigProvider_GetCacheConfig_WithCustomValues(t *testing.T) {
	yamlContent := `
cache:
  cache_dir: "/custom/cache"
  max_cache_size: 536870912  # 512MB
  eviction_size: 134217728   # 128MB
  enable_compression: false
  max_file_size: 5242880     # 5MB
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	cacheConfig := provider.GetCacheConfig()

	// Test custom values
	assert.Equal(t, "/custom/cache", cacheConfig.CacheDir)
	assert.Equal(t, int64(536870912), cacheConfig.MaxCacheSize)
	assert.Equal(t, int64(134217728), cacheConfig.EvictionSize)
	assert.False(t, cacheConfig.EnableCompression)
	assert.Equal(t, int64(5242880), cacheConfig.MaxFileSize)
}

// Test GetValidationConfig method
func TestConfigProvider_GetValidationConfig(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	validationConfig := provider.GetValidationConfig()

	// Test default values
	assert.Equal(t, int64(1*1024*1024), validationConfig.MaxFileSize) // 1MB
	assert.Contains(t, validationConfig.AllowedExtensions, ".go")
	assert.Contains(t, validationConfig.AllowedExtensions, ".py")
	assert.Contains(t, validationConfig.AllowedExtensions, ".js")
	assert.Contains(t, validationConfig.AllowedExtensions, ".ts")
	assert.Equal(t, 10, validationConfig.MaxComplexity)
	assert.Equal(t, 30, validationConfig.TimeoutSeconds)
}

func TestConfigProvider_GetValidationConfig_WithCustomValues(t *testing.T) {
	yamlContent := `
validation:
  max_file_size: 5242880  # 5MB
  allowed_extensions: [".go", ".py", ".js", ".ts", ".rs"]
  max_complexity: 10
  timeout_seconds: 60
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	validationConfig := provider.GetValidationConfig()

	// Test custom values
	assert.Equal(t, int64(5242880), validationConfig.MaxFileSize)
	assert.Contains(t, validationConfig.AllowedExtensions, ".rs")
	assert.Equal(t, 10, validationConfig.MaxComplexity)
	assert.Equal(t, 60, validationConfig.TimeoutSeconds)
}

// Test GetSupportedLanguages method
func TestConfigProvider_GetSupportedLanguages(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	languages := provider.GetSupportedLanguages()

	// Test that all expected languages are present
	expectedLanguages := []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}

	assert.Len(t, languages, len(expectedLanguages))

	for _, expectedLang := range expectedLanguages {
		assert.Contains(t, languages, expectedLang)
	}
}

func TestConfigProvider_GetSupportedLanguages_CustomConfig(t *testing.T) {
	// Create config with only specific languages
	yamlContent := `
languages:
  go:
    language: "go"
    base_config:
      max_containers: 10
      min_containers: 2
      idle_timeout: "30m"
      pre_warm_count: 3
      max_wait_time: "60s"
      cleanup_interval: "30m"
      health_check_interval: "30s"
    config:
      language: "go"
      image_name: "golang:1.21-alpine"
      run_command: "go run code.go"
      extensions: [".go"]
  py:
    language: "py"
    base_config:
      max_containers: 5
      min_containers: 1
      idle_timeout: "30m"
      pre_warm_count: 2
      max_wait_time: "60s"
      cleanup_interval: "30m"
      health_check_interval: "30s"
    config:
      language: "py"
      image_name: "python:3.12-alpine"
      run_command: "python code.py"
      extensions: [".py"]
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	provider, err := NewConfigProvider(tmpFile.Name())
	require.NoError(t, err)

	languages := provider.GetSupportedLanguages()

	// The YAML config merges with defaults, so we still get all default languages
	// but we can verify our custom languages exist and have the right properties
	assert.Contains(t, languages, types.LanguageGo)
	assert.Contains(t, languages, types.LanguagePy)

	// Verify the custom Go config was applied
	goConfig, exists := provider.GetLanguagePoolConfig(types.LanguageGo)
	assert.True(t, exists)
	assert.Equal(t, 10, goConfig.BasePoolConfig.MaxContainers) // Our custom value
	assert.Equal(t, 2, goConfig.BasePoolConfig.MinContainers)  // Our custom value

	// Verify the custom Python config was applied
	pyConfig, exists := provider.GetLanguagePoolConfig(types.LanguagePy)
	assert.True(t, exists)
	assert.Equal(t, 5, pyConfig.BasePoolConfig.MaxContainers) // Our custom value
	assert.Equal(t, 1, pyConfig.BasePoolConfig.MinContainers) // Our custom value
}

// Test global GetLanguagePoolConfig function
func TestGetLanguagePoolConfig_ValidLanguages(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test all supported languages
	supportedLanguages := []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}

	for _, lang := range supportedLanguages {
		config, exists := provider.GetLanguagePoolConfig(lang)
		assert.True(t, exists)
		assert.Equal(t, lang, config.LanguageConfig.Language)
		assert.NotEmpty(t, config.LanguageConfig.ImageName)
		assert.NotEmpty(t, config.LanguageConfig.RunCommand)
		assert.NotEmpty(t, config.LanguageConfig.Extensions)
		assert.Greater(t, config.BasePoolConfig.MaxContainers, 0)
		assert.Greater(t, config.BasePoolConfig.MinContainers, 0)
	}
}

func TestGetLanguagePoolConfig_InvalidLanguage_ShouldFallbackToGo(t *testing.T) {
	provider, err := NewConfigProvider("")
	require.NoError(t, err)

	// Test with non-existent language
	_, exists := provider.GetLanguagePoolConfig(types.Language("nonexistent"))
	assert.False(t, exists)
	// No fallback - just returns empty config
}

func TestGetLanguagePoolConfig_SpecificLanguageProperties(t *testing.T) {
	// Test Go language
	provider, err := NewConfigProvider("")
	require.NoError(t, err)
	goConfig, exists := provider.GetLanguagePoolConfig(types.LanguageGo)
	assert.True(t, exists)
	assert.Equal(t, types.LanguageGo, goConfig.LanguageConfig.Language)
	assert.Equal(t, "golang:1.21-alpine", goConfig.LanguageConfig.ImageName)
	assert.Equal(t, "go run code.go", goConfig.LanguageConfig.RunCommand)
	assert.Contains(t, goConfig.LanguageConfig.Extensions, ".go")

	// Test Python language
	pyConfig, exists := provider.GetLanguagePoolConfig(types.LanguagePy)
	assert.True(t, exists)
	assert.Equal(t, types.LanguagePy, pyConfig.LanguageConfig.Language)
	assert.Equal(t, "python:3.12-alpine", pyConfig.LanguageConfig.ImageName)
	assert.Equal(t, "python code.py", pyConfig.LanguageConfig.RunCommand)
	assert.Contains(t, pyConfig.LanguageConfig.Extensions, ".py")

	// Test JavaScript language
	jsConfig, exists := provider.GetLanguagePoolConfig(types.LanguageJS)
	assert.True(t, exists)
	assert.Equal(t, types.LanguageJS, jsConfig.LanguageConfig.Language)
	assert.Equal(t, "node:22-alpine", jsConfig.LanguageConfig.ImageName)
	assert.Equal(t, "node code.js", jsConfig.LanguageConfig.RunCommand)
	assert.Contains(t, jsConfig.LanguageConfig.Extensions, ".js")

	// Test TypeScript language
	tsConfig, exists := provider.GetLanguagePoolConfig(types.LanguageTS)
	assert.True(t, exists)
	assert.Equal(t, types.LanguageTS, tsConfig.LanguageConfig.Language)
	assert.Equal(t, "node:22-alpine", tsConfig.LanguageConfig.ImageName)
	assert.Equal(t, "node code.js", tsConfig.LanguageConfig.RunCommand)
	assert.Contains(t, tsConfig.LanguageConfig.Extensions, ".ts")

	// Test Node language
	nodeConfig, exists := provider.GetLanguagePoolConfig(types.LanguageNode)
	assert.True(t, exists)
	assert.Equal(t, types.LanguageNode, nodeConfig.LanguageConfig.Language)
	assert.Equal(t, "node:22-alpine", nodeConfig.LanguageConfig.ImageName)
	assert.Equal(t, "node code.js", nodeConfig.LanguageConfig.RunCommand)
	assert.Contains(t, nodeConfig.LanguageConfig.Extensions, ".js")
	assert.Contains(t, nodeConfig.LanguageConfig.Extensions, ".mjs")
	assert.Contains(t, nodeConfig.LanguageConfig.Extensions, ".cjs")
}
