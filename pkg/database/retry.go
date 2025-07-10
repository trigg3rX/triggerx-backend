package database

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// RetryableConfig holds configuration for retryable database operations
type RetryableConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	LogRetryAttempt bool
}

// DefaultRetryableConfig returns default configuration for database operations
func DefaultRetryableConfig() *RetryableConfig {
	return &RetryableConfig{
		MaxRetries:      5,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
	}
}

// isRetryableError determines if the error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable errors
	switch err.(type) {
	case *gocql.RequestErrWriteTimeout:
		return true
	case *gocql.RequestErrReadTimeout:
		return true
	case *gocql.RequestErrUnavailable:
		return true
	case *gocql.RequestErrReadFailure:
		return true
	case *gocql.RequestErrWriteFailure:
		return true
	case *gocql.RequestErrFunctionFailure:
		return true
	case gocql.RequestErrWriteTimeout:
		return true
	case gocql.RequestErrReadTimeout:
		return true
	case gocql.RequestErrUnavailable:
		return true
	case gocql.RequestErrReadFailure:
		return true
	case gocql.RequestErrWriteFailure:
		return true
	case gocql.RequestErrFunctionFailure:
		return true
	}

	// Check error message for other retryable conditions
	errMsg := err.Error()
	retryableMessages := []string{
		"no connections available",
		"connection refused",
		"connection reset by peer",
		"i/o timeout",
		"timeout",
		"network is unreachable",
		"host is unreachable",
		"broken pipe",
		"connection aborted",
		"connection timed out",
		"temporary failure",
		"service unavailable",
		"server overloaded",
		"too many connections",
		"connection lost",
		"connection closed",
		"no route to host",
		"operation timed out",
		"network unreachable",
		"connection reset",
		"connection refused",
		"host not found",
	}

	for _, msg := range retryableMessages {
		if errMsg == msg {
			return true
		}
	}

	return false
}

// Query returns a query without retry logic for direct use
func (c *Connection) Query(query string, values ...interface{}) *gocql.Query {
	return c.session.Query(query, values...)
}

// RetryableExec executes a query with retry logic and returns error
func (c *Connection) RetryableExec(query string, values ...interface{}) error {
	return c.RetryableExecWithContext(context.Background(), query, values...)
}

// RetryableExecWithContext executes a query with retry logic and context support
func (c *Connection) RetryableExecWithContext(ctx context.Context, query string, values ...interface{}) error {
	config := DefaultRetryableConfig()

	if c.logger != nil {
		c.logger.Info("Starting retryable exec operation",
			"query", query,
			"max_retries", config.MaxRetries,
			"initial_delay", config.InitialDelay)
	}

	_, err := retry.Retry(ctx, func() (interface{}, error) {
		if err := c.session.Query(query, values...).Exec(); err != nil {
			if !isRetryableError(err) {
				if c.logger != nil {
					c.logger.Error("Non-retryable error during exec operation",
						"error", err.Error(),
						"query", query)
				}
				return nil, err
			}
			if c.logger != nil {
				c.logger.Warn("Retryable error during exec operation",
					"error", err.Error(),
					"query", query)
			}
			return nil, err
		}
		return nil, nil
	}, &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		LogRetryAttempt: config.LogRetryAttempt,
	}, c.logger)

	if err != nil && c.logger != nil {
		c.logger.Error("Exec operation failed after all retries",
			"error", err.Error(),
			"query", query,
			"max_retries", config.MaxRetries)
	}

	return err
}

// RetryableScan executes a query with retry logic and scans the result
func (c *Connection) RetryableScan(query string, dest ...interface{}) error {
	return c.RetryableScanWithContext(context.Background(), query, dest...)
}

// RetryableScanWithContext executes a query with retry logic and context support
func (c *Connection) RetryableScanWithContext(ctx context.Context, query string, dest ...interface{}) error {
	config := DefaultRetryableConfig()

	if c.logger != nil {
		c.logger.Info("Starting retryable scan operation",
			"query", query,
			"max_retries", config.MaxRetries,
			"initial_delay", config.InitialDelay)
	}

	_, err := retry.Retry(ctx, func() (interface{}, error) {
		err := c.session.Query(query).Scan(dest...)
		if err != nil && !isRetryableError(err) {
			if c.logger != nil {
				c.logger.Error("Non-retryable error during scan operation",
					"error", err.Error(),
					"query", query)
			}
			return nil, err
		}
		if err != nil && c.logger != nil {
			c.logger.Warn("Retryable error during scan operation",
				"error", err.Error(),
				"query", query)
		}
		return nil, err
	}, &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		LogRetryAttempt: config.LogRetryAttempt,
	}, c.logger)

	if err != nil && c.logger != nil {
		c.logger.Error("Scan operation failed after all retries",
			"error", err.Error(),
			"query", query,
			"max_retries", config.MaxRetries)
	}

	return err
}

// RetryableIter executes a query with retry logic and returns an iterator
func (c *Connection) RetryableIter(query string, values ...interface{}) *gocql.Iter {
	return c.RetryableIterWithContext(context.Background(), query, values...)
}

// RetryableIterWithContext executes a query with retry logic and context support
func (c *Connection) RetryableIterWithContext(ctx context.Context, query string, values ...interface{}) *gocql.Iter {
	config := DefaultRetryableConfig()

	result, err := retry.Retry(ctx, func() (interface{}, error) {
		iter := c.session.Query(query, values...).Iter()
		if err := iter.Close(); err != nil && isRetryableError(err) {
			return nil, err
		}
		// Return a fresh iterator for use
		return c.session.Query(query, values...).Iter(), nil
	}, &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		LogRetryAttempt: config.LogRetryAttempt,
	}, c.logger)

	if err != nil {
		// Return an empty iterator if all retries failed
		return c.session.Query("SELECT * FROM system.local WHERE key = 'unavailable'").Iter()
	}

	return result.(*gocql.Iter)
}

// RetryableBatch executes a batch with retry logic
func (c *Connection) RetryableBatch(batch *gocql.Batch) error {
	return c.RetryableBatchWithContext(context.Background(), batch)
}

// RetryableBatchWithContext executes a batch with retry logic and context support
func (c *Connection) RetryableBatchWithContext(ctx context.Context, batch *gocql.Batch) error {
	config := DefaultRetryableConfig()

	if c.logger != nil {
		c.logger.Info("Starting retryable batch operation",
			"batch_size", len(batch.Entries),
			"max_retries", config.MaxRetries,
			"initial_delay", config.InitialDelay)
	}

	_, err := retry.Retry(ctx, func() (interface{}, error) {
		err := c.session.ExecuteBatch(batch)
		if err != nil && !isRetryableError(err) {
			if c.logger != nil {
				c.logger.Error("Non-retryable error during batch operation",
					"error", err.Error(),
					"batch_size", len(batch.Entries))
			}
			return nil, err
		}
		if err != nil && c.logger != nil {
			c.logger.Warn("Retryable error during batch operation",
				"error", err.Error(),
				"batch_size", len(batch.Entries))
		}
		return nil, err
	}, &retry.RetryConfig{
		MaxRetries:      config.MaxRetries,
		InitialDelay:    config.InitialDelay,
		MaxDelay:        config.MaxDelay,
		BackoffFactor:   config.BackoffFactor,
		JitterFactor:    config.JitterFactor,
		LogRetryAttempt: config.LogRetryAttempt,
	}, c.logger)

	if err != nil && c.logger != nil {
		c.logger.Error("Batch operation failed after all retries",
			"error", err.Error(),
			"batch_size", len(batch.Entries),
			"max_retries", config.MaxRetries)
	}

	return err
}
