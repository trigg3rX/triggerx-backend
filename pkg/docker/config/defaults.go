package config

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/scripts"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

// GetDefaultConfig returns the complete default configuration for the CodeExecutor.
// This is the main entry point for obtaining a default setup.
func GetDefaultConfig() CodeExecutorConfig {
	return CodeExecutorConfig{
		Manager:    getDefaultManagerConfig(),
		Fees:       getDefaultExecutionFeeConfig(),
		Languages:  getDefaultLanguagePoolConfigs(),
		Cache:      getDefaultFileCacheConfig(),
		Validation: getDefaultValidationConfig(),
		Monitoring: getDefaultMonitoringConfig(),
	}
}

// getDefaultManagerConfig provides the default manager settings.
func getDefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		AutoCleanup: true,
	}
}

// getDefaultDockerContainerConfig provides the default Docker settings.
func getDefaultDockerContainerConfig() DockerContainerConfig {
	return DockerContainerConfig{
		Image:          "golang:1.21-alpine", // A base image; language-specific images are in their configs.
		TimeoutSeconds: 300,
		AutoCleanup:    true,
		MemoryLimit:    "1024m",
		CPULimit:       1.0,
		NetworkMode:    "bridge",
		SecurityOpt:    []string{"no-new-privileges"},
		ReadOnlyRootFS: false,
		Environment: []string{
			"GODEBUG=http2client=0",
			"GOCACHE=/tmp/go-cache",
			"GOPROXY=direct",
		},
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
			"/tmp/go-cache:/tmp/go-cache",
		},
		Languages: getSupportedLanguages(),
	}
}

// getDefaultExecutionFeeConfig provides the default fee settings.
func getDefaultExecutionFeeConfig() ExecutionFeeConfig {
	return ExecutionFeeConfig{
		PricePerTG:      0.0001,
		FixedCost:       1.0,
		TransactionCost: 1.0,
		OverheadCost:    0.1,
	}
}

// getDefaultFileCacheConfig provides the default file cache settings.
func getDefaultFileCacheConfig() FileCacheConfig {
	return FileCacheConfig{
		CacheDir:          "data/cache",
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		EvictionSize:      25 * 1024 * 1024,  // 25MB
		EnableCompression: true,
		MaxFileSize:       1 * 1024 * 1024, // 1MB
	}
}

// getDefaultValidationConfig provides the default code validation settings.
func getDefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxFileSize:       1 * 1024 * 1024, // 1MB
		AllowedExtensions: []string{".go", ".py", ".js", ".ts"},
		MaxComplexity:     50.0,
		TimeoutSeconds:    30,
	}
}

// getDefaultBasePoolConfig provides the default container pool settings.
func getDefaultBasePoolConfig() BasePoolConfig {
	return BasePoolConfig{
		MaxContainers:       5,
		MinContainers:       2,
		MaxWaitTime:         60 * time.Second,
		HealthCheckInterval: 10 * time.Minute,
	}
}

// getSupportedLanguages returns an array of all supported languages.
func getSupportedLanguages() []types.Language {
	return []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}
}

// getDefaultLanguagePoolConfigs creates and returns a map of default configurations
// for all supported languages.
func getDefaultLanguagePoolConfigs() map[string]LanguagePoolConfig {
	configs := make(map[string]LanguagePoolConfig)
	basePoolConfig := getDefaultBasePoolConfig()

	for _, lang := range getSupportedLanguages() {
		langStr := string(lang)
		configs[langStr] = LanguagePoolConfig{
			BasePoolConfig: basePoolConfig,
			DockerConfig:   getDefaultDockerContainerConfig(),
			LanguageConfig: getLanguageConfig(lang),
		}
	}
	return configs
}

// getLanguageConfig contains the switch statement to return specific configurations
// for a given language. The default case is Go.
func getLanguageConfig(lang types.Language) LanguageConfig {
	switch lang {
	case types.LanguagePy:
		return LanguageConfig{
			Language:    types.LanguagePy,
			ImageName:   "python:3.12-alpine",
			SetupScript: scripts.GetSetupScript(types.LanguagePy),
			RunCommand:  "python code.py",
			Extensions:  []string{".py"},
			Environment: []string{},
		}
	case types.LanguageJS:
		return LanguageConfig{
			Language:    types.LanguageJS,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetSetupScript(types.LanguageJS),
			RunCommand:  "node code.js",
			Extensions:  []string{".js"},
			Environment: []string{"V8_MEMORY_LIMIT=256"},
		}
	case types.LanguageTS:
		return LanguageConfig{
			Language:    types.LanguageTS,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetSetupScript(types.LanguageTS),
			RunCommand:  "node code.js",
			Extensions:  []string{".ts"},
			Environment: []string{"V8_MEMORY_LIMIT=256"},
		}
	case types.LanguageNode:
		return LanguageConfig{
			Language:    types.LanguageNode,
			ImageName:   "node:22-alpine",
			SetupScript: scripts.GetSetupScript(types.LanguageNode),
			RunCommand:  "node code.js",
			Extensions:  []string{".js", ".mjs", ".cjs"},
			Environment: []string{"V8_MEMORY_LIMIT=256"},
		}
	case types.LanguageGo:
		fallthrough
	default:
		return LanguageConfig{
			Language:    types.LanguageGo,
			ImageName:   "golang:1.21-alpine",
			SetupScript: scripts.GetSetupScript(types.LanguageGo),
			RunCommand:  "go run code.go",
			Extensions:  []string{".go"},
			Environment: []string{},
		}
	}
}

// getDefaultMonitoringConfig provides the default monitoring settings.
func getDefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		HealthCheckInterval:     30 * time.Second,
		MaxExecutionTime:        5 * time.Minute,
		MinSuccessRate:          0.8, // 80%
		MaxAverageExecutionTime: 2 * time.Minute,
		MaxAlerts:               100,
		AlertRetentionTime:      time.Hour,
		CriticalAlertPenalty:    20.0,
		WarningAlertPenalty:     5.0,
		HealthScoreThresholds: struct {
			Critical float64 `yaml:"critical"`	
			Warning  float64 `yaml:"warning"`
		}{
			Critical: 50.0,
			Warning:  80.0,
		},
	}
}
