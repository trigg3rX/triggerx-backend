package metrics

import (
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// TrackDBOperation is a helper function to track database operations
// It tracks operation duration, success/failure, errors, and slow queries
func TrackDBOperation(operation string, table string) func(error) {
	startTime := time.Now()
	return func(err error) {
		duration := time.Since(startTime).Seconds()
		status := "success"
		if err != nil {
			status = "error"
			TrackDBError(err)
		}

		// Record operation metrics
		DatabaseOperationsTotal.WithLabelValues(operation, table, status).Inc()
		DatabaseOperationDuration.WithLabelValues(operation, table).Observe(duration)

		// Track slow queries
		if duration > 1.0 { // Consider queries taking more than 1 second as slow
			DBSlowQueriesTotal.WithLabelValues("1s").Inc()
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

	DatabaseErrorsTotal.WithLabelValues(errorType).Inc()
}

// TrackRetry tracks retry mechanism metrics
func TrackRetry(endpoint string, attempt int, success bool) {
	RetryAttemptsTotal.WithLabelValues(endpoint, string(attempt)).Inc()
	if success {
		RetrySuccessesTotal.WithLabelValues(endpoint).Inc()
	} else {
		RetryFailuresTotal.WithLabelValues(endpoint).Inc()
	}
}
