package config

import "time"

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		CacheTTL:          1 * time.Hour,
		CleanupInterval:   10 * time.Minute,
		EnableCompression: true,
		MaxFileSize:       10 * 1024 * 1024, // 10MB
	}
}

func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		EnableCodeValidation: true,
		MaxFileSize:          10 * 1024 * 1024, // 10MB
		AllowedExtensions:    []string{".go", ".py", ".js", ".ts"},
		BlockedPatterns: []string{
			"os.RemoveAll",
			"exec.Command",
			"syscall",
			"runtime.GC",
			"os.Exit",
			"panic(",
		},
		TimeoutSeconds: 30,
	}
}
