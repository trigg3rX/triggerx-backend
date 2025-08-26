package config

import "github.com/trigg3rX/triggerx-backend/pkg/docker/types"

//go:generate mockgen -destination=mock_provider.go -package=config . ConfigProviderInterface

// ConfigProviderInterface defines the interface for configuration providers
// This enables dependency injection and makes the code more testable
type ConfigProviderInterface interface {
	GetConfig() CodeExecutorConfig
	GetLanguagePoolConfig(language types.Language) (LanguagePoolConfig, bool)
	GetFeesConfig() ExecutionFeeConfig
	GetCacheConfig() FileCacheConfig
	GetValidationConfig() ValidationConfig
	GetMonitoringConfig() MonitoringConfig
	GetManagerConfig() ManagerConfig
	GetSupportedLanguages() []types.Language
}
