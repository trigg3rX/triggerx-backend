package database

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// Queryx is a wrapper around gocql.Query that provides retry logic via the generic retry package.
type Queryx struct {
	query  *gocql.Query
	conn   *Connection
	isIdem bool
}

// NewQuery wraps a gocql.Query to provide retry functionality.
func (c *Connection) NewQuery(stmt string, values ...interface{}) *Queryx {
	return &Queryx{
		query: c.session.Query(stmt, values...),
		conn:  c,
	}
}

// Exec executes a query with retry logic.
// The query should be marked as Idempotent() for safe retries on CUD operations.
func (q *Queryx) Exec() error {
	// if !q.isIdem {
	// q.conn.logger.Warnf("Executing a non-idempotent query with retry logic. Ensure this is intended.")
	// }

	operation := func() error {
		return q.query.Exec()
	}

	// Use default retry config if none is provided
	var cfg retry.RetryConfig
	if q.conn.config.RetryConfig != nil {
		cfg = *q.conn.config.RetryConfig
	} else {
		cfg = *retry.DefaultRetryConfig()
	}
	cfg.ShouldRetry = gocqlShouldRetry

	return retry.RetryFunc(q.query.Context(), operation, &cfg, q.conn.logger)
}

// Scan executes a query and scans the result, with retry logic.
func (q *Queryx) Scan(dest ...interface{}) error {
	operation := func() error {
		return q.query.Scan(dest...)
	}

	// Use default retry config if none is provided
	var cfg retry.RetryConfig
	if q.conn.config.RetryConfig != nil {
		cfg = *q.conn.config.RetryConfig
	} else {
		cfg = *retry.DefaultRetryConfig()
	}
	cfg.ShouldRetry = gocqlShouldRetry

	_, err := retry.Retry(q.query.Context(), func() (struct{}, error) {
		return struct{}{}, operation()
	}, &cfg, q.conn.logger)

	return err
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
