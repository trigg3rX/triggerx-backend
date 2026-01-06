package metrics

import (
	"context"
	"strings"
	"time"

	"fmt"

	"github.com/gocql/gocql"
)

// TrackDBOperation is a helper function to track database operations
// It tracks operation duration, success/failure, errors, and slow queries
func TrackDBOperation(operation string, table string) func(error) {
	startTime := time.Now()
	ctx := context.Background() // Helper usually called in defer, so bg context is safer/easier unless passed
	return func(err error) {
		duration := time.Since(startTime).Seconds()
		status := "success"
		if err != nil {
			status = "error"
			TrackDBError(err)
		}

		// Record operation metrics
		if DatabaseOperationsTotal != nil {
			DatabaseOperationsTotal.WithLabelValues(operation, table, status).Inc(ctx)
		}
		if DatabaseOperationDuration != nil {
			DatabaseOperationDuration.WithLabelValues(operation, table).Record(ctx, duration)
		}

		// Track slow queries with multiple thresholds
		if duration > 0.1 { // 100ms
			if DBSlowQueriesTotal != nil {
				DBSlowQueriesTotal.WithLabelValues("100ms").Inc(ctx)
			}
		}
		if duration > 0.5 { // 500ms
			if DBSlowQueriesTotal != nil {
				DBSlowQueriesTotal.WithLabelValues("500ms").Inc(ctx)
			}
		}
		if duration > 1.0 { // 1 second
			if DBSlowQueriesTotal != nil {
				DBSlowQueriesTotal.WithLabelValues("1s").Inc(ctx)
			}
		}
		if duration > 5.0 { // 5 seconds - very slow
			if DBSlowQueriesTotal != nil {
				DBSlowQueriesTotal.WithLabelValues("5s").Inc(ctx)
			}
		}
	}
}

// TrackDBError is a helper function to track database errors
func TrackDBError(err error) {
	if err == nil {
		return
	}

	errorType := classifyDBError(err)

	if DatabaseErrorsTotal != nil {
		DatabaseErrorsTotal.WithLabelValues(errorType).Inc(context.Background())
	}
}

// classifyDBError classifies a database error into a category
func classifyDBError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case err == gocql.ErrTimeoutNoResponse:
		return "timeout"
	case err == gocql.ErrConnectionClosed:
		return "connection_closed"
	case err == gocql.ErrNoConnections:
		return "no_connections"
	case err == gocql.ErrNotFound:
		return "not_found"
	case err == gocql.ErrUnavailable:
		return "unavailable"
	case err == gocql.ErrTooManyTimeouts:
		return "too_many_timeouts"
	case err == gocql.ErrSessionClosed:
		return "session_closed"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "constraint"):
		return "constraint"
	case strings.Contains(errStr, "syntax"):
		return "syntax"
	case strings.Contains(errStr, "unauthorized"):
		return "unauthorized"
	case strings.Contains(errStr, "invalid"):
		return "invalid"
	case strings.Contains(errStr, "already exists"):
		return "already_exists"
	default:
		return "unknown"
	}
}

// TrackRetry tracks retry mechanism metrics
func TrackRetry(endpoint string, attempt int, success bool) {
	ctx := context.Background()
	if RetryAttemptsTotal != nil {
		RetryAttemptsTotal.WithLabelValues(endpoint, fmt.Sprintf("%d", attempt)).Inc(ctx)
	}
	if success {
		if RetrySuccessesTotal != nil {
			RetrySuccessesTotal.WithLabelValues(endpoint).Inc(ctx)
		}
	} else {
		if RetryFailuresTotal != nil {
			RetryFailuresTotal.WithLabelValues(endpoint).Inc(ctx)
		}
	}
}
