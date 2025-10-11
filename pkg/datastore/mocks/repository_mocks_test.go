package mocks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestMockGenericRepository(t *testing.T) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	t.Run("Create", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 1, Name: "test"}
		expectedError := errors.New("create failed")

		mockRepo.On("Create", ctx, data).Return(expectedError)

		result := mockRepo.Create(ctx, data)
		assert.Equal(t, expectedError, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 1, Name: "test"}

		mockRepo.On("Create", ctx, data).Return(nil)

		result := mockRepo.Create(ctx, data)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 1, Name: "updated"}
		expectedError := errors.New("update failed")

		mockRepo.On("Update", ctx, data).Return(expectedError)

		result := mockRepo.Update(ctx, data)
		assert.Equal(t, expectedError, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 1, Name: "updated"}

		mockRepo.On("Update", ctx, data).Return(nil)

		result := mockRepo.Update(ctx, data)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		id := 1
		expectedEntity := &TestEntity{ID: 1, Name: "test"}
		expectedError := errors.New("get by id failed")

		mockRepo.On("GetByID", ctx, id).Return(expectedEntity, expectedError)

		result, err := mockRepo.GetByID(ctx, id)
		assert.Equal(t, expectedEntity, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		id := 1
		expectedEntity := &TestEntity{ID: 1, Name: "test"}

		mockRepo.On("GetByID", ctx, id).Return(expectedEntity, nil)

		result, err := mockRepo.GetByID(ctx, id)
		assert.Equal(t, expectedEntity, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		id := 999
		expectedError := errors.New("not found")

		mockRepo.On("GetByID", ctx, id).Return(nil, expectedError)

		result, err := mockRepo.GetByID(ctx, id)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByNonID", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedEntity := &TestEntity{ID: 1, Name: "test"}
		expectedError := errors.New("get by non id failed")

		mockRepo.On("GetByNonID", ctx, field, value).Return(expectedEntity, expectedError)

		result, err := mockRepo.GetByNonID(ctx, field, value)
		assert.Equal(t, expectedEntity, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByNonID_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedEntity := &TestEntity{ID: 1, Name: "test"}

		mockRepo.On("GetByNonID", ctx, field, value).Return(expectedEntity, nil)

		result, err := mockRepo.GetByNonID(ctx, field, value)
		assert.Equal(t, expectedEntity, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test1"},
			{ID: 2, Name: "test2"},
		}
		expectedError := errors.New("list failed")

		mockRepo.On("List", ctx).Return(expectedEntities, expectedError)

		result, err := mockRepo.List(ctx)
		assert.Equal(t, expectedEntities, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test1"},
			{ID: 2, Name: "test2"},
		}

		mockRepo.On("List", ctx).Return(expectedEntities, nil)

		result, err := mockRepo.List(ctx)
		assert.Equal(t, expectedEntities, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExecuteQuery", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "SELECT * FROM test WHERE name = ?"
		values := []interface{}{"test"}
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}
		expectedError := errors.New("execute query failed")

		mockRepo.On("ExecuteQuery", ctx, query, values).Return(expectedEntities, expectedError)

		result, err := mockRepo.ExecuteQuery(ctx, query, values...)
		assert.Equal(t, expectedEntities, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExecuteQuery_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "SELECT * FROM test WHERE name = ?"
		values := []interface{}{"test"}
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}

		mockRepo.On("ExecuteQuery", ctx, query, values).Return(expectedEntities, nil)

		result, err := mockRepo.ExecuteQuery(ctx, query, values...)
		assert.Equal(t, expectedEntities, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExecuteCustomQuery", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "UPDATE test SET name = ? WHERE id = ?"
		values := []interface{}{"updated", 1}
		expectedError := errors.New("execute custom query failed")

		mockRepo.On("ExecuteCustomQuery", ctx, query, values).Return(expectedError)

		result := mockRepo.ExecuteCustomQuery(ctx, query, values...)
		assert.Equal(t, expectedError, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExecuteCustomQuery_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "UPDATE test SET name = ? WHERE id = ?"
		values := []interface{}{"updated", 1}

		mockRepo.On("ExecuteCustomQuery", ctx, query, values).Return(nil)

		result := mockRepo.ExecuteCustomQuery(ctx, query, values...)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("BatchCreate", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := []*TestEntity{
			{ID: 1, Name: "test1"},
			{ID: 2, Name: "test2"},
		}
		expectedError := errors.New("batch create failed")

		mockRepo.On("BatchCreate", ctx, data).Return(expectedError)

		result := mockRepo.BatchCreate(ctx, data)
		assert.Equal(t, expectedError, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("BatchCreate_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := []*TestEntity{
			{ID: 1, Name: "test1"},
			{ID: 2, Name: "test2"},
		}

		mockRepo.On("BatchCreate", ctx, data).Return(nil)

		result := mockRepo.BatchCreate(ctx, data)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByField", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}
		expectedError := errors.New("get by field failed")

		mockRepo.On("GetByField", ctx, field, value).Return(expectedEntities, expectedError)

		result, err := mockRepo.GetByField(ctx, field, value)
		assert.Equal(t, expectedEntities, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByField_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}

		mockRepo.On("GetByField", ctx, field, value).Return(expectedEntities, nil)

		result, err := mockRepo.GetByField(ctx, field, value)
		assert.Equal(t, expectedEntities, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByFields", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		conditions := map[string]interface{}{
			"name": "test",
			"id":   1,
		}
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}
		expectedError := errors.New("get by fields failed")

		mockRepo.On("GetByFields", ctx, conditions).Return(expectedEntities, expectedError)

		result, err := mockRepo.GetByFields(ctx, conditions)
		assert.Equal(t, expectedEntities, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByFields_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		conditions := map[string]interface{}{
			"name": "test",
			"id":   1,
		}
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test"},
		}

		mockRepo.On("GetByFields", ctx, conditions).Return(expectedEntities, nil)

		result, err := mockRepo.GetByFields(ctx, conditions)
		assert.Equal(t, expectedEntities, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Count", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		expectedCount := int64(100)
		expectedError := errors.New("count failed")

		mockRepo.On("Count", ctx).Return(expectedCount, expectedError)

		result, err := mockRepo.Count(ctx)
		assert.Equal(t, expectedCount, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Count_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		expectedCount := int64(100)

		mockRepo.On("Count", ctx).Return(expectedCount, nil)

		result, err := mockRepo.Count(ctx)
		assert.Equal(t, expectedCount, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Exists", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		id := 1
		expectedExists := true
		expectedError := errors.New("exists failed")

		mockRepo.On("Exists", ctx, id).Return(expectedExists, expectedError)

		result, err := mockRepo.Exists(ctx, id)
		assert.Equal(t, expectedExists, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Exists_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		id := 1
		expectedExists := true

		mockRepo.On("Exists", ctx, id).Return(expectedExists, nil)

		result, err := mockRepo.Exists(ctx, id)
		assert.Equal(t, expectedExists, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExistsByField", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedExists := true
		expectedError := errors.New("exists by field failed")

		mockRepo.On("ExistsByField", ctx, field, value).Return(expectedExists, expectedError)

		result, err := mockRepo.ExistsByField(ctx, field, value)
		assert.Equal(t, expectedExists, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExistsByField_Success", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		field := "name"
		value := "test"
		expectedExists := true

		mockRepo.On("ExistsByField", ctx, field, value).Return(expectedExists, nil)

		result, err := mockRepo.ExistsByField(ctx, field, value)
		assert.Equal(t, expectedExists, result)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Close", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}

		mockRepo.On("Close").Return()

		mockRepo.Close()
		mockRepo.AssertExpectations(t)
	})
}

func TestMockRepositoryFactory(t *testing.T) {
	t.Run("CreateUserRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.UserDataEntity]{}

		mockFactory.On("CreateUserRepository").Return(expectedRepo)

		result := mockFactory.CreateUserRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateJobRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.JobDataEntity]{}

		mockFactory.On("CreateJobRepository").Return(expectedRepo)

		result := mockFactory.CreateJobRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateTimeJobRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.TimeJobDataEntity]{}

		mockFactory.On("CreateTimeJobRepository").Return(expectedRepo)

		result := mockFactory.CreateTimeJobRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateEventJobRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.EventJobDataEntity]{}

		mockFactory.On("CreateEventJobRepository").Return(expectedRepo)

		result := mockFactory.CreateEventJobRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateConditionJobRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.ConditionJobDataEntity]{}

		mockFactory.On("CreateConditionJobRepository").Return(expectedRepo)

		result := mockFactory.CreateConditionJobRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateTaskRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.TaskDataEntity]{}

		mockFactory.On("CreateTaskRepository").Return(expectedRepo)

		result := mockFactory.CreateTaskRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateKeeperRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.KeeperDataEntity]{}

		mockFactory.On("CreateKeeperRepository").Return(expectedRepo)

		result := mockFactory.CreateKeeperRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})

	t.Run("CreateApiKeyRepository", func(t *testing.T) {
		mockFactory := &MockRepositoryFactory{}
		expectedRepo := &MockGenericRepository[types.ApiKeyDataEntity]{}

		mockFactory.On("CreateApiKeyRepository").Return(expectedRepo)

		result := mockFactory.CreateApiKeyRepository()
		assert.Equal(t, expectedRepo, result)
		mockFactory.AssertExpectations(t)
	})
}

// TestMockGenericRepositoryCRUDOperations tests complete CRUD operations
func TestMockGenericRepositoryCRUDOperations(t *testing.T) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	t.Run("CompleteCRUDCycle", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()

		// Create
		createData := &TestEntity{ID: 1, Name: "test"}
		mockRepo.On("Create", ctx, createData).Return(nil)

		// Read
		readID := 1
		expectedEntity := &TestEntity{ID: 1, Name: "test"}
		mockRepo.On("GetByID", ctx, readID).Return(expectedEntity, nil)

		// Update
		updateData := &TestEntity{ID: 1, Name: "updated"}
		mockRepo.On("Update", ctx, updateData).Return(nil)

		// Execute operations
		err := mockRepo.Create(ctx, createData)
		assert.NoError(t, err)

		result, err := mockRepo.GetByID(ctx, readID)
		assert.NoError(t, err)
		assert.Equal(t, expectedEntity, result)

		err = mockRepo.Update(ctx, updateData)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// TestMockGenericRepositoryBatchOperations tests batch operations
func TestMockGenericRepositoryBatchOperations(t *testing.T) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	t.Run("BatchCreateLargeDataset", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()

		// Create a large dataset
		data := make([]*TestEntity, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = &TestEntity{ID: i, Name: "test"}
		}

		mockRepo.On("BatchCreate", ctx, data).Return(nil)

		result := mockRepo.BatchCreate(ctx, data)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})
}

// TestMockGenericRepositoryComplexQueries tests complex query operations
func TestMockGenericRepositoryComplexQueries(t *testing.T) {
	type TestEntity struct {
		ID       int    `cql:"id"`
		Name     string `cql:"name"`
		Category string `cql:"category"`
		Status   string `cql:"status"`
	}

	t.Run("ComplexSelectQuery", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "SELECT * FROM test WHERE category = ? AND status = ? AND name LIKE ?"
		values := []interface{}{"tech", "active", "%test%"}
		expectedEntities := []*TestEntity{
			{ID: 1, Name: "test1", Category: "tech", Status: "active"},
			{ID: 2, Name: "test2", Category: "tech", Status: "active"},
		}

		mockRepo.On("ExecuteQuery", ctx, query, values).Return(expectedEntities, nil)

		result, err := mockRepo.ExecuteQuery(ctx, query, values...)
		assert.NoError(t, err)
		assert.Equal(t, expectedEntities, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ComplexUpdateQuery", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		query := "UPDATE test SET status = ? WHERE category = ? AND name LIKE ?"
		values := []interface{}{"inactive", "tech", "%test%"}

		mockRepo.On("ExecuteCustomQuery", ctx, query, values).Return(nil)

		result := mockRepo.ExecuteCustomQuery(ctx, query, values...)
		assert.NoError(t, result)
		mockRepo.AssertExpectations(t)
	})
}

// TestMockGenericRepositoryErrorHandling tests error scenarios
func TestMockGenericRepositoryErrorHandling(t *testing.T) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	t.Run("ConnectionTimeout", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 1, Name: "test"}
		expectedError := errors.New("connection timeout")

		mockRepo.On("Create", ctx, data).Return(expectedError)

		result := mockRepo.Create(ctx, data)
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "connection timeout")
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidData", func(t *testing.T) {
		mockRepo := &MockGenericRepository[TestEntity]{}
		ctx := context.Background()
		data := &TestEntity{ID: 0, Name: ""} // Invalid data
		expectedError := errors.New("invalid data: ID and Name are required")

		mockRepo.On("Create", ctx, data).Return(expectedError)

		result := mockRepo.Create(ctx, data)
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "invalid data")
		mockRepo.AssertExpectations(t)
	})
}

