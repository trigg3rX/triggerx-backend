package query_builder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestUser represents a test entity for query builder tests
type TestUser struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

// TestGocqlxQueryBuilder_NewGocqlxQueryBuilder tests the constructor
func TestGocqlxQueryBuilder_NewGocqlxQueryBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	assert.NotNil(t, qb)
	assert.Equal(t, tableName, qb.GetTableName())
}

// TestGocqlxQueryBuilder_Insert tests the Insert operation
func TestGocqlxQueryBuilder_Insert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	// Test data
	user := &TestUser{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Mock the session.Query method to return a mock query
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(nil)

	// Test successful insert
	err := qb.Insert(ctx, user)
	assert.NoError(t, err)
}

// TestGocqlxQueryBuilder_Insert_Error tests the Insert operation with error
func TestGocqlxQueryBuilder_Insert_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	expectedError := errors.New("database connection failed")

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution to return error
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(expectedError)

	// Test insert with error
	err := qb.Insert(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx insert failed")
}

// TestGocqlxQueryBuilder_Update tests the Update operation
func TestGocqlxQueryBuilder_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID:        "user-123",
		Name:      "John Doe Updated",
		Email:     "john.updated@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(nil)

	// Test successful update
	err := qb.Update(ctx, user)
	assert.NoError(t, err)
}

// TestGocqlxQueryBuilder_Delete tests the Delete operation
func TestGocqlxQueryBuilder_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID: "user-123",
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(nil)

	// Test successful delete
	err := qb.Delete(ctx, user)
	assert.NoError(t, err)
}

// TestGocqlxQueryBuilder_Get tests the Get operation
func TestGocqlxQueryBuilder_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID: "user-123",
	}

	expectedUser := &TestUser{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().GetRelease(gomock.Any()).DoAndReturn(func(result interface{}) error {
		// Simulate setting the result
		*result.(*TestUser) = *expectedUser
		return nil
	})

	// Test successful get
	result, err := qb.Get(ctx, user)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.Email, result.Email)
}

// TestGocqlxQueryBuilder_Get_NotFound tests the Get operation when record not found
func TestGocqlxQueryBuilder_Get_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID: "user-123",
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution to return "not found" error
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().GetRelease(gomock.Any()).Return(errors.New("not found"))

	// Test get when not found
	result, err := qb.Get(ctx, user)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestGocqlxQueryBuilder_Select tests the Select operation
func TestGocqlxQueryBuilder_Select(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		Name: "John",
	}

	expectedUsers := []TestUser{
		{
			ID:        "user-1",
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			Name:      "John Smith",
			Email:     "john.smith@example.com",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().Select(gomock.Any()).DoAndReturn(func(results interface{}) error {
		// Simulate setting the results
		*results.(*[]TestUser) = expectedUsers
		return nil
	})

	// Test successful select
	results, err := qb.Select(ctx, user)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, expectedUsers[0].ID, results[0].ID)
	assert.Equal(t, expectedUsers[1].ID, results[1].ID)
}

// TestGocqlxQueryBuilder_SelectAll tests the SelectAll operation
func TestGocqlxQueryBuilder_SelectAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	expectedUsers := []TestUser{
		{
			ID:        "user-1",
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().SelectRelease(gomock.Any()).DoAndReturn(func(results interface{}) error {
		// Simulate setting the results
		*results.(*[]TestUser) = expectedUsers
		return nil
	})

	// Test successful select all
	results, err := qb.SelectAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, expectedUsers[0].ID, results[0].ID)
	assert.Equal(t, expectedUsers[1].ID, results[1].ID)
}

// TestGocqlxQueryBuilder_BatchInsert tests the BatchInsert operation
func TestGocqlxQueryBuilder_BatchInsert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	users := []*TestUser{
		{
			ID:        "user-1",
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()

	// Mock the session.Query method
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	// Mock the query execution
	mockQuery.EXPECT().BindStruct(gomock.Any()).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(nil)

	// Test successful batch insert
	err := qb.BatchInsert(ctx, users)
	assert.NoError(t, err)
}

// TestGocqlxQueryBuilder_BatchInsert_EmptySlice tests the BatchInsert operation with empty slice
func TestGocqlxQueryBuilder_BatchInsert_EmptySlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	ctx := context.Background()

	// Test batch insert with empty slice - should return immediately
	err := qb.BatchInsert(ctx, []*TestUser{})
	assert.NoError(t, err)
}

