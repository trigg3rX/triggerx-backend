package config

import (
	"os"
	"time"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWT Configuration
	JWTSecretKey     string        `json:"jwt_secret_key"`
	JWTExpiration    time.Duration `json:"jwt_expiration"`
	JWTRefreshExpiry time.Duration `json:"jwt_refresh_expiry"`

	// Token Configuration
	AccessTokenExpiry  time.Duration `json:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry"`

	// Security Configuration
	EnableRateLimit bool          `json:"enable_rate_limit"`
	MaxRequests     int           `json:"max_requests"`
	WindowSize      time.Duration `json:"window_size"`

	// Service Configuration
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
}

// NewAuthConfig creates a new authentication configuration with defaults
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWTSecretKey:       getEnvOrDefault("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		JWTExpiration:      getDurationEnvOrDefault("JWT_EXPIRATION", 24*time.Hour),
		JWTRefreshExpiry:   getDurationEnvOrDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		AccessTokenExpiry:  getDurationEnvOrDefault("ACCESS_TOKEN_EXPIRY", 1*time.Hour),
		RefreshTokenExpiry: getDurationEnvOrDefault("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		EnableRateLimit:    getBoolEnvOrDefault("ENABLE_RATE_LIMIT", false),
		MaxRequests:        getIntEnvOrDefault("MAX_REQUESTS", 100),
		WindowSize:         getDurationEnvOrDefault("RATE_LIMIT_WINDOW", 1*time.Minute),
		ServiceName:        getEnvOrDefault("SERVICE_NAME", "triggerx-backend"),
		Environment:        getEnvOrDefault("ENVIRONMENT", "development"),
	}
}

// GetJWTSecretKey returns the JWT secret key
func (c *AuthConfig) GetJWTSecretKey() string {
	return c.JWTSecretKey
}

// GetJWTExpiration returns the JWT expiration duration
func (c *AuthConfig) GetJWTExpiration() time.Duration {
	return c.JWTExpiration
}

// GetAccessTokenExpiry returns the access token expiry duration
func (c *AuthConfig) GetAccessTokenExpiry() time.Duration {
	return c.AccessTokenExpiry
}

// GetRefreshTokenExpiry returns the refresh token expiry duration
func (c *AuthConfig) GetRefreshTokenExpiry() time.Duration {
	return c.RefreshTokenExpiry
}

// IsRateLimitEnabled returns whether rate limiting is enabled
func (c *AuthConfig) IsRateLimitEnabled() bool {
	return c.EnableRateLimit
}

// GetMaxRequests returns the maximum number of requests allowed
func (c *AuthConfig) GetMaxRequests() int {
	return c.MaxRequests
}

// GetWindowSize returns the rate limit window size
func (c *AuthConfig) GetWindowSize() time.Duration {
	return c.WindowSize
}

// IsDevelopment returns whether the environment is development
func (c *AuthConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns whether the environment is production
func (c *AuthConfig) IsProduction() bool {
	return c.Environment == "production"
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := time.ParseDuration(value); err == nil {
			return int(intValue)
		}
	}
	return defaultValue
}
