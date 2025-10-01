package retry

import (
	"errors"
	"net"
	"strings"

	"github.com/gocql/gocql"
)

// gocqlShouldRetry is the predicate that determines if a gocql error is transient and should be retried.
func gocqlShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Do NOT retry if the resource is simply not found. This is a final state.
	if errors.Is(err, gocql.ErrNotFound) {
		return false
	}

	// Do NOT retry on query validation errors (e.g., syntax error).
	// Check for invalid query error codes (8704 = Invalid query, 2000 = Syntax error)
	var reqErr gocql.RequestError
	if errors.As(err, &reqErr) {
		code := reqErr.Code()
		// Error codes for invalid queries: 8704 (Invalid query), 2000 (Syntax error)
		if code == 8704 || code == 2000 {
			return false
		}
	}

	// Use type assertions for specific, known-retryable gocql error types.
	// This is the most reliable method.
	switch err.(type) {
	case *gocql.RequestErrWriteTimeout,
		*gocql.RequestErrReadTimeout,
		*gocql.RequestErrUnavailable,
		*gocql.RequestErrReadFailure,
		*gocql.RequestErrWriteFailure:
		return true
	}

	// Check for generic network errors, which are often retryable.
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// As a fallback, check for common error strings. This is more brittle but can
	// catch transient errors that don't conform to the types above.
	errMsg := strings.ToLower(err.Error())
	retryableMessages := []string{
		"no connections available",
		"connection reset by peer",
		"i/o timeout",
		"broken pipe",
		"connection timed out",
		"connection refused",
	}

	for _, msg := range retryableMessages {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}

	return false
}
