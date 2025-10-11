package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockDatastoreService is a mock implementation of DatastoreService interface
type MockDatastoreService struct {
	mock.Mock
}

// HealthCheck mocks the HealthCheck method
func (m *MockDatastoreService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockDatastoreService) Close() {
	m.Called()
}

// GetFactory mocks the GetFactory method
func (m *MockDatastoreService) GetFactory() interfaces.RepositoryFactory {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.RepositoryFactory)
}

// User mocks the User repository access
func (m *MockDatastoreService) User() interfaces.GenericRepository[types.UserDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.UserDataEntity])
}

// Job mocks the Job repository access
func (m *MockDatastoreService) Job() interfaces.GenericRepository[types.JobDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.JobDataEntity])
}

// TimeJob mocks the TimeJob repository access
func (m *MockDatastoreService) TimeJob() interfaces.GenericRepository[types.TimeJobDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.TimeJobDataEntity])
}

// EventJob mocks the EventJob repository access
func (m *MockDatastoreService) EventJob() interfaces.GenericRepository[types.EventJobDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.EventJobDataEntity])
}

// ConditionJob mocks the ConditionJob repository access
func (m *MockDatastoreService) ConditionJob() interfaces.GenericRepository[types.ConditionJobDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.ConditionJobDataEntity])
}

// Task mocks the Task repository access
func (m *MockDatastoreService) Task() interfaces.GenericRepository[types.TaskDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.TaskDataEntity])
}

// Keeper mocks the Keeper repository access
func (m *MockDatastoreService) Keeper() interfaces.GenericRepository[types.KeeperDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.KeeperDataEntity])
}

// ApiKey mocks the ApiKey repository access
func (m *MockDatastoreService) ApiKey() interfaces.GenericRepository[types.ApiKeyDataEntity] {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.GenericRepository[types.ApiKeyDataEntity])
}
