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

		// Track slow queries
		if duration > 1.0 { // Consider queries taking more than 1 second as slow
			if DBSlowQueriesTotal != nil {
				DBSlowQueriesTotal.WithLabelValues("1s").Inc(ctx)
			}
		}
	}
}

// TrackDBError is a helper function to track database errors
func TrackDBError(err error) {
	if err == nil {
		return
	}

	errorType := "unknown"
	switch {
	case err == gocql.ErrTimeoutNoResponse:
		errorType = "timeout"
	case err == gocql.ErrConnectionClosed:
		errorType = "connection"
	case strings.Contains(err.Error(), "query"):
		errorType = "query"
	case strings.Contains(err.Error(), "constraint"):
		errorType = "constraint"
	}

	if DatabaseErrorsTotal != nil {
		DatabaseErrorsTotal.WithLabelValues(errorType).Inc(context.Background())
	}
}

// TrackRetry tracks retry mechanism metrics
func TrackRetry(endpoint string, attempt int, success bool) {
	ctx := context.Background()
	if RetryAttemptsTotal != nil {
		RetryAttemptsTotal.WithLabelValues(endpoint, fmt.Sprint(rune(attempt))).Inc(ctx)
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
