package config

import "time"

type DockerConfig struct {
	Image          string   `json:"image"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	AutoCleanup    bool     `json:"auto_cleanup"`
	MemoryLimit    string   `json:"memory_limit"`
	CPULimit       float64  `json:"cpu_limit"`
	NetworkMode    string   `json:"network_mode"`
	SecurityOpt    []string `json:"security_opt"`
	ReadOnlyRootFS bool     `json:"read_only_root_fs"`
	Environment    []string `json:"environment"`
	Binds          []string `json:"binds"`
}

type FeeConfig struct {
	PricePerTG            float64 `json:"price_per_tg"`
	FixedCost             float64 `json:"fixed_cost"`
	TransactionSimulation float64 `json:"transaction_simulation"`
	OverheadCost          float64 `json:"overhead_cost"`
}

type PoolConfig struct {
	MaxContainers       int           `json:"max_containers"`
	MinContainers       int           `json:"min_containers"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	PreWarmCount        int           `json:"pre_warm_count"`
	MaxWaitTime         time.Duration `json:"max_wait_time"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

type CacheConfig struct {
	MaxCacheSize      int64         `json:"max_cache_size"`
	EvictionSize      int64         `json:"eviction_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	EnableCompression bool          `json:"enable_compression"`
	MaxFileSize       int64         `json:"max_file_size"`
}

type ValidationConfig struct {
	EnableCodeValidation bool     `json:"enable_code_validation"`
	MaxFileSize          int64    `json:"max_file_size"`
	AllowedExtensions    []string `json:"allowed_extensions"`
	BlockedPatterns      []string `json:"blocked_patterns"`
	TimeoutSeconds       int      `json:"timeout_seconds"`
}

type ExecutorConfig struct {
	Docker     DockerConfig     `json:"docker"`
	Fees       FeeConfig        `json:"fees"`
	Pool       PoolConfig       `json:"pool"`
	Cache      CacheConfig      `json:"cache"`
	Validation ValidationConfig `json:"validation"`
}
