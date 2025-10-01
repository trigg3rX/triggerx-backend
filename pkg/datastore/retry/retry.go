package retry

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// Queryx is a wrapper around gocql.Query that provides retry logic via the generic retry package.
type Queryx struct {
	query       *gocql.Query
	retryConfig *retry.RetryConfig
	logger      logging.Logger
	isIdem      bool
}

// NewQuery wraps a gocql.Query to provide retry functionality.
func NewQuery(query *gocql.Query, retryConfig *retry.RetryConfig, logger logging.Logger) *Queryx {
	return &Queryx{
		query:       query,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// Exec executes a query with retry logic.
// The query should be marked as Idempotent() for safe retries on CUD operations.
func (q *Queryx) Exec() error {
	operation := func() error {
		return q.query.Exec()
	}

	return q.performWithRetry(operation)
}

// Scan executes a query and scans the result, with retry logic.
func (q *Queryx) Scan(dest ...interface{}) error {
	operation := func() error {
		return q.query.Scan(dest...)
	}

	return q.performWithRetry(operation)
}

// Iter returns an iterator for the query.
// Retries on iterators are handled internally by gocql's paging mechanism.
// No custom retry wrapper is needed here.
func (q *Queryx) Iter() *gocql.Iter {
	return q.query.Iter()
}

// WithContext sets the context for the underlying gocql.Query.
// Note: The retry mechanism itself also uses context for timeouts/cancellation.
func (q *Queryx) WithContext(ctx context.Context) *Queryx {
	q.query.WithContext(ctx)
	return q
}

// Idempotent marks the query as idempotent.
// This is critical for Exec() to be retried safely.
func (q *Queryx) Idempotent() *Queryx {
	q.query.Idempotent(true)
	q.isIdem = true
	return q
}

// performWithRetry executes an operation with retry logic.
func (q *Queryx) performWithRetry(op func() error) error {
	var cfg retry.RetryConfig
	if q.retryConfig != nil {
		cfg = *q.retryConfig
	} else {
		cfg = *retry.DefaultRetryConfig()
	}

	// Always override the ShouldRetry predicate with the gocql-specific one.
	cfg.ShouldRetry = func(err error, attempt int) bool {
		return gocqlShouldRetry(err)
	}

	// Use a wrapper that returns a value to work with the generic Retry function
	operation := func() (struct{}, error) {
		return struct{}{}, op()
	}

	_, err := retry.Retry(q.query.Context(), operation, &cfg, q.logger)
	return err
}