// TestGocqlxQueryBuilder_SetBatchSize tests the SetBatchSize method
func TestGocqlxQueryBuilder_SetBatchSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	// Test setting batch size
	qb.SetBatchSize(500)
	// Note: We can't directly test the internal batchSize field, but we can verify the method doesn't panic
	assert.NotNil(t, qb)
}

// TestGocqlxQueryBuilder_SetConsistencyLevel tests the SetConsistencyLevel method
func TestGocqlxQueryBuilder_SetConsistencyLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	// Test setting consistency level
	qb.SetConsistencyLevel("ALL")
	// Note: We can't directly test the internal consistencyLevel field, but we can verify the method doesn't panic
	assert.NotNil(t, qb)
}

// TestGocqlxQueryBuilder_Close tests the Close method
func TestGocqlxQueryBuilder_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	// Test close method - should not panic
	qb.Close()
	assert.NotNil(t, qb)
}

// TestGocqlxQueryBuilder_GetTableName tests the GetTableName method
func TestGocqlxQueryBuilder_GetTableName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	// Test getting table name
	result := qb.GetTableName()
	assert.Equal(t, tableName, result)
}

// TestGocqlxQueryBuilder_PreparedStatementCaching tests that prepared statements are cached
func TestGocqlxQueryBuilder_PreparedStatementCaching(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Mock the session.Query method - should be called only once due to caching
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery).Times(1)

	// Mock the query execution - should be called twice (for two operations)
	mockQuery.EXPECT().BindStruct(user).Return(mockQuery).Times(2)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery).Times(2)
	mockQuery.EXPECT().ExecRelease().Return(nil).Times(2)

	// Perform two insert operations - second should use cached prepared statement
	err1 := qb.Insert(ctx, user)
	assert.NoError(t, err1)

	err2 := qb.Insert(ctx, user)
	assert.NoError(t, err2)
}

// TestGocqlxQueryBuilder_ErrorHandling tests error handling across different operations
func TestGocqlxQueryBuilder_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := logging.NewNoOpLogger()
	tableName := "test_users"

	qb := NewGocqlxQueryBuilder[TestUser](mockSession, mockLogger, tableName)

	user := &TestUser{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	expectedError := errors.New("database connection lost")

	// Test error handling for Update
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery)

	mockQuery.EXPECT().BindStruct(user).Return(mockQuery)
	mockQuery.EXPECT().WithContext(ctx).Return(mockQuery)
	mockQuery.EXPECT().ExecRelease().Return(expectedError)

	err := qb.Update(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx update failed")

	// Test error handling for Delete
	mockQuery2 := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery2)

	mockQuery2.EXPECT().BindStruct(user).Return(mockQuery2)
	mockQuery2.EXPECT().WithContext(ctx).Return(mockQuery2)
	mockQuery2.EXPECT().ExecRelease().Return(expectedError)

	err = qb.Delete(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx delete failed")

	// Test error handling for Select
	mockQuery3 := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery3)

	mockQuery3.EXPECT().BindStruct(user).Return(mockQuery3)
	mockQuery3.EXPECT().WithContext(ctx).Return(mockQuery3)
	mockQuery3.EXPECT().Select(gomock.Any()).Return(expectedError)

	_, err = qb.Select(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx select failed")

	// Test error handling for SelectAll
	mockQuery4 := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery4)

	mockQuery4.EXPECT().WithContext(ctx).Return(mockQuery4)
	mockQuery4.EXPECT().SelectRelease(gomock.Any()).Return(expectedError)

	_, err = qb.SelectAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx select all failed")

	// Test error handling for BatchInsert
	users := []*TestUser{user}
	mockQuery5 := mocks.NewMockGocqlxQueryer(ctrl)
	mockSession.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockQuery5)

	mockQuery5.EXPECT().BindStruct(gomock.Any()).Return(mockQuery5)
	mockQuery5.EXPECT().WithContext(ctx).Return(mockQuery5)
	mockQuery5.EXPECT().ExecRelease().Return(expectedError)

	err = qb.BatchInsert(ctx, users)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gocqlx batch insert failed")
}
