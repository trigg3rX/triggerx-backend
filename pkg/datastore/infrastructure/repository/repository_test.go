package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// MockScyllaQueryBuilder is a mock implementation of ScyllaQueryBuilder interface
type MockScyllaQueryBuilder[T any] struct {
	mock.Mock
}

// MockQuery is a mock implementation of Query interface for testing
type MockQuery struct {
	mock.Mock
}

func (m *MockQuery) WithContext(ctx context.Context) interface{} {
	args := m.Called(ctx)
	return args.Get(0)
}

func (m *MockQuery) Iter() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockQuery) Exec() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockScyllaQueryBuilder[T]) Insert(ctx context.Context, data *T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockScyllaQueryBuilder[T]) Update(ctx context.Context, data *T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockScyllaQueryBuilder[T]) Delete(ctx context.Context, data *T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockScyllaQueryBuilder[T]) Get(ctx context.Context, data *T) (*T, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

func (m *MockScyllaQueryBuilder[T]) Select(ctx context.Context, data *T) ([]T, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockScyllaQueryBuilder[T]) SelectAll(ctx context.Context) ([]T, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockScyllaQueryBuilder[T]) BatchInsert(ctx context.Context, data []*T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockScyllaQueryBuilder[T]) SetBatchSize(size int) {
	m.Called(size)
}

func (m *MockScyllaQueryBuilder[T]) SetConsistencyLevel(level string) {
	m.Called(level)
}

func (m *MockScyllaQueryBuilder[T]) Close() {
	m.Called()
}

func (m *MockScyllaQueryBuilder[T]) GetTableName() string {
	args := m.Called()
	return args.String(0)
}

// TestEntity represents a test entity for testing the repository
type TestEntity struct {
	ID   string `cql:"id"`
	Name string `cql:"name"`
	Age  int    `cql:"age"`
}

func TestNewGenericRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockLogger := &logging.NoOpLogger{}

	// Setup expectations
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)

	// Create repository
	repo := NewGenericRepository[TestEntity](
		mockConnection,
		mockLogger,
		"test_table",
		"id",
	)

	// Verify repository was created correctly
	assert.NotNil(t, repo)
	assert.Equal(t, "test_table", repo.GetTableName())
	assert.Equal(t, "id", repo.GetPrimaryKey())

	// Cleanup
	repo.Close()
}

func TestGenericRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("success", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := &TestEntity{
			ID:   "test-id",
			Name: "Test Name",
			Age:  25,
		}

		mockQueryBuilder.On("Insert", mock.Anything, testEntity).Return(nil)

		err := repo.Create(context.Background(), testEntity)
		assert.NoError(t, err)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := &TestEntity{
			ID:   "test-id",
			Name: "Test Name",
			Age:  25,
		}

		expectedError := errors.New("insert failed")
		mockQueryBuilder.On("Insert", mock.Anything, testEntity).Return(expectedError)

		err := repo.Create(context.Background(), testEntity)
		assert.Equal(t, expectedError, err)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("success", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := &TestEntity{
			ID:   "test-id",
			Name: "Updated Name",
			Age:  30,
		}

		mockQueryBuilder.On("Update", mock.Anything, testEntity).Return(nil)

		err := repo.Update(context.Background(), testEntity)
		assert.NoError(t, err)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := &TestEntity{
			ID:   "test-id",
			Name: "Updated Name",
			Age:  30,
		}

		expectedError := errors.New("update failed")
		mockQueryBuilder.On("Update", mock.Anything, testEntity).Return(expectedError)

		err := repo.Update(context.Background(), testEntity)
		assert.Equal(t, expectedError, err)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("success", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := TestEntity{
			ID:   "test-id",
			Name: "Test Name",
			Age:  25,
		}

		searchEntity := &TestEntity{ID: "test-id"}
		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(&testEntity, nil)

		result, err := repo.GetByID(context.Background(), "test-id")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testEntity.ID, result.ID)
		assert.Equal(t, testEntity.Name, result.Name)
		assert.Equal(t, testEntity.Age, result.Age)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("record not found", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		searchEntity := &TestEntity{ID: "non-existent"}
		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(nil, nil)

		result, err := repo.GetByID(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
		assert.Nil(t, result)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		expectedError := errors.New("database error")
		searchEntity := &TestEntity{ID: "test-id"}
		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(nil, expectedError)

		result, err := repo.GetByID(context.Background(), "test-id")
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_GetByNonID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("success", func(t *testing.T) {
		testEntity := TestEntity{
			ID:   "test-id",
			Name: "Test Name",
			Age:  25,
		}

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Test Name").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*TestEntity) = testEntity
			return nil
		})

		result, err := repo.GetByNonID(context.Background(), "name", "Test Name")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testEntity.ID, result.ID)
		assert.Equal(t, testEntity.Name, result.Name)
		assert.Equal(t, testEntity.Age, result.Age)
	})

	t.Run("record not found", func(t *testing.T) {
		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Non-existent").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).Return(errors.New("not found"))

		result, err := repo.GetByNonID(context.Background(), "name", "Non-existent")
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
		assert.Nil(t, result)
	})

	t.Run("database error", func(t *testing.T) {
		expectedError := errors.New("database connection failed")
		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Test Name").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).Return(expectedError)

		result, err := repo.GetByNonID(context.Background(), "name", "Test Name")
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
	})
}

