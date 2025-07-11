package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/retry"
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
			http.StatusTooManyRequests,
			http.StatusRequestTimeout,
			http.StatusConflict,
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

		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Create a copy of the request body if it exists
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Errorf("Failed to read request body: %v", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
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
			metrics.RetryAttemptsTotal.WithLabelValues(endpoint, fmt.Sprintf("%d", attempts)).Inc()

			// Reset the response writer for this attempt
			w.body.Reset()
			w.statusCode = 0

			// Create a new request for this attempt
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

			// Create a new context for this attempt
			newCtx := c.Copy()
			newCtx.Request = newReq
			newCtx.Writer = w

			// Process the request
			c.Handler()(newCtx)

			// If the request was aborted, don't retry
			if newCtx.IsAborted() {
				return nil, nil
			}

			statusCode := w.statusCode
			retryable := false
			for _, retryCode := range config.RetryStatusCodes {
				if statusCode == retryCode {
					retryable = true
					break
				}
			}

			if !retryable {
				finalStatus = w.statusCode
				finalBody = w.body.Bytes()
				metrics.RetrySuccessesTotal.WithLabelValues(endpoint).Inc()
				return nil, nil
			}

			if config.LogRetryAttempt {
				logger.Warnf("Retry attempt %d for %s %s with status code %d",
					attempts, c.Request.Method, c.Request.URL.Path, statusCode)
			}

			lastErr = fmt.Errorf("received retryable status code: %d", statusCode)
			finalStatus = w.statusCode
			finalBody = w.body.Bytes()
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
			metrics.RetryFailuresTotal.WithLabelValues(endpoint).Inc()
			if finalStatus == 0 {
				finalStatus = http.StatusInternalServerError
				finalBody = []byte("Internal server error during retry operation")
			}
		}

		// Write the final response only once
		origWriter.WriteHeader(finalStatus)
		if _, err := origWriter.Write(finalBody); err != nil {
			logger.Errorf("Error writing final response: %v", err)
		}

		// Abort the context to prevent further handlers from writing
		c.Abort()
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
