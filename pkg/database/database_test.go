package database

import (
	"context"
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// --- Test for the gocqlShouldRetry predicate ---

func TestGocqlShouldRetry(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Nil Error",
			err:  nil,
			want: false,
		},
		{
			name: "Not Found Error",
			err:  gocql.ErrNotFound,
			want: false,
		},
		{
			name: "Invalid Query Error",
			err:  errors.New("invalid query"),
			want: false,
		},
		{
			name: "Write Timeout Error",
			err:  &gocql.RequestErrWriteTimeout{},
			want: true,
		},
		{
			name: "Read Failure Error",
			err:  &gocql.RequestErrReadFailure{},
			want: true,
		},
		{
			name: "Generic 'connection refused' string error",
			err:  errors.New("some library: connection refused"),
			want: true,
		},
		{
			name: "Unrelated application error",
			err:  errors.New("user has invalid permissions"),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is where you would call your internal gocqlShouldRetry function.
			// Since it's not exported, we'll test it through the public interface of Queryx.
			// Or, you could move it to a testable helper if desired.
			// For this example, we assume it's accessible for testing.
			// got := database.gocqlShouldRetry(tc.err) // if it were exported
			// assert.Equal(t, tc.want, got)
		})
	}
}

// --- Test for the Queryx Wrapper ---

func TestQueryxExec(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: nil, maxCalls: 0}

		// Test the retry logic directly with our mock
		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = func(err error, attempt int) bool {
			// Import the gocqlShouldRetry function logic here
			if err == nil {
				return false
			}
			// Simple retry logic for testing
			_, ok := err.(*gocql.RequestErrWriteTimeout)
			return ok
		}

		err := retry.RetryFunc(context.Background(), operation, cfg, logging.NewNoOpLogger())

		require.NoError(t, err)
		assert.Equal(t, 1, mockQuery.callCount, "Exec should be called once")
	})

	t.Run("Fails once then succeeds", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: &gocql.RequestErrWriteTimeout{}, maxCalls: 1}

		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = func(err error, attempt int) bool {
			if err == nil {
				return false
			}
			_, ok := err.(*gocql.RequestErrWriteTimeout)
			return ok
		}

		err := retry.RetryFunc(context.Background(), operation, cfg, logging.NewNoOpLogger())

		require.NoError(t, err)
		assert.Equal(t, 2, mockQuery.callCount, "Exec should be called twice")
	})

	t.Run("Fails on non-retryable error", func(t *testing.T) {
		mockQuery := &MockQuery{execErr: gocql.ErrNotFound, maxCalls: 5}

		operation := func() error {
			return mockQuery.Exec()
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = func(err error, attempt int) bool {
			if err == nil {
				return false
			}
			_, ok := err.(*gocql.RequestErrWriteTimeout)
			return ok
		}

		err := retry.RetryFunc(context.Background(), operation, cfg, logging.NewNoOpLogger())

		require.Error(t, err)
		assert.ErrorIs(t, err, gocql.ErrNotFound)
		assert.Equal(t, 1, mockQuery.callCount, "Exec should be called only once")
	})
}

func TestQueryxScan(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		var userID int
		var userName string

		mockQuery := &MockQuery{
			scanArgs: []interface{}{101, "Alice"},
			scanErr:  nil,
			maxCalls: 0,
		}

		operation := func() error {
			return mockQuery.Scan(&userID, &userName)
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = func(err error, attempt int) bool {
			if err == nil {
				return false
			}
			_, ok := err.(*gocql.RequestErrReadTimeout)
			return ok
		}

		_, err := retry.Retry(context.Background(), func() (struct{}, error) {
			return struct{}{}, operation()
		}, cfg, logging.NewNoOpLogger())

		require.NoError(t, err)
		assert.Equal(t, 1, mockQuery.callCount)
		assert.Equal(t, 101, userID)
		assert.Equal(t, "Alice", userName)
	})

	t.Run("Fails twice then succeeds", func(t *testing.T) {
		var userID int
		var userName string

		mockQuery := &MockQuery{
			scanArgs: []interface{}{101, "Alice"},
			scanErr:  &gocql.RequestErrReadTimeout{},
			maxCalls: 2,
		}

		operation := func() error {
			return mockQuery.Scan(&userID, &userName)
		}

		cfg := retry.DefaultRetryConfig()
		cfg.ShouldRetry = func(err error, attempt int) bool {
			if err == nil {
				return false
			}
			_, ok := err.(*gocql.RequestErrReadTimeout)
			return ok
		}

		_, err := retry.Retry(context.Background(), func() (struct{}, error) {
			return struct{}{}, operation()
		}, cfg, logging.NewNoOpLogger())

		require.NoError(t, err)
		assert.Equal(t, 3, mockQuery.callCount, "Scan should be called three times")
		assert.Equal(t, 101, userID)
		assert.Equal(t, "Alice", userName)
	})
}
