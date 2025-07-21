package config

import "time"

type CacheConfig struct {
	MaxCacheSize      int64         `json:"max_cache_size"`
	CacheTTL          time.Duration `json:"cache_ttl"`
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
