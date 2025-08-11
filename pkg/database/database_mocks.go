package database

import (
	"context"
	"reflect"

	"github.com/gocql/gocql"
)

// --- Mock Query ---

// MockQuery implements the methods of gocql.Query that our Queryx wrapper uses.
type MockQuery struct {
	ctx          context.Context
	scanArgs     []interface{} // Data to populate Scan's destination args with
	execErr      error         // Error to return on Exec()
	scanErr      error         // Error to return on Scan()
	callCount    int           // How many times Exec() or Scan() was called
	maxCalls     int           // How many times to fail before succeeding
	isIdempotent bool
}

func (m *MockQuery) Exec() error {
	m.callCount++
	if m.callCount > m.maxCalls {
		return nil // Success after N failures
	}
	return m.execErr
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	m.callCount++
	if m.callCount > m.maxCalls {
		// On success, populate the dest args
		for i, arg := range m.scanArgs {
			if i < len(dest) {
				reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(arg))
			}
		}
		return nil
	}
	return m.scanErr
}

func (m *MockQuery) Iter() *gocql.Iter {
	// Not testing iterators in this example, but you could create a mock iterator too
	return &gocql.Iter{}
}

func (m *MockQuery) WithContext(ctx context.Context) *gocql.Query {
	m.ctx = ctx
	// The return value needs to be a *gocql.Query, so we return a dummy one.
	// Our mock methods will be called through the MockSession's setup.
	return &gocql.Query{}
}

func (m *MockQuery) Idempotent(idem bool) *gocql.Query {
	m.isIdempotent = idem
	return &gocql.Query{}
}

// Context returns the query's context
func (m *MockQuery) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

// --- Mock Session ---

// MockSession implements the database.Sessioner interface for testing.
type MockSession struct {
	// The Query method will return this query, allowing us to control its behavior.
	mockQuery *MockQuery
}

// Query returns our controllable MockQuery.
func (m *MockSession) Query(stmt string, values ...interface{}) *gocql.Query {
	// Create a minimal gocql.Query that won't cause nil pointer issues
	// We'll use a dummy query that has the minimum required structure
	dummyQuery := &gocql.Query{}

	// Store our mock query for later access
	m.mockQuery = &MockQuery{}

	// Return the dummy query - the actual mocking will happen in the test setup
	return dummyQuery
}

func (m *MockSession) Close() {
	// Clean up if needed
}

// GetMockQuery returns the mock query for assertions
func (m *MockSession) GetMockQuery() *MockQuery {
	return m.mockQuery
}