func TestGenericRepository_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("success", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntities := []TestEntity{
			{ID: "id1", Name: "Entity 1", Age: 25},
			{ID: "id2", Name: "Entity 2", Age: 30},
		}

		mockQueryBuilder.On("SelectAll", mock.Anything).Return(testEntities, nil)

		result, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, testEntities[0].ID, result[0].ID)
		assert.Equal(t, testEntities[1].ID, result[1].ID)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		mockQueryBuilder.On("SelectAll", mock.Anything).Return([]TestEntity{}, nil)

		result, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		expectedError := errors.New("select all failed")
		mockQueryBuilder.On("SelectAll", mock.Anything).Return(nil, expectedError)

		result, err := repo.List(context.Background())
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_BatchCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("success", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntities := []*TestEntity{
			{ID: "id1", Name: "Entity 1", Age: 25},
			{ID: "id2", Name: "Entity 2", Age: 30},
		}

		mockQueryBuilder.On("BatchInsert", mock.Anything, testEntities).Return(nil)

		err := repo.BatchCreate(context.Background(), testEntities)
		assert.NoError(t, err)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("empty data", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		err := repo.BatchCreate(context.Background(), []*TestEntity{})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntities := []*TestEntity{
			{ID: "id1", Name: "Entity 1", Age: 25},
		}
		expectedError := errors.New("batch insert failed")

		mockQueryBuilder.On("BatchInsert", mock.Anything, testEntities).Return(expectedError)

		err := repo.BatchCreate(context.Background(), testEntities)
		assert.Equal(t, expectedError, err)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_GetByField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("success", func(t *testing.T) {
		testEntities := []TestEntity{
			{ID: "id1", Name: "Entity 1", Age: 25},
			{ID: "id2", Name: "Entity 1", Age: 30},
		}

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Entity 1").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().SelectRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*[]TestEntity) = testEntities
			return nil
		})

		result, err := repo.GetByField(context.Background(), "name", "Entity 1")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, testEntities[0].ID, result[0].ID)
		assert.Equal(t, testEntities[1].ID, result[1].ID)
	})

	t.Run("empty results", func(t *testing.T) {
		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Non-existent").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().SelectRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*[]TestEntity) = []TestEntity{}
			return nil
		})

		result, err := repo.GetByField(context.Background(), "name", "Non-existent")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("select failed")

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Test Name").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().SelectRelease(gomock.Any()).Return(expectedError)

		result, err := repo.GetByField(context.Background(), "name", "Test Name")
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
	})
}

func TestGenericRepository_GetByFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockLogger := &logging.NoOpLogger{}

	// Create a mock query builder that implements the correct interface
	mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: mockQueryBuilder,
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("success with conditions", func(t *testing.T) {
		conditions := map[string]interface{}{
			"name": "Test Name",
			"age":  25,
		}
		testEntities := []TestEntity{
			{ID: "id1", Name: "Test Name", Age: 25},
		}

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)

		// Use gomock.Any() for the query string since map iteration order is not guaranteed
		// The query should contain both fields but the order may vary
		mockGocqlxSession.EXPECT().Query(gomock.Any(), gomock.Any()).DoAndReturn(func(query string, names []string) interface{} {
			// Verify the query contains the expected elements
			assert.Contains(t, query, "SELECT * FROM test_table WHERE")
			assert.Contains(t, query, "name = ?")
			assert.Contains(t, query, "age = ?")
			assert.Contains(t, query, "AND")
			// Verify the names slice contains both fields
			assert.Len(t, names, 2)
			assert.Contains(t, names, "name")
			assert.Contains(t, names, "age")
			return mockQuery
		})

		// Use gomock.Any() for BindStruct since the order depends on map iteration
		mockQuery.EXPECT().BindStruct(gomock.Any()).DoAndReturn(func(values []interface{}) interface{} {
			// Verify the values slice contains the expected values
			assert.Len(t, values, 2)
			assert.Contains(t, values, "Test Name")
			assert.Contains(t, values, 25)
			return mockQuery
		})
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().SelectRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*[]TestEntity) = testEntities
			return nil
		})

		result, err := repo.GetByFields(context.Background(), conditions)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, testEntities[0].ID, result[0].ID)
	})

	t.Run("empty conditions - calls List", func(t *testing.T) {
		testEntities := []TestEntity{
			{ID: "id1", Name: "Entity 1", Age: 25},
			{ID: "id2", Name: "Entity 2", Age: 30},
		}

		mockQueryBuilder.On("SelectAll", mock.Anything).Return(testEntities, nil)

		result, err := repo.GetByFields(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, testEntities[0].ID, result[0].ID)
		assert.Equal(t, testEntities[1].ID, result[1].ID)
	})

	t.Run("error", func(t *testing.T) {
		conditions := map[string]interface{}{
			"name": "Test Name",
		}
		expectedError := errors.New("select failed")

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct([]interface{}{"Test Name"}).Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().SelectRelease(gomock.Any()).Return(expectedError)

		result, err := repo.GetByFields(context.Background(), conditions)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
	})
}

