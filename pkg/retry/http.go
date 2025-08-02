package retry

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// HTTPRetryConfig holds configuration for HTTP retry operations
type HTTPRetryConfig struct {
	RetryConfig     *RetryConfig
	Timeout         time.Duration
	IdleConnTimeout time.Duration
	MaxResponseSize int64 // Maximum response size to read for error messages
}

// DefaultHTTPRetryConfig returns default configuration for HTTP retry operations
func DefaultHTTPRetryConfig() *HTTPRetryConfig {
	return &HTTPRetryConfig{
		RetryConfig:     DefaultRetryConfig(),
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
	return c.RetryConfig.validate()
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

// DoWithRetry performs an HTTP request with retry logic.
// The caller is responsible for closing the response body.
func (c *HTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	var (
		lastErr  error
		lastResp *http.Response
		attempt  int
		delay    = c.HTTPConfig.RetryConfig.InitialDelay
	)

	// Use GetBody if available to avoid reading into memory
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

	for attempt = 1; attempt <= c.HTTPConfig.RetryConfig.MaxRetries; attempt++ {
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
			lastErr = fmt.Errorf("http request failed: %w", err)
			if attempt == c.HTTPConfig.RetryConfig.MaxRetries {
				break
			}

			if c.HTTPConfig.RetryConfig.LogRetryAttempt {
				c.logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt, c.HTTPConfig.RetryConfig.MaxRetries, err, delay)
			}

			select {
			case <-time.After(delay):
				delay = c.nextDelay(delay)
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		// Check if we should retry based on status code or custom predicate
		if !c.shouldRetry(resp.StatusCode) && (c.HTTPConfig.RetryConfig.ShouldRetry == nil || !c.HTTPConfig.RetryConfig.ShouldRetry(nil)) {
			return resp, nil
		}

		// For retryable responses, read and close the body (unless it's the final attempt)
		if attempt < c.HTTPConfig.RetryConfig.MaxRetries {
			// Drain body without storing for non-final attempts
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, c.HTTPConfig.MaxResponseSize))
			_ = resp.Body.Close()
		} else {
			// On final attempt, keep the body open for the caller
			lastResp = resp
		}

		// Create error with limited body content
		var bodyPreview string
		if attempt == c.HTTPConfig.RetryConfig.MaxRetries && lastResp != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(lastResp.Body, c.HTTPConfig.MaxResponseSize))
			_ = lastResp.Body.Close()
			lastResp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			bodyPreview = fmt.Sprintf(", body preview: %q", truncate(string(bodyBytes), 200))
		}

		lastErr = fmt.Errorf("received retryable status code: %d%s", resp.StatusCode, bodyPreview)

		if attempt == c.HTTPConfig.RetryConfig.MaxRetries {
			break
		}

		if c.HTTPConfig.RetryConfig.LogRetryAttempt {
			c.logger.Warnf("Attempt %d/%d failed with status %d%s. Retrying in %v...",
				attempt, c.HTTPConfig.RetryConfig.MaxRetries, resp.StatusCode, bodyPreview, delay)
		}

		select {
		case <-time.After(delay):
			delay = c.nextDelay(delay)
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	if lastResp != nil {
		return lastResp, lastErr
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", attempt, lastErr)
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// shouldRetry checks if the status code is in the list of retryable status codes
func (c *HTTPClient) shouldRetry(statusCode int) bool {
	for _, retryCode := range c.HTTPConfig.RetryConfig.StatusCodes {
		if statusCode == retryCode {
			return true
		}
	}
	return false
}

// nextDelay calculates the next delay with backoff and jitter
func (c *HTTPClient) nextDelay(currentDelay time.Duration) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * c.HTTPConfig.RetryConfig.BackoffFactor)
	jitter := time.Duration(float64(nextDelay) * c.HTTPConfig.RetryConfig.JitterFactor)
	nextDelay += time.Duration(float64(jitter) * (0.5 - secureFloat64()))

	// Cap at max delay
	if nextDelay > c.HTTPConfig.RetryConfig.MaxDelay {
		nextDelay = c.HTTPConfig.RetryConfig.MaxDelay
	}

	return nextDelay
}

func (c *HTTPClient) Close() {
	c.client.CloseIdleConnections()
}

func (c *HTTPClient) GetTimeout() time.Duration {
	return c.HTTPConfig.Timeout
}

func (c *HTTPClient) GetIdleConnTimeout() time.Duration {
	return c.HTTPConfig.IdleConnTimeout
}
