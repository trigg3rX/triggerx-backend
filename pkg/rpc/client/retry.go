package client

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	JitterFactor   float64
	RetryableCodes []codes.Code
}

// DefaultRetryConfig returns default retry configuration for gRPC client
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		RetryableCodes: []codes.Code{
			codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
			codes.Internal,
		},
	}
}

// RetryWithBackoff retries a function with exponential backoff using the retry package
func RetryWithBackoff(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	// Convert to the retry package config
	retryConfig := &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		ShouldRetry: func(err error, attempt int) bool {
			return isRetryableError(err, config.RetryableCodes)
		},
	}

	// Use the retry package
	return retry.RetryFunc(ctx, fn, retryConfig)
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableCodes []codes.Code) bool {
	if err == nil {
		return false
	}

	// Check for protocol mismatch errors (non-retryable)
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "http2: frame too large") ||
		strings.Contains(errMsg, "frame header looked like an http/1.1 header") ||
		strings.Contains(errMsg, "protocol mismatch") ||
		strings.Contains(errMsg, "malformed http response") {
		return false // Protocol mismatch - don't retry
	}

	// Check if it's a gRPC status error
	if st, ok := status.FromError(err); ok {
		// Protocol mismatch errors often come as Unavailable, but we check the message first
		for _, code := range retryableCodes {
			if st.Code() == code {
				return true
			}
		}
	}

	// Check for connection errors (only retry on canceled if it's a temporary issue)
	if st, ok := status.FromError(err); ok && st.Code() == codes.Canceled {
		// Don't retry canceled errors by default - they're usually intentional
		return false
	}

	return false
}
