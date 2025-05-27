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
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	LogRetryAttempt bool
	// Status codes that should trigger a retry
	RetryStatusCodes []int
	// HTTP client configuration
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

// DefaultHTTPRetryConfig returns default configuration for HTTP retry operations
func DefaultHTTPRetryConfig() *HTTPRetryConfig {
	return &HTTPRetryConfig{
		MaxRetries:      3,
		InitialDelay:    time.Second,
		MaxDelay:        10 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:             3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
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
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
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

// DoWithRetry performs an HTTP request with retry logic
func (c *HTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	attempts := 0
	delay := c.config.InitialDelay

	for attempts < c.config.MaxRetries {
		attempts++

		// Clone the request to ensure we can retry it
		reqClone := req.Clone(req.Context())
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading request body: %v", err)
			}
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			reqClone.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err := c.client.Do(reqClone)
		if err != nil {
			lastErr = err
			if attempts == c.config.MaxRetries {
				break
			}
			if c.config.LogRetryAttempt {
				c.logger.Warnf("Attempt %d failed: %v. Retrying in %v...", attempts, err, delay)
			}
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * c.config.BackoffFactor)
			continue
		}

		// Check if we should retry based on status code
		retryable := false
		for _, retryCode := range c.config.RetryStatusCodes {
			if resp.StatusCode == retryCode {
				retryable = true
				break
			}
		}

		if !retryable {
			return resp, nil
		}

		// Read and close the response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastResp = resp
		lastErr = fmt.Errorf("received retryable status code: %d, body: %s", resp.StatusCode, string(body))

		if attempts == c.config.MaxRetries {
			break
		}

		if c.config.LogRetryAttempt {
			c.logger.Warnf("Attempt %d failed with status %d. Retrying in %v...", attempts, resp.StatusCode, delay)
		}
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * c.config.BackoffFactor)
	}

	if lastResp != nil {
		return lastResp, lastErr
	}
	return nil, fmt.Errorf("failed after %d attempts: %v", attempts, lastErr)
}
