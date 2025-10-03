package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockGenericRepository is a mock implementation of GenericRepository interface
type MockGenericRepository[T any] struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockGenericRepository[T]) Create(ctx context.Context, data *T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockGenericRepository[T]) Update(ctx context.Context, data *T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockGenericRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

// GetByNonID mocks the GetByNonID method
func (m *MockGenericRepository[T]) GetByNonID(ctx context.Context, field string, value interface{}) (*T, error) {
	args := m.Called(ctx, field, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

// List mocks the List method
func (m *MockGenericRepository[T]) List(ctx context.Context, limit int, offset int) ([]*T, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*T), args.Error(1)
}

// ExecuteQuery mocks the ExecuteQuery method
func (m *MockGenericRepository[T]) ExecuteQuery(ctx context.Context, query string, values ...interface{}) ([]*T, error) {
	args := m.Called(ctx, query, values)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*T), args.Error(1)
}

// ExecuteCustomQuery mocks the ExecuteCustomQuery method
func (m *MockGenericRepository[T]) ExecuteCustomQuery(ctx context.Context, query string, values ...interface{}) error {
	args := m.Called(ctx, query, values)
	return args.Error(0)
}

// BatchCreate mocks the BatchCreate method
func (m *MockGenericRepository[T]) BatchCreate(ctx context.Context, data []*T) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// GetByField mocks the GetByField method
func (m *MockGenericRepository[T]) GetByField(ctx context.Context, field string, value interface{}) ([]*T, error) {
	args := m.Called(ctx, field, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*T), args.Error(1)
}

// GetByFields mocks the GetByFields method
func (m *MockGenericRepository[T]) GetByFields(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	args := m.Called(ctx, conditions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*T), args.Error(1)
}

// Count mocks the Count method
func (m *MockGenericRepository[T]) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Exists mocks the Exists method
func (m *MockGenericRepository[T]) Exists(ctx context.Context, id interface{}) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// ExistsByField mocks the ExistsByField method
func (m *MockGenericRepository[T]) ExistsByField(ctx context.Context, field string, value interface{}) (bool, error) {
	args := m.Called(ctx, field, value)
	return args.Bool(0), args.Error(1)
}

// Close mocks the Close method
func (m *MockGenericRepository[T]) Close() {
	m.Called()
}

// MockRepositoryFactory is a mock implementation of RepositoryFactory interface
type MockRepositoryFactory struct {
	mock.Mock
}

// CreateUserRepository mocks the CreateUserRepository method
func (m *MockRepositoryFactory) CreateUserRepository() *MockGenericRepository[types.UserDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.UserDataEntity])
}

// CreateJobRepository mocks the CreateJobRepository method
func (m *MockRepositoryFactory) CreateJobRepository() *MockGenericRepository[types.JobDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.JobDataEntity])
}

// CreateTimeJobRepository mocks the CreateTimeJobRepository method
func (m *MockRepositoryFactory) CreateTimeJobRepository() *MockGenericRepository[types.TimeJobDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.TimeJobDataEntity])
}

// CreateEventJobRepository mocks the CreateEventJobRepository method
func (m *MockRepositoryFactory) CreateEventJobRepository() *MockGenericRepository[types.EventJobDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.EventJobDataEntity])
}

// CreateConditionJobRepository mocks the CreateConditionJobRepository method
func (m *MockRepositoryFactory) CreateConditionJobRepository() *MockGenericRepository[types.ConditionJobDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.ConditionJobDataEntity])
}

// CreateTaskRepository mocks the CreateTaskRepository method
func (m *MockRepositoryFactory) CreateTaskRepository() *MockGenericRepository[types.TaskDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.TaskDataEntity])
}

// CreateKeeperRepository mocks the CreateKeeperRepository method
func (m *MockRepositoryFactory) CreateKeeperRepository() *MockGenericRepository[types.KeeperDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.KeeperDataEntity])
}

// CreateApiKeyRepository mocks the CreateApiKeyRepository method
func (m *MockRepositoryFactory) CreateApiKeyRepository() *MockGenericRepository[types.ApiKeyDataEntity] {
	args := m.Called()
	return args.Get(0).(*MockGenericRepository[types.ApiKeyDataEntity])
}
