package http

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// HTTPRetryConfig holds configuration for HTTP retry operations
type HTTPRetryConfig struct {
	RetryConfig     *retry.RetryConfig
	Timeout         time.Duration
	IdleConnTimeout time.Duration
	MaxResponseSize int64 // Maximum response size to read for error messages
}

// DefaultHTTPRetryConfig returns default configuration for HTTP retry operations
func DefaultHTTPRetryConfig() *HTTPRetryConfig {
	return &HTTPRetryConfig{
		RetryConfig:     retry.DefaultRetryConfig(),
		Timeout:         10 * time.Second,
		IdleConnTimeout: 30 * time.Second,
		MaxResponseSize: 4096, // 4KB default max for error messages
	}
}

// Validate checks the HTTP configuration for reasonable values
func (c *HTTPRetryConfig) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.IdleConnTimeout <= 0 {
		return fmt.Errorf("idleConnTimeout must be positive")
	}
	if c.MaxResponseSize < 0 {
		return fmt.Errorf("maxResponseSize must be >= 0")
	}
	return nil
}

// HTTPError represents an HTTP-specific error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// HTTPClient is a wrapper around http.Client that includes retry logic
type HTTPClient struct {
	client     *http.Client
	HTTPConfig *HTTPRetryConfig
	logger     logging.Logger
}

// NewHTTPClient creates a new HTTP client with retry capabilities
func NewHTTPClient(httpConfig *HTTPRetryConfig, logger logging.Logger) (*HTTPClient, error) {
	if httpConfig == nil {
		httpConfig = DefaultHTTPRetryConfig()
	}

	if err := httpConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HTTP retry config: %w", err)
	}

	client := &http.Client{
		Timeout: httpConfig.Timeout,
		Transport: &http.Transport{
			IdleConnTimeout:   httpConfig.IdleConnTimeout,
			DisableKeepAlives: false,
			DialContext: (&net.Dialer{
				Timeout:   httpConfig.Timeout / 2,
				KeepAlive: httpConfig.IdleConnTimeout,
			}).DialContext,
			TLSHandshakeTimeout:   httpConfig.Timeout / 2,
			ResponseHeaderTimeout: httpConfig.Timeout / 2,
			ExpectContinueTimeout: httpConfig.Timeout / 3,
		},
	}

	return &HTTPClient{
		client:     client,
		HTTPConfig: httpConfig,
		logger:     logger,
	}, nil
}

// DoWithRetry performs an HTTP request with retry logic using the retry package.
// The caller is responsible for closing the response body.
func (c *HTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	// Prepare request body for retries
	var getBody func() (io.ReadCloser, error)
	if req.GetBody != nil {
		getBody = req.GetBody
	} else if req.Body != nil {
		// Fallback for requests without GetBody
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %w", err)
		}
		if err := req.Body.Close(); err != nil {
			c.logger.Warnf("Failed to close request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(bodyBytes)), nil
		}
	}

	// Create operation function for retry package
	operation := func() (*http.Response, error) {
		// Clone the request for each attempt
		reqClone := req.Clone(req.Context())
		if getBody != nil {
			body, err := getBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body: %w", err)
			}
			reqClone.Body = body
		}

		resp, err := c.client.Do(reqClone)
		if err != nil {
			return nil, fmt.Errorf("http request failed: %w", err)
		}

		// Check if response status code indicates retryable error
		if c.shouldRetryResponse(resp.StatusCode) {
			// Read and close body for retryable responses
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, c.HTTPConfig.MaxResponseSize))
			_ = resp.Body.Close()

			bodyPreview := fmt.Sprintf(", body preview: %q", truncate(string(bodyBytes), 200))
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("retryable status code%s", bodyPreview),
			}
		}

		return resp, nil
	}

	// Use retry package to execute the operation
	return retry.Retry(req.Context(), operation, c.HTTPConfig.RetryConfig, c.logger)
}

// Get performs a GET request with retry logic
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	return c.DoWithRetry(req)
}

// Post performs a POST request with retry logic
func (c *HTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	return c.DoWithRetry(req)
}

// Put performs a PUT request with retry logic
func (c *HTTPClient) Put(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	return c.DoWithRetry(req)
}

// Delete performs a DELETE request with retry logic
func (c *HTTPClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}
	return c.DoWithRetry(req)
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// shouldRetryResponse checks if the status code should be retried based on default logic and custom retry function
func (c *HTTPClient) shouldRetryResponse(statusCode int) bool {
	// Create a mock error to test with custom retry function
	mockErr := &HTTPError{
		StatusCode: statusCode,
		Message:    "mock error for retry check",
	}

	// If custom retry function is set, use it
	if c.HTTPConfig.RetryConfig.ShouldRetry != nil {
		return c.HTTPConfig.RetryConfig.ShouldRetry(mockErr)
	}

	// Default logic: retry on server errors (5xx) and some specific client errors
	return statusCode >= 500 || statusCode == 429 // 429 is rate limit
}

// isRetryableStatusCode checks if the status code indicates a retryable error (deprecated, use shouldRetryResponse)
func (c *HTTPClient) isRetryableStatusCode(statusCode int) bool {
	return c.shouldRetryResponse(statusCode)
}

// Close closes idle connections
func (c *HTTPClient) Close() {
	c.client.CloseIdleConnections()
}

// GetTimeout returns the configured timeout
func (c *HTTPClient) GetTimeout() time.Duration {
	return c.HTTPConfig.Timeout
}

// GetIdleConnTimeout returns the configured idle connection timeout
func (c *HTTPClient) GetIdleConnTimeout() time.Duration {
	return c.HTTPConfig.IdleConnTimeout
}
