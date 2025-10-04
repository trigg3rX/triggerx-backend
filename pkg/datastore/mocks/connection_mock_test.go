package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMockConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := NewMockConnection(ctrl)

	t.Run("Close", func(t *testing.T) {
		mockConnection.EXPECT().Close()
		mockConnection.Close()
	})

	t.Run("GetGocqlxSession", func(t *testing.T) {
		expectedSession := NewMockGocqlxSessioner(ctrl)
		mockConnection.EXPECT().GetGocqlxSession().Return(expectedSession)

		result := mockConnection.GetGocqlxSession()
		assert.Equal(t, expectedSession, result)
	})

	t.Run("GetSession", func(t *testing.T) {
		expectedSession := NewMockSessioner(ctrl)
		mockConnection.EXPECT().GetSession().Return(expectedSession)

		result := mockConnection.GetSession()
		assert.Equal(t, expectedSession, result)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("health check failed")

		mockConnection.EXPECT().HealthCheck(ctx).Return(expectedError)

		result := mockConnection.HealthCheck(ctx)
		assert.Equal(t, expectedError, result)
	})

	t.Run("HealthCheck_Success", func(t *testing.T) {
		ctx := context.Background()

		mockConnection.EXPECT().HealthCheck(ctx).Return(nil)

		result := mockConnection.HealthCheck(ctx)
		assert.NoError(t, result)
	})
}

func TestMockSessioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessioner := NewMockSessioner(ctrl)

	t.Run("Close", func(t *testing.T) {
		mockSessioner.EXPECT().Close()
		mockSessioner.Close()
	})

	t.Run("Query", func(t *testing.T) {
		stmt := "SELECT * FROM test"
		values := []interface{}{"value1", "value2"}
		expectedQuery := &gocql.Query{}

		mockSessioner.EXPECT().Query(stmt, values[0], values[1]).Return(expectedQuery)

		result := mockSessioner.Query(stmt, values...)
		assert.Equal(t, expectedQuery, result)
	})

	t.Run("Query_NoValues", func(t *testing.T) {
		stmt := "SELECT * FROM test"
		expectedQuery := &gocql.Query{}

		mockSessioner.EXPECT().Query(stmt).Return(expectedQuery)

		result := mockSessioner.Query(stmt)
		assert.Equal(t, expectedQuery, result)
	})
}

func TestMockQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuery := NewMockQuery(ctrl)

	t.Run("BindStruct", func(t *testing.T) {
		data := struct{ ID int }{ID: 1}
		expectedQuery := NewMockQuery(ctrl)

		mockQuery.EXPECT().BindStruct(data).Return(expectedQuery)

		result := mockQuery.BindStruct(data)
		assert.Equal(t, expectedQuery, result)
	})

	t.Run("Exec", func(t *testing.T) {
		expectedError := errors.New("exec failed")

		mockQuery.EXPECT().Exec().Return(expectedError)

		result := mockQuery.Exec()
		assert.Equal(t, expectedError, result)
	})

	t.Run("Exec_Success", func(t *testing.T) {
		mockQuery.EXPECT().Exec().Return(nil)

		result := mockQuery.Exec()
		assert.NoError(t, result)
	})

	t.Run("Iter", func(t *testing.T) {
		expectedIter := NewMockIter(ctrl)

		mockQuery.EXPECT().Iter().Return(expectedIter)

		result := mockQuery.Iter()
		assert.Equal(t, expectedIter, result)
	})

	t.Run("Scan", func(t *testing.T) {
		dest := []interface{}{new(string), new(int)}
		expectedError := errors.New("scan failed")

		mockQuery.EXPECT().Scan(dest[0], dest[1]).Return(expectedError)

		result := mockQuery.Scan(dest...)
		assert.Equal(t, expectedError, result)
	})

	t.Run("Scan_Success", func(t *testing.T) {
		dest := []interface{}{new(string), new(int)}

		mockQuery.EXPECT().Scan(dest[0], dest[1]).Return(nil)

		result := mockQuery.Scan(dest...)
		assert.NoError(t, result)
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		expectedQuery := NewMockQuery(ctrl)

		mockQuery.EXPECT().WithContext(ctx).Return(expectedQuery)

		result := mockQuery.WithContext(ctx)
		assert.Equal(t, expectedQuery, result)
	})
}

