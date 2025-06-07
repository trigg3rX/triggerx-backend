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
	Config
	// Status codes that should trigger a retry
	RetryStatusCodes []int
	// HTTP client configuration
	Timeout         time.Duration
	IdleConnTimeout time.Duration
}

// DefaultHTTPRetryConfig returns default configuration for HTTP retry operations
func DefaultHTTPRetryConfig() *HTTPRetryConfig {
	return &HTTPRetryConfig{
		Config: Config{
			MaxRetries:      3,
			InitialDelay:    time.Second,
			MaxDelay:        10 * time.Second,
			BackoffFactor:   2.0,
			JitterFactor:    0.1,
			LogRetryAttempt: true,
		},
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:         3 * time.Second,
		IdleConnTimeout: 30 * time.Second,
	}
}

// HTTPClient is a wrapper around http.Client that includes retry logic
type HTTPClient struct {
	client *http.Client
	config *HTTPRetryConfig
	logger logging.Logger
}

// NewHTTPClient creates a new HTTP client with retry capabilities
func NewHTTPClient(config *HTTPRetryConfig, logger logging.Logger) *HTTPClient {
	if config == nil {
		config = DefaultHTTPRetryConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			IdleConnTimeout:     config.IdleConnTimeout,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   config.Timeout / 2,
				KeepAlive: config.IdleConnTimeout,
			}).DialContext,
			TLSHandshakeTimeout:   config.Timeout / 2,
			ResponseHeaderTimeout: config.Timeout / 2,
			ExpectContinueTimeout: config.Timeout / 3,
		},
	}

	return &HTTPClient{
		client: client,
		config: config,
		logger: logger,
	}
}

// DoWithRetry performs an HTTP request with retry logic.
// The caller is responsible for closing the response body.
func (c *HTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	var (
		lastErr  error
		lastResp *http.Response
		attempt  int
		delay    = c.config.InitialDelay
	)

	// Store the original body if it exists
	var bodyBytes []byte
	var err error
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %w", err)
		}
		if err := req.Body.Close(); err != nil {
			c.logger.Warnf("Failed to close request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	for attempt = 1; attempt <= c.config.MaxRetries; attempt++ {
		// Clone the request for each attempt
		reqClone := req.Clone(req.Context())
		if bodyBytes != nil {
			reqClone.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err := c.client.Do(reqClone)
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			if attempt == c.config.MaxRetries {
				break
			}

			if c.config.LogRetryAttempt {
				c.logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", attempt, c.config.MaxRetries, err, delay)
			}

			select {
			case <-time.After(delay):
				delay = c.nextDelay(delay)
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		// Check if we should retry based on status code
		if !c.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// Read and close the response body for retryable responses
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		lastResp = resp
		lastErr = fmt.Errorf("received retryable status code: %d, body: %s", resp.StatusCode, string(body))

		if attempt == c.config.MaxRetries {
			break
		}

		if c.config.LogRetryAttempt {
			c.logger.Warnf("Attempt %d/%d failed with status %d. Retrying in %v...", attempt, c.config.MaxRetries, resp.StatusCode, delay)
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

// shouldRetry checks if the status code is in the list of retryable status codes
func (c *HTTPClient) shouldRetry(statusCode int) bool {
	for _, retryCode := range c.config.RetryStatusCodes {
		if statusCode == retryCode {
			return true
		}
	}
	return false
}

// nextDelay calculates the next delay with backoff and jitter
func (c *HTTPClient) nextDelay(currentDelay time.Duration) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * c.config.BackoffFactor)

	// Add jitter
	if c.config.JitterFactor > 0 {
		jitter := time.Duration(float64(nextDelay) * c.config.JitterFactor)
		nextDelay += time.Duration(float64(jitter) * (0.5 - secureFloat64()))
	}

	// Cap at max delay
	if nextDelay > c.config.MaxDelay {
		nextDelay = c.config.MaxDelay
	}

	return nextDelay
}
