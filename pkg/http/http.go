package http

import (
	"bytes"
	"context"
	"errors"
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

	// Set up default HTTP retry predicate if none provided
	if httpConfig.RetryConfig.ShouldRetry == nil {
		httpConfig.RetryConfig.ShouldRetry = func(err error, attempt int) bool {
			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				// This is an error we created from a status code.
				// Retry on 5xx and 429.
				return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
			}
			// For all other errors (network errors, etc.), assume they are retryable.
			return true
		}
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
func (c *HTTPClient) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Prepare request body for retries
	var getBody func() (io.ReadCloser, error)
	if req.GetBody != nil {
		getBody = func() (io.ReadCloser, error) {
			return req.GetBody()
		}
	} else if req.Body != nil {
		// Fallback for requests without GetBody
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body for retry: %w", err)
		}
		if err := req.Body.Close(); err != nil {
			c.logger.Warnf("Failed to close request body: %v", err)
		}
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(bodyBytes)), nil
		}
	}

	// Also reset the body for the *first* attempt if GetBody is available.
	if req.GetBody != nil {
		req.Body, _ = req.GetBody()
	}

	// Create operation function for retry package
	operation := func() (*http.Response, error) {
		// Clone the request for each attempt
		reqClone := req.Clone(ctx)
		if getBody != nil {
			body, err := getBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body: %w", err)
			}
			reqClone.Body = body
		}

		resp, err := c.client.Do(reqClone)
		if err != nil {
			// Network errors are now handled directly by the retry predicate.
			return nil, err
		}

		// Check if status code indicates a potentially retryable error (5xx or 429)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			// Create a test error to check if this should be retried
			testErr := &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    "test error for retry check",
			}

			// Check if the custom retry logic would retry this error
			shouldRetry := true // Default behavior
			if c.HTTPConfig.RetryConfig.ShouldRetry != nil {
				shouldRetry = c.HTTPConfig.RetryConfig.ShouldRetry(testErr, 1)
			}

			if shouldRetry {
				// Convert to error for retry logic
				bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, c.HTTPConfig.MaxResponseSize))
				err := resp.Body.Close()
				if err != nil {
					c.logger.Warnf("Failed to close response body: %v", err)
				}

				return nil, &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    fmt.Sprintf("received retryable status code, body: %q", truncate(string(bodyBytes), 200)),
				}
			}
			// If not retryable, return the response as-is
		}

		return resp, nil
	}

	// Use retry package to execute the operation
	return retry.Retry(ctx, operation, c.HTTPConfig.RetryConfig, c.logger)
}

// Get performs a GET request with retry logic
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	return c.DoWithRetry(ctx, req)
}

// Post performs a POST request with retry logic
func (c *HTTPClient) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	return c.DoWithRetry(ctx, req)
}

// Put performs a PUT request with retry logic
func (c *HTTPClient) Put(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	return c.DoWithRetry(ctx, req)
}

// Delete performs a DELETE request with retry logic
func (c *HTTPClient) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}
	return c.DoWithRetry(ctx, req)
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

// GetClient returns the underlying http.Client for use with other libraries
func (c *HTTPClient) GetClient() *http.Client {
	return c.client
}
