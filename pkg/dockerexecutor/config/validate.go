package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/go-units"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

// Validate checks the entire CodeExecutorConfig structure.
// It orchestrates validation by calling the Validate method on each sub-configuration.
func (c *CodeExecutorConfig) Validate() error {
	var errors []string

	if err := c.Fees.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("fees config error: %v", err))
	}
	if err := c.Cache.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("cache config error: %v", err))
	}
	if err := c.Validation.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("validation config error: %v", err))
	}
	if err := c.Monitoring.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("monitoring config error: %v", err))
	}

	for langKey, langPoolCfg := range c.Languages {
		if err := langPoolCfg.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("language config '%s' error: %v", langKey, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the DockerContainerConfig fields.
func (c *DockerContainerConfig) Validate() error {
	var errors []string

	if c.Image == "" {
		errors = append(errors, "image cannot be empty")
	} else if !isValidDockerImage(c.Image) {
		errors = append(errors, "invalid docker image format")
	}

	if c.TimeoutSeconds <= 0 {
		errors = append(errors, "timeout_seconds must be positive")
	}

	if _, err := units.RAMInBytes(c.MemoryLimit); err != nil {
		errors = append(errors, fmt.Sprintf("invalid memory_limit format: %v", err))
	}

	if c.CPULimit <= 0 {
		errors = append(errors, "cpu_limit must be positive")
	}

	validNetworkModes := []string{"bridge", "host", "none", "container:", "default"}
	if c.NetworkMode != "" && !isValidNetworkMode(c.NetworkMode, validNetworkModes) {
		errors = append(errors, "invalid network_mode")
	}

	for _, env := range c.Environment {
		if !strings.Contains(env, "=") {
			errors = append(errors, fmt.Sprintf("invalid environment variable format: %s", env))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the ExecutionFeeConfig fields.
func (c *ExecutionFeeConfig) Validate() error {
	var errors []string

	if c.PricePerTG < 0 {
		errors = append(errors, "price_per_tg cannot be negative")
	}
	if c.FixedCost < 0 {
		errors = append(errors, "fixed_cost cannot be negative")
	}
	if c.TransactionCost < 0 {
		errors = append(errors, "transaction_cost cannot be negative")
	}
	if c.OverheadCost < 0 {
		errors = append(errors, "overhead_cost cannot be negative")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the BasePoolConfig fields.
func (c *BasePoolConfig) Validate() error {
	var errors []string

	if c.MaxContainers <= 0 {
		errors = append(errors, "max_containers must be positive")
	}
	if c.MinContainers < 0 {
		errors = append(errors, "min_containers cannot be negative")
	}
	if c.MinContainers > c.MaxContainers {
		errors = append(errors, "min_containers cannot exceed max_containers")
	}
	if c.MaxWaitTime <= 0 {
		errors = append(errors, "max_wait_time must be positive")
	}
	if c.HealthCheckInterval <= 0 {
		errors = append(errors, "health_check_interval must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the LanguageConfig fields.
func (c *LanguageConfig) Validate() error {
	var errors []string

	if !isValidLanguage(c.Language) {
		errors = append(errors, fmt.Sprintf("unsupported language: %s", c.Language))
	}
	if c.ImageName == "" {
		errors = append(errors, "image_name cannot be empty")
	}
	if c.RunCommand == "" {
		errors = append(errors, "run_command cannot be empty")
	}
	if len(c.Extensions) == 0 {
		errors = append(errors, "at least one file extension must be specified")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the LanguagePoolConfig fields by validating its composed parts.
func (c *LanguagePoolConfig) Validate() error {
	if err := c.BasePoolConfig.Validate(); err != nil {
		return fmt.Errorf("base pool config is invalid: %w", err)
	}
	if err := c.LanguageConfig.Validate(); err != nil {
		return fmt.Errorf("language-specific config is invalid: %w", err)
	}
	return nil
}

// Validate checks the FileCacheConfig fields.
func (c *FileCacheConfig) Validate() error {
	var errors []string

	if c.CacheDir == "" {
		errors = append(errors, "cache_dir cannot be empty")
	} else if !filepath.IsAbs(c.CacheDir) && !strings.HasPrefix(c.CacheDir, "data/") {
		errors = append(errors, "cache_dir should be an absolute path or relative to 'data/'")
	}
	if c.MaxCacheSize <= 0 {
		errors = append(errors, "max_cache_size must be positive")
	}
	if c.EvictionSize > 0 && c.EvictionSize >= c.MaxCacheSize {
		errors = append(errors, "eviction_size must be less than max_cache_size")
	}
	if c.MaxFileSize <= 0 {
		errors = append(errors, "max_file_size must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate checks the ValidationConfig fields.
func (c *ValidationConfig) Validate() error {
	var errors []string

	if c.MaxFileSize <= 0 {
		errors = append(errors, "max_file_size must be positive")
	}
	for _, ext := range c.AllowedExtensions {
		if !strings.HasPrefix(ext, ".") {
			errors = append(errors, fmt.Sprintf("extension must start with a dot: %s", ext))
		}
	}
	if c.MaxComplexity < 0 {
		errors = append(errors, "max_complexity cannot be negative")
	}
	if c.TimeoutSeconds <= 0 {
		errors = append(errors, "timeout_seconds must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func isValidDockerImage(image string) bool {
	if image == "" {
		return false
	}
	// A simple regex for Docker image format. More complex validation could be used.
	imageRegex := regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$`)
	return imageRegex.MatchString(strings.ToLower(image))
}

func isValidNetworkMode(mode string, validModes []string) bool {
	for _, valid := range validModes {
		if mode == valid || strings.HasPrefix(mode, valid) {
			return true
		}
	}
	return false
}

func isValidLanguage(lang types.Language) bool {
	supportedLanguages := []types.Language{
		types.LanguageGo,
		types.LanguagePy,
		types.LanguageJS,
		types.LanguageTS,
		types.LanguageNode,
	}
	for _, supported := range supportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// Validate checks the MonitoringConfig fields.
func (c *MonitoringConfig) Validate() error {
	var errors []string

	if c.HealthCheckInterval <= 0 {
		errors = append(errors, "health_check_interval must be positive")
	}
	if c.MaxExecutionTime <= 0 {
		errors = append(errors, "max_execution_time must be positive")
	}
	if c.MinSuccessRate < 0 || c.MinSuccessRate > 1 {
		errors = append(errors, "min_success_rate must be between 0 and 1")
	}
	if c.MaxAverageExecutionTime <= 0 {
		errors = append(errors, "max_average_execution_time must be positive")
	}
	if c.MaxAlerts <= 0 {
		errors = append(errors, "max_alerts must be positive")
	}
	if c.AlertRetentionTime <= 0 {
		errors = append(errors, "alert_retention_time must be positive")
	}
	if c.CriticalAlertPenalty < 0 {
		errors = append(errors, "critical_alert_penalty cannot be negative")
	}
	if c.WarningAlertPenalty < 0 {
		errors = append(errors, "warning_alert_penalty cannot be negative")
	}
	if c.HealthScoreThresholds.Critical < 0 || c.HealthScoreThresholds.Critical > 100 {
		errors = append(errors, "health_score_thresholds.critical must be between 0 and 100")
	}
	if c.HealthScoreThresholds.Warning < 0 || c.HealthScoreThresholds.Warning > 100 {
		errors = append(errors, "health_score_thresholds.warning must be between 0 and 100")
	}
	if c.HealthScoreThresholds.Critical >= c.HealthScoreThresholds.Warning {
		errors = append(errors, "health_score_thresholds.critical must be less than warning")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}