func TestGenericRepository_Count(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("success", func(t *testing.T) {
		expectedCount := int64(42)

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT COUNT(*) FROM test_table", []string{}).Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*int64) = expectedCount
			return nil
		})

		count, err := repo.Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("count failed")

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT COUNT(*) FROM test_table", []string{}).Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).Return(expectedError)

		count, err := repo.Count(context.Background())
		assert.Equal(t, expectedError, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestGenericRepository_Exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	t.Run("exists - true", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		testEntity := TestEntity{ID: "test-id", Name: "Test Name", Age: 25}
		searchEntity := &TestEntity{ID: "test-id"}

		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(&testEntity, nil)

		exists, err := repo.Exists(context.Background(), "test-id")
		assert.NoError(t, err)
		assert.True(t, exists)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("not exists - false", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		searchEntity := &TestEntity{ID: "non-existent"}

		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(nil, nil)

		exists, err := repo.Exists(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.False(t, exists)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		// Create a fresh mock query builder for each test
		mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

		repo := &genericRepository[TestEntity]{
			db:           mockConnection,
			queryBuilder: mockQueryBuilder,
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		expectedError := errors.New("database error")
		searchEntity := &TestEntity{ID: "test-id"}

		mockQueryBuilder.On("Get", mock.Anything, searchEntity).Return(nil, expectedError)

		exists, err := repo.Exists(context.Background(), "test-id")
		assert.Equal(t, expectedError, err)
		assert.False(t, exists)
		mockQueryBuilder.AssertExpectations(t)
	})
}

func TestGenericRepository_ExistsByField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockGocqlxSession := mocks.NewMockGocqlxSessioner(ctrl)
	mockQuery := mocks.NewMockGocqlxQueryer(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("exists - true", func(t *testing.T) {
		testEntity := TestEntity{ID: "test-id", Name: "Test Name", Age: 25}

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Test Name").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).DoAndReturn(func(dest interface{}) error {
			*dest.(*TestEntity) = testEntity
			return nil
		})

		exists, err := repo.ExistsByField(context.Background(), "name", "Test Name")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("not exists - false", func(t *testing.T) {
		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Non-existent").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).Return(errors.New("not found"))

		exists, err := repo.ExistsByField(context.Background(), "name", "Non-existent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error", func(t *testing.T) {
		expectedError := errors.New("database error")

		mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSession)
		mockGocqlxSession.EXPECT().Query("SELECT * FROM test_table WHERE name = ?", []string{"name"}).Return(mockQuery)
		mockQuery.EXPECT().BindStruct("Test Name").Return(mockQuery)
		mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().GetRelease(gomock.Any()).Return(expectedError)

		exists, err := repo.ExistsByField(context.Background(), "name", "Test Name")
		assert.Equal(t, expectedError, err)
		assert.False(t, exists)
	})
}

func TestGenericRepository_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	// Create a mock query builder that implements the correct interface
	mockQueryBuilder := &MockScyllaQueryBuilder[TestEntity]{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: mockQueryBuilder,
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	mockQueryBuilder.On("Close").Return()

	repo.Close()
	mockQueryBuilder.AssertExpectations(t)
}

func TestGenericRepository_GetTableName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	assert.Equal(t, "test_table", repo.GetTableName())
}

func TestGenericRepository_GetPrimaryKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	assert.Equal(t, "id", repo.GetPrimaryKey())
}

func TestGenericRepository_createSearchEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.NoOpLogger{}

	repo := &genericRepository[TestEntity]{
		db:           mockConnection,
		queryBuilder: &MockScyllaQueryBuilder[TestEntity]{},
		logger:       mockLogger,
		tableName:    "test_table",
		primaryKey:   "id",
	}

	t.Run("string ID", func(t *testing.T) {
		result := repo.createSearchEntity("test-id")
		assert.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("int ID", func(t *testing.T) {
		// Create a repository with int primary key for testing
		type TestEntityWithIntID struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}

		repoWithIntID := &genericRepository[TestEntityWithIntID]{
			db:           mockConnection,
			queryBuilder: &MockScyllaQueryBuilder[TestEntityWithIntID]{},
			logger:       mockLogger,
			tableName:    "test_table",
			primaryKey:   "id",
		}

		result := repoWithIntID.createSearchEntity(42)
		assert.NotNil(t, result)
		assert.Equal(t, 42, result.ID)
	})
}
