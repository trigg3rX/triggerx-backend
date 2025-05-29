package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// QueryExecutor defines the interface for executing queries
type QueryExecutor interface {
	Exec() error
	Iter() *gocql.Iter
	Scan(dest ...interface{}) error
}

// MockQuery is a mock implementation of QueryExecutor
type MockQuery struct {
	execFunc func() error
}

func (m *MockQuery) Exec() error {
	return m.execFunc()
}

func (m *MockQuery) Iter() *gocql.Iter {
	return &gocql.Iter{}
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	return nil
}

// MockSession is a mock implementation of gocql.Session
type MockSession struct {
	execFunc  func(query string, values ...interface{}) error
	scanFunc  func(query string, dest ...interface{}) error
	batchFunc func(batch *gocql.Batch) error
}

func (m *MockSession) Query(query string, values ...interface{}) *gocql.Query {
	// Create a mock query that will be type asserted in the test
	return &gocql.Query{}
}

func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
	return m.batchFunc(batch)
}

func (m *MockSession) Close() {}

func (m *MockSession) Closed() bool {
	return false
}

func (m *MockSession) ExecuteBatchCAS(batch *gocql.Batch, dest ...interface{}) (bool, error) {
	return false, nil
}

func (m *MockSession) ExecuteBatchCASWithContext(ctx context.Context, batch *gocql.Batch, dest ...interface{}) (bool, error) {
	return false, nil
}

func (m *MockSession) ExecuteBatchWithContext(ctx context.Context, batch *gocql.Batch) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithContextAndTimeout(ctx context.Context, batch *gocql.Batch, timeout time.Duration) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithTimeout(batch *gocql.Batch, timeout time.Duration) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithContextAndTimeoutAndConsistency(ctx context.Context, batch *gocql.Batch, timeout time.Duration, consistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithTimeoutAndConsistency(batch *gocql.Batch, timeout time.Duration, consistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithContextAndConsistency(ctx context.Context, batch *gocql.Batch, consistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithConsistency(batch *gocql.Batch, consistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithContextAndTimeoutAndConsistencyAndSerialConsistency(ctx context.Context, batch *gocql.Batch, timeout time.Duration, consistency gocql.Consistency, serialConsistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithTimeoutAndConsistencyAndSerialConsistency(batch *gocql.Batch, timeout time.Duration, consistency gocql.Consistency, serialConsistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithContextAndConsistencyAndSerialConsistency(ctx context.Context, batch *gocql.Batch, consistency gocql.Consistency, serialConsistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func (m *MockSession) ExecuteBatchWithConsistencyAndSerialConsistency(batch *gocql.Batch, consistency gocql.Consistency, serialConsistency gocql.Consistency) error {
	return m.batchFunc(batch)
}

func init() {
	// Initialize logger for tests
	config := logging.NewDefaultConfig("retryable_test")
	config.Environment = logging.Development
	config.UseColors = true
	if err := logging.InitServiceLogger(config); err != nil {
		panic(err)
	}
}

func TestRetryableExec(t *testing.T) {
	tests := []struct {
		name        string
		execFunc    func() error
		expectError bool
	}{
		{
			name: "success on first try",
			execFunc: func() error {
				return errors.New("mock execution")
			},
			expectError: false,
		},
		{
			name: "retryable error",
			execFunc: func() error {
				return &gocql.RequestErrUnavailable{}
			},
			expectError: true,
		},
		{
			name: "non-retryable error",
			execFunc: func() error {
				return errors.New("non-retryable error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &MockSession{
				execFunc: func(query string, values ...interface{}) error {
					return tt.execFunc()
				},
			}

			conn := &Connection{
				session: mockSession,
			}

			err := conn.RetryableExec("SELECT * FROM test", "value1", "value2")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetryableScan(t *testing.T) {
	tests := []struct {
		name        string
		scanFunc    func(dest ...interface{}) error
		expectError bool
	}{
		{
			name: "success on first try",
			scanFunc: func(dest ...interface{}) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "retryable error",
			scanFunc: func(dest ...interface{}) error {
				return &gocql.RequestErrUnavailable{}
			},
			expectError: true,
		},
		{
			name: "non-retryable error",
			scanFunc: func(dest ...interface{}) error {
				return errors.New("non-retryable error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &MockSession{
				scanFunc: func(query string, dest ...interface{}) error {
					return tt.scanFunc(dest...)
				},
			}

			conn := &Connection{
				session: mockSession,
			}

			var result string
			err := conn.RetryableScan("SELECT * FROM test", &result)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetryableBatch(t *testing.T) {
	tests := []struct {
		name        string
		batchFunc   func(batch *gocql.Batch) error
		expectError bool
	}{
		{
			name: "success on first try",
			batchFunc: func(batch *gocql.Batch) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "retryable error",
			batchFunc: func(batch *gocql.Batch) error {
				return &gocql.RequestErrUnavailable{}
			},
			expectError: true,
		},
		{
			name: "non-retryable error",
			batchFunc: func(batch *gocql.Batch) error {
				return errors.New("non-retryable error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &MockSession{
				batchFunc: tt.batchFunc,
			}

			conn := &Connection{
				session: mockSession,
			}

			batch := &gocql.Batch{}
			err := conn.RetryableBatch(batch)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