// TestMockGenericRepositoryWithRealTypes tests with actual types from the project
func TestMockGenericRepositoryWithRealTypes(t *testing.T) {
	t.Run("UserDataEntityRepository", func(t *testing.T) {
		mockRepo := &MockGenericRepository[types.UserDataEntity]{}
		ctx := context.Background()
		user := &types.UserDataEntity{
			UserAddress:   "0x123456789",
			EmailID:       "test@example.com",
			JobIDs:        []string{},
			UserPoints:    "0",
			TotalJobs:     0,
			TotalTasks:    0,
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}

		mockRepo.On("Create", ctx, user).Return(nil)
		mockRepo.On("GetByID", ctx, int64(123)).Return(user, nil)

		err := mockRepo.Create(ctx, user)
		assert.NoError(t, err)

		result, err := mockRepo.GetByID(ctx, int64(123))
		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("JobDataEntityRepository", func(t *testing.T) {
		mockRepo := &MockGenericRepository[types.JobDataEntity]{}
		ctx := context.Background()
		job := &types.JobDataEntity{
			JobID:             "123",
			JobTitle:          "test job",
			TaskDefinitionID:  1,
			CreatedChainID:    "chain1",
			UserAddress:       "0x123456789",
			LinkJobID:         "0",
			ChainStatus:       1,
			Timezone:          "UTC",
			IsImua:            false,
			JobType:           "time",
			TimeFrame:         3600,
			Recurring:         false,
			Status:            "active",
			JobCostPrediction: "100",
			JobCostActual:     "10",
			TaskIDs:           []int64{},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			LastExecutedAt:    time.Now(),
		}

		mockRepo.On("Create", ctx, job).Return(nil)
		mockRepo.On("List", ctx).Return([]*types.JobDataEntity{job}, nil)

		err := mockRepo.Create(ctx, job)
		assert.NoError(t, err)

		result, err := mockRepo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, job, result[0])
		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests for performance validation
func BenchmarkMockGenericRepositoryCreate(b *testing.B) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	mockRepo := &MockGenericRepository[TestEntity]{}
	ctx := context.Background()
	data := &TestEntity{ID: 1, Name: "test"}

	mockRepo.On("Create", ctx, data).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockRepo.Create(ctx, data)
	}
}

func BenchmarkMockGenericRepositoryGetByID(b *testing.B) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	mockRepo := &MockGenericRepository[TestEntity]{}
	ctx := context.Background()
	id := 1
	expectedEntity := &TestEntity{ID: 1, Name: "test"}

	mockRepo.On("GetByID", ctx, id).Return(expectedEntity, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mockRepo.GetByID(ctx, id)
	}
}

func BenchmarkMockGenericRepositoryList(b *testing.B) {
	type TestEntity struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}

	mockRepo := &MockGenericRepository[TestEntity]{}
	ctx := context.Background()
	expectedEntities := []*TestEntity{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "test2"},
	}

	mockRepo.On("List", ctx).Return(expectedEntities, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mockRepo.List(ctx)
	}
}

func BenchmarkMockRepositoryFactoryCreateUserRepository(b *testing.B) {
	mockFactory := &MockRepositoryFactory{}
	expectedRepo := &MockGenericRepository[types.UserDataEntity]{}

	mockFactory.On("CreateUserRepository").Return(expectedRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockFactory.CreateUserRepository()
	}
}
