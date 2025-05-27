package middleware

import (
	"bytes"
	"fmt"
	"io"
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
		},
	}
}

// RetryMiddleware creates a new retry middleware
func RetryMiddleware(config *RetryConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return func(c *gin.Context) {
		logger := logging.GetServiceLogger()

		// Skip retry for non-idempotent methods
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Create a copy of the request body if it exists
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create a response recorder
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// Remove the initial handler call and status check
		var lastErr error
		attempts := 0
		_, err := retry.Retry(func() (interface{}, error) {
			attempts++
			// Reset the response writer for each attempt
			w.body.Reset()
			w.statusCode = 0

			// Restore the request body for each attempt
			if bodyBytes != nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Process the request
			c.Handler()(c)

			// Check if we should retry based on status code
			statusCode := w.statusCode
			retryable := false
			for _, retryCode := range config.RetryStatusCodes {
				if statusCode == retryCode {
					retryable = true
					break
				}
			}

			if !retryable {
				// Success, do not retry
				return nil, nil
			}

			lastErr = fmt.Errorf("received retryable status code: %d", statusCode)
			return nil, lastErr
		}, &retry.Config{
			MaxRetries:      config.MaxRetries,
			InitialDelay:    config.InitialDelay,
			MaxDelay:        config.MaxDelay,
			BackoffFactor:   config.BackoffFactor,
			JitterFactor:    config.JitterFactor,
			LogRetryAttempt: config.LogRetryAttempt,
		}, logger)

		if err != nil {
			// If all retries failed, return the last error
			c.JSON(w.statusCode, gin.H{
				"error": fmt.Sprintf("Request failed after %d attempts: %v", attempts, lastErr),
			})
			c.Abort()
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
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
