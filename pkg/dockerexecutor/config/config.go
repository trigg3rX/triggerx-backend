package config

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

// DockerContainerConfig is the configuration for a Docker container used by ContainerManager
type DockerContainerConfig struct {
	Image          string           `yaml:"image"`
	TimeoutSeconds int              `yaml:"timeout_seconds"`
	AutoCleanup    bool             `yaml:"auto_cleanup"`
	MemoryLimit    string           `yaml:"memory_limit"`
	CPULimit       float64          `yaml:"cpu_limit"`
	NetworkMode    string           `yaml:"network_mode"`
	SecurityOpt    []string         `yaml:"security_opt"`
	ReadOnlyRootFS bool             `yaml:"read_only_root_fs"`
	Environment    []string         `yaml:"environment"`
	Binds          []string         `yaml:"binds"`
	Languages      []types.Language `yaml:"languages"`
}

// ExecutionFeeConfig is the configuration for the execution fee
type ExecutionFeeConfig struct {
	PricePerTG      float64 `yaml:"price_per_tg"`     // Price per TG in ether
	TransactionCost float64 `yaml:"transaction_cost"` // Cost of the action transaction
	FixedCost       float64 `yaml:"fixed_cost"`       // TriggerX fee - 0.1%
	StaticComplexityFactor    float64 `yaml:"static_complexity_factor"`    // Static complexity factor
	DynamicComplexityFactor   float64 `yaml:"dynamic_complexity_factor"`   // Dynamic complexity factor
}

type BasePoolConfig struct {
	MaxContainers       int           `yaml:"max_containers"`
	MinContainers       int           `yaml:"min_containers"`
	MaxWaitTime         time.Duration `yaml:"max_wait_time"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
}

type LanguageConfig struct {
	Language    types.Language `yaml:"language"`
	ImageName   string         `yaml:"image_name"`
	SetupScript string         `yaml:"setup_script"`
	RunCommand  string         `yaml:"run_command"`
	Extensions  []string       `yaml:"extensions"`
	Environment []string       `yaml:"environment"`
}

// LanguagePoolConfig is the configuration for a language pool
type LanguagePoolConfig struct {
	BasePoolConfig BasePoolConfig        `yaml:"base_config"`
	DockerConfig   DockerContainerConfig `yaml:"docker_config"`
	LanguageConfig LanguageConfig        `yaml:"language_config"`
}

// FileCacheConfig is the configuration for the file cache
type FileCacheConfig struct {
	CacheDir          string `yaml:"cache_dir"`
	MaxCacheSize      int64  `yaml:"max_cache_size"`
	EvictionSize      int64  `yaml:"eviction_size"`
	EnableCompression bool   `yaml:"enable_compression"`
	MaxFileSize       int64  `yaml:"max_file_size"`
}

// ValidationConfig is the configuration for the validation of code before execution
type ValidationConfig struct {
	MaxFileSize       int64    `yaml:"max_file_size"`
	AllowedExtensions []string `yaml:"allowed_extensions"`
	MaxComplexity     float64  `yaml:"max_complexity"`
	TimeoutSeconds    int      `yaml:"timeout_seconds"`
}

// MonitoringConfig is the configuration for execution monitoring
type MonitoringConfig struct {
	HealthCheckInterval     time.Duration `yaml:"health_check_interval"`
	MaxExecutionTime        time.Duration `yaml:"max_execution_time"`
	MinSuccessRate          float64       `yaml:"min_success_rate"`
	MaxAverageExecutionTime time.Duration `yaml:"max_average_execution_time"`
	MaxAlerts               int           `yaml:"max_alerts"`
	AlertRetentionTime      time.Duration `yaml:"alert_retention_time"`
	CriticalAlertPenalty    float64       `yaml:"critical_alert_penalty"`
	WarningAlertPenalty     float64       `yaml:"warning_alert_penalty"`
	HealthScoreThresholds   struct {
		Critical float64 `yaml:"critical"`
		Warning  float64 `yaml:"warning"`
	} `yaml:"health_score_thresholds"`
}

type ManagerConfig struct {
	AutoCleanup bool `yaml:"auto_cleanup"`
}

// CodeExecutorConfig is the configuration for the executor
type CodeExecutorConfig struct {
	Manager    ManagerConfig                 `yaml:"manager"`
	Fees       ExecutionFeeConfig            `yaml:"fees"`
	Languages  map[string]LanguagePoolConfig `yaml:"languages"`
	Cache      FileCacheConfig               `yaml:"cache"`
	Validation ValidationConfig              `yaml:"validation"`
	Monitoring MonitoringConfig              `yaml:"monitoring"`
}
