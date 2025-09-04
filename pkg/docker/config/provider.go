package config

import (
	"fmt"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"gopkg.in/yaml.v3"
)

// ConfigProvider holds the final, validated configuration and implements the Provider interface.
type ConfigProvider struct {
	cfg CodeExecutorConfig
}

// NewConfigProvider creates a new configuration provider.
// If the file does not exist, it is ignored and defaults are used without error.
// An error is returned for any other file reading issue, YAML parsing error, or validation failure.
func NewConfigProvider(yamlFilePath string) (*ConfigProvider, error) {
	var cfg CodeExecutorConfig
	if yamlFilePath != "" {
		yamlFile, err := os.ReadFile(yamlFilePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file at %s: %w", yamlFilePath, err)
			}
		} else {
			if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &ConfigProvider{cfg: cfg}, nil
}

// GetConfig returns the complete configuration
func (cp *ConfigProvider) GetConfig() CodeExecutorConfig {
	return cp.cfg
}

// GetLanguagePoolConfig returns the configuration for a specific language
func (cp *ConfigProvider) GetLanguagePoolConfig(language types.Language) (LanguagePoolConfig, bool) {
	langStr := string(language)
	config, exists := cp.cfg.Languages[langStr]
	return config, exists
}

// GetFeesConfig returns the execution fees configuration
func (cp *ConfigProvider) GetFeesConfig() ExecutionFeeConfig {
	return cp.cfg.Fees
}

// GetCacheConfig returns the file cache configuration
func (cp *ConfigProvider) GetCacheConfig() FileCacheConfig {
	return cp.cfg.Cache
}

// GetValidationConfig returns the validation configuration
func (cp *ConfigProvider) GetValidationConfig() ValidationConfig {
	return cp.cfg.Validation
}

// GetMonitoringConfig returns the monitoring configuration
func (cp *ConfigProvider) GetMonitoringConfig() MonitoringConfig {
	return cp.cfg.Monitoring
}

// GetManagerConfig returns the manager configuration
func (cp *ConfigProvider) GetManagerConfig() ManagerConfig {
	return cp.cfg.Manager
}

// GetSupportedLanguages returns all supported languages from the configuration
func (cp *ConfigProvider) GetSupportedLanguages() []types.Language {
	languages := make([]types.Language, 0, len(cp.cfg.Languages))
	for langStr := range cp.cfg.Languages {
		languages = append(languages, types.Language(langStr))
	}
	return languages
}