func TestMockIter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIter := NewMockIter(ctrl)

	t.Run("Close", func(t *testing.T) {
		expectedError := errors.New("close failed")

		mockIter.EXPECT().Close().Return(expectedError)

		result := mockIter.Close()
		assert.Equal(t, expectedError, result)
	})

	t.Run("Close_Success", func(t *testing.T) {
		mockIter.EXPECT().Close().Return(nil)

		result := mockIter.Close()
		assert.NoError(t, result)
	})

	t.Run("Scan", func(t *testing.T) {
		dest := []interface{}{new(string), new(int)}
		expectedResult := true

		mockIter.EXPECT().Scan(dest[0], dest[1]).Return(expectedResult)

		result := mockIter.Scan(dest...)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Scan_NoMoreRows", func(t *testing.T) {
		dest := []interface{}{new(string)}
		expectedResult := false

		mockIter.EXPECT().Scan(dest[0]).Return(expectedResult)

		result := mockIter.Scan(dest...)
		assert.Equal(t, expectedResult, result)
	})
}

func TestMockGocqlxSessioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGocqlxSessioner := NewMockGocqlxSessioner(ctrl)

	t.Run("Close", func(t *testing.T) {
		mockGocqlxSessioner.EXPECT().Close()
		mockGocqlxSessioner.Close()
	})

	t.Run("Query", func(t *testing.T) {
		stmt := "SELECT * FROM test WHERE id = ?"
		names := []string{"id"}
		expectedQueryer := NewMockGocqlxQueryer(ctrl)

		mockGocqlxSessioner.EXPECT().Query(stmt, names).Return(expectedQueryer)

		result := mockGocqlxSessioner.Query(stmt, names)
		assert.Equal(t, expectedQueryer, result)
	})
}

// TestMockConnectionIntegration tests the interaction between different mock types
func TestMockConnectionIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := NewMockConnection(ctrl)
	mockSessioner := NewMockSessioner(ctrl)
	mockQuery := &gocql.Query{}

	t.Run("GetSessionAndQuery", func(t *testing.T) {
		stmt := "SELECT * FROM test"

		// Setup expectations
		mockConnection.EXPECT().GetSession().Return(mockSessioner)
		mockSessioner.EXPECT().Query(stmt).Return(mockQuery)

		// Execute
		session := mockConnection.GetSession()
		query := session.Query(stmt)

		// Verify
		assert.Equal(t, mockSessioner, session)
		assert.Equal(t, mockQuery, query)
	})
}

// TestMockQueryChaining tests method chaining on MockQuery
func TestMockQueryChaining(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuery := NewMockQuery(ctrl)
	mockQueryWithContext := NewMockQuery(ctrl)
	mockQueryWithBind := NewMockQuery(ctrl)

	t.Run("WithContextAndBindStruct", func(t *testing.T) {
		ctx := context.Background()
		data := struct{ ID int }{ID: 1}

		// Setup expectations for chaining
		mockQuery.EXPECT().WithContext(ctx).Return(mockQueryWithContext)
		mockQueryWithContext.EXPECT().BindStruct(data).Return(mockQueryWithBind)
		mockQueryWithBind.EXPECT().Exec().Return(nil)

		// Execute chained calls
		result := mockQuery.WithContext(ctx).BindStruct(data)
		err := result.Exec()

		// Verify
		assert.Equal(t, mockQueryWithBind, result)
		assert.NoError(t, err)
	})
}

// TestMockIterScanLoop tests scanning multiple rows
func TestMockIterScanLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIter := NewMockIter(ctrl)

	t.Run("ScanMultipleRows", func(t *testing.T) {
		dest := []interface{}{new(string), new(int)}

		// Setup expectations for multiple scans
		mockIter.EXPECT().Scan(dest[0], dest[1]).Return(true).Times(2)
		mockIter.EXPECT().Close().Return(nil)

		// Simulate scanning loop
		rowsScanned := 0

		// First scan
		if mockIter.Scan(dest...) {
			rowsScanned++
		}
		// Second scan
		if mockIter.Scan(dest...) {
			rowsScanned++
		}

		err := mockIter.Close()

		// Verify
		assert.Equal(t, 2, rowsScanned)
		assert.NoError(t, err)
	})
}

// Benchmark tests for performance validation
func BenchmarkMockConnectionGetSession(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockConnection := NewMockConnection(ctrl)
	mockSessioner := NewMockSessioner(ctrl)

	mockConnection.EXPECT().GetSession().Return(mockSessioner).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockConnection.GetSession()
	}
}

func BenchmarkMockQueryExec(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockQuery := NewMockQuery(ctrl)

	mockQuery.EXPECT().Exec().Return(nil).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockQuery.Exec()
	}
}
