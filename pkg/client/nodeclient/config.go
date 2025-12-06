package nodeclient

import (
	"fmt"
	"time"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	wsclient "github.com/trigg3rX/triggerx-backend/pkg/websocket"
)

// Config holds the configuration for the NodeClient
type Config struct {
	// APIKey is the Alchemy API key
	APIKey string

	// Network is the blockchain network to connect to
	Network Network

	// BaseURL is the base URL for the Alchemy API (optional, will be derived from Network if not set)
	BaseURL string

	// HTTPConfig is the HTTP client configuration
	HTTPConfig *httppkg.HTTPRetryConfig

	// WebSocketConfig is the WebSocket client configuration
	WebSocketConfig *wsclient.WebSocketRetryConfig

	// WebSocketURL is the WebSocket URL (optional, will be derived from BaseURL if not set)
	WebSocketURL string

	// Logger is the logger instance
	Logger logging.Logger

	// RequestTimeout is the timeout for individual requests
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Allow empty API key if BaseURL is set (BaseURL may already contain authentication)
	if c.APIKey == "" && c.BaseURL == "" {
		return fmt.Errorf("API key cannot be empty when BaseURL is not set")
	}

	if c.Network == "" && c.BaseURL == "" {
		return fmt.Errorf("either network or base URL must be specified")
	}

	if c.Logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}

	if c.HTTPConfig != nil {
		if err := c.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("invalid HTTP config: %w", err)
		}
	}

	if c.WebSocketConfig != nil {
		if err := c.WebSocketConfig.Validate(); err != nil {
			return fmt.Errorf("invalid WebSocket config: %w", err)
		}
	}

	return nil
}

// GetBaseURL returns the base URL for the API
func (c *Config) GetBaseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return c.Network.GetAlchemyURL()
}

// GetFullURL returns the full URL including the API key
// If BaseURL is set and APIKey is empty, returns BaseURL as-is (assumes BaseURL already contains authentication)
func (c *Config) GetFullURL() string {
	baseURL := c.GetBaseURL()
	// If BaseURL is explicitly set and APIKey is empty, assume BaseURL already contains authentication
	if c.BaseURL != "" && c.APIKey == "" {
		return baseURL
	}
	return fmt.Sprintf("%s%s", baseURL, c.APIKey)
}

// DefaultConfig returns a default configuration with sensible defaults
func DefaultConfig(apiKey string, network Network, logger logging.Logger) *Config {
	return &Config{
		APIKey:         apiKey,
		Network:        network,
		HTTPConfig:     httppkg.DefaultHTTPRetryConfig(),
		Logger:         logger,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
	}
}

// WithBaseURL sets a custom base URL (useful for testing or custom endpoints)
func (c *Config) WithBaseURL(baseURL string) *Config {
	c.BaseURL = baseURL
	return c
}

// WithHTTPConfig sets a custom HTTP configuration
func (c *Config) WithHTTPConfig(httpConfig *httppkg.HTTPRetryConfig) *Config {
	c.HTTPConfig = httpConfig
	return c
}

// WithRequestTimeout sets a custom request timeout
func (c *Config) WithRequestTimeout(timeout time.Duration) *Config {
	c.RequestTimeout = timeout
	return c
}

// WithMaxRetries sets the maximum number of retries
func (c *Config) WithMaxRetries(maxRetries int) *Config {
	c.MaxRetries = maxRetries
	return c
}

// WithWebSocketConfig sets a custom WebSocket configuration
func (c *Config) WithWebSocketConfig(wsConfig *wsclient.WebSocketRetryConfig) *Config {
	c.WebSocketConfig = wsConfig
	return c
}

// WithWebSocketURL sets a custom WebSocket URL
func (c *Config) WithWebSocketURL(wsURL string) *Config {
	c.WebSocketURL = wsURL
	return c
}
