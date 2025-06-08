package middleware

import (
	"bytes"
	"fmt"
	"io"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// RetryConfig holds configuration for API retry operations
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	LogRetryAttempt bool
	// Status codes that should trigger a retry
	RetryStatusCodes []int
}

// DefaultRetryConfig returns default configuration for API retry operations
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
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
			http.StatusTooManyRequests, // Add rate limit status code
			http.StatusRequestTimeout,
			http.StatusConflict, // Add conflict status code
		},
	}
}

// RetryMiddleware creates a new retry middleware
func RetryMiddleware(config *RetryConfig, logger logging.Logger) gin.HandlerFunc {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return func(c *gin.Context) {
		// Skip retry for non-idempotent methods
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Create a copy of the request body if it exists
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
				if err := c.Request.Body.Close(); err != nil {
					logger.Warnf("Failed to close request body: %v", err)
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create a response recorder
		origWriter := c.Writer
		w := &responseWriter{
			ResponseWriter: origWriter,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		var lastErr error
		attempts := 0
		var finalStatus int
		var finalBody []byte
		_, err := retry.Retry(context.Background(), func() (interface{}, error) {
			attempts++
			w.body.Reset()
			w.statusCode = 0

			// Create a new request for each retry attempt
			newReq, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), bytes.NewBuffer(bodyBytes))
			if err != nil {
				return nil, fmt.Errorf("failed to create new request: %v", err)
			}

			// Copy headers from original request
			for key, values := range c.Request.Header {
				for _, value := range values {
					newReq.Header.Add(key, value)
				}
			}

			// Create a new response writer for this attempt
			newWriter := &responseWriter{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
			}

			// Create a new context for this attempt
			newCtx := c.Copy()
			newCtx.Request = newReq
			newCtx.Writer = newWriter

			// Process the request with the new context
			c.Handler()(newCtx)

			statusCode := newWriter.statusCode
			retryable := false
			for _, retryCode := range config.RetryStatusCodes {
				if statusCode == retryCode {
					retryable = true
					break
				}
			}

			if !retryable {
				finalStatus = newWriter.statusCode
				finalBody = newWriter.body.Bytes()
				return nil, nil
			}

			if config.LogRetryAttempt {
				logger.Warnf("Retry attempt %d for %s %s with status code %d",
					attempts, c.Request.Method, c.Request.URL.Path, statusCode)
			}

			lastErr = fmt.Errorf("received retryable status code: %d", statusCode)
			finalStatus = newWriter.statusCode
			finalBody = newWriter.body.Bytes()
			return nil, lastErr
		}, &retry.RetryConfig{
			MaxRetries:      config.MaxRetries,
			InitialDelay:    config.InitialDelay,
			MaxDelay:        config.MaxDelay,
			BackoffFactor:   config.BackoffFactor,
			JitterFactor:    config.JitterFactor,
			LogRetryAttempt: config.LogRetryAttempt,
		}, logger)

		if err != nil {
			logger.Errorf("Error retrying request: %v", err)
			if finalStatus == 0 {
				finalStatus = http.StatusInternalServerError
				finalBody = []byte("Internal server error during retry operation")
			}
		}

		// Write the final response to the original writer
		origWriter.WriteHeader(finalStatus)
		status, err := origWriter.Write(finalBody)
		if err != nil {
			logger.Errorf("Error writing final response: %v", err)
		} else {
			logger.Infof("Wrote %d bytes to response", status)
		}

		if err != nil {
			// If all retries failed, return the last error as JSON
			c.JSON(finalStatus, gin.H{
				"error": fmt.Sprintf("Request failed after %d attempts: %v", attempts, lastErr),
			})
			c.Abort()
			return
		}
	}
}

// responseWriter is a custom response writer that captures the response
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// Only write to the buffer, not to the original writer
	return w.body.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	// Don't write to the original writer yet
}
