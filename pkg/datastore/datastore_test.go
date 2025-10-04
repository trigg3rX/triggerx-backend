package datastore

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestNewService tests the NewService function
func TestNewService(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := connection.NewConfig("localhost", "9042")

	// Test service creation - this will fail to connect to actual DB, but tests the validation and setup
	service, err := NewService(config, logger)

	// Should return error due to no actual DB connection, but config should be valid
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to create connection")
}

// TestNewService_NilConfig tests NewService with nil config
func TestNewService_NilConfig(t *testing.T) {
	logger := logging.NewNoOpLogger()

	service, err := NewService(nil, logger)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

// TestNewService_NilLogger tests NewService with nil logger
func TestNewService_NilLogger(t *testing.T) {
	config := connection.NewConfig("localhost", "9042")

	service, err := NewService(config, nil)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "logger cannot be nil")
}

// TestDatastoreService_GetFactory tests the GetFactory method
func TestDatastoreService_GetFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	factory := service.GetFactory()
	assert.Equal(t, mockFactory, factory)
}

// TestDatastoreService_HealthCheck tests the HealthCheck method
func TestDatastoreService_HealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	ctx := context.Background()
	mockConnection.EXPECT().HealthCheck(ctx).Return(nil)

	err := service.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestDatastoreService_HealthCheck_Error tests HealthCheck with error
func TestDatastoreService_HealthCheck_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	ctx := context.Background()
	expectedError := assert.AnError
	mockConnection.EXPECT().HealthCheck(ctx).Return(expectedError)

	err := service.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestDatastoreService_Close tests the Close method
func TestDatastoreService_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockConnection.EXPECT().Close().Times(1)

	service.Close()
}

// TestDatastoreService_User tests the User method
func TestDatastoreService_User(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockUserRepo := &mocks.MockGenericRepository[types.UserDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateUserRepository").Return(mockUserRepo)

	userRepo := service.User()
	assert.Equal(t, mockUserRepo, userRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_Job tests the Job method
func TestDatastoreService_Job(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockJobRepo := &mocks.MockGenericRepository[types.JobDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateJobRepository").Return(mockJobRepo)

	jobRepo := service.Job()
	assert.Equal(t, mockJobRepo, jobRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_TimeJob tests the TimeJob method
func TestDatastoreService_TimeJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockTimeJobRepo := &mocks.MockGenericRepository[types.TimeJobDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateTimeJobRepository").Return(mockTimeJobRepo)

	timeJobRepo := service.TimeJob()
	assert.Equal(t, mockTimeJobRepo, timeJobRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_EventJob tests the EventJob method
func TestDatastoreService_EventJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockEventJobRepo := &mocks.MockGenericRepository[types.EventJobDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateEventJobRepository").Return(mockEventJobRepo)

	eventJobRepo := service.EventJob()
	assert.Equal(t, mockEventJobRepo, eventJobRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_ConditionJob tests the ConditionJob method
func TestDatastoreService_ConditionJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockConditionJobRepo := &mocks.MockGenericRepository[types.ConditionJobDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateConditionJobRepository").Return(mockConditionJobRepo)

	conditionJobRepo := service.ConditionJob()
	assert.Equal(t, mockConditionJobRepo, conditionJobRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_Task tests the Task method
func TestDatastoreService_Task(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockTaskRepo := &mocks.MockGenericRepository[types.TaskDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateTaskRepository").Return(mockTaskRepo)

	taskRepo := service.Task()
	assert.Equal(t, mockTaskRepo, taskRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_Keeper tests the Keeper method
func TestDatastoreService_Keeper(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockKeeperRepo := &mocks.MockGenericRepository[types.KeeperDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateKeeperRepository").Return(mockKeeperRepo)

	keeperRepo := service.Keeper()
	assert.Equal(t, mockKeeperRepo, keeperRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_ApiKey tests the ApiKey method
func TestDatastoreService_ApiKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockApiKeyRepo := &mocks.MockGenericRepository[types.ApiKeyDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	mockFactory.On("CreateApiKeyRepository").Return(mockApiKeyRepo)

	apiKeyRepo := service.ApiKey()
	assert.Equal(t, mockApiKeyRepo, apiKeyRepo)
	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_ConcurrentAccess tests concurrent access to repositories
func TestDatastoreService_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockUserRepo := &mocks.MockGenericRepository[types.UserDataEntity]{}
	mockJobRepo := &mocks.MockGenericRepository[types.JobDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	// Setup mock expectations for concurrent access
	mockFactory.On("CreateUserRepository").Return(mockUserRepo).Maybe()
	mockFactory.On("CreateJobRepository").Return(mockJobRepo).Maybe()

	// Test concurrent access to different repositories
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go func() {
			userRepo := service.User()
			assert.Equal(t, mockUserRepo, userRepo)
			done <- true
		}()
		go func() {
			jobRepo := service.Job()
			assert.Equal(t, mockJobRepo, jobRepo)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_AllRepositories tests all repository methods return correct types
func TestDatastoreService_AllRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	mockUserRepo := &mocks.MockGenericRepository[types.UserDataEntity]{}
	mockJobRepo := &mocks.MockGenericRepository[types.JobDataEntity]{}
	mockTimeJobRepo := &mocks.MockGenericRepository[types.TimeJobDataEntity]{}
	mockEventJobRepo := &mocks.MockGenericRepository[types.EventJobDataEntity]{}
	mockConditionJobRepo := &mocks.MockGenericRepository[types.ConditionJobDataEntity]{}
	mockTaskRepo := &mocks.MockGenericRepository[types.TaskDataEntity]{}
	mockKeeperRepo := &mocks.MockGenericRepository[types.KeeperDataEntity]{}
	mockApiKeyRepo := &mocks.MockGenericRepository[types.ApiKeyDataEntity]{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	// Setup mock expectations
	mockFactory.On("CreateUserRepository").Return(mockUserRepo)
	mockFactory.On("CreateJobRepository").Return(mockJobRepo)
	mockFactory.On("CreateTimeJobRepository").Return(mockTimeJobRepo)
	mockFactory.On("CreateEventJobRepository").Return(mockEventJobRepo)
	mockFactory.On("CreateConditionJobRepository").Return(mockConditionJobRepo)
	mockFactory.On("CreateTaskRepository").Return(mockTaskRepo)
	mockFactory.On("CreateKeeperRepository").Return(mockKeeperRepo)
	mockFactory.On("CreateApiKeyRepository").Return(mockApiKeyRepo)

	// Test all repository methods
	assert.Equal(t, mockUserRepo, service.User())
	assert.Equal(t, mockJobRepo, service.Job())
	assert.Equal(t, mockTimeJobRepo, service.TimeJob())
	assert.Equal(t, mockEventJobRepo, service.EventJob())
	assert.Equal(t, mockConditionJobRepo, service.ConditionJob())
	assert.Equal(t, mockTaskRepo, service.Task())
	assert.Equal(t, mockKeeperRepo, service.Keeper())
	assert.Equal(t, mockApiKeyRepo, service.ApiKey())

	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_InterfaceCompliance tests that the service implements all interface methods
func TestDatastoreService_InterfaceCompliance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	// Verify the service implements the DatastoreService interface
	var _ DatastoreService = service

	// Test that all interface methods exist and are callable
	ctx := context.Background()

	// Setup mock expectations for methods that make actual calls
	mockConnection.EXPECT().HealthCheck(ctx).Return(nil)
	mockConnection.EXPECT().Close()

	// Setup mock expectations for repository methods
	mockUserRepo := &mocks.MockGenericRepository[types.UserDataEntity]{}
	mockJobRepo := &mocks.MockGenericRepository[types.JobDataEntity]{}
	mockTimeJobRepo := &mocks.MockGenericRepository[types.TimeJobDataEntity]{}
	mockEventJobRepo := &mocks.MockGenericRepository[types.EventJobDataEntity]{}
	mockConditionJobRepo := &mocks.MockGenericRepository[types.ConditionJobDataEntity]{}
	mockTaskRepo := &mocks.MockGenericRepository[types.TaskDataEntity]{}
	mockKeeperRepo := &mocks.MockGenericRepository[types.KeeperDataEntity]{}
	mockApiKeyRepo := &mocks.MockGenericRepository[types.ApiKeyDataEntity]{}

	mockFactory.On("CreateUserRepository").Return(mockUserRepo)
	mockFactory.On("CreateJobRepository").Return(mockJobRepo)
	mockFactory.On("CreateTimeJobRepository").Return(mockTimeJobRepo)
	mockFactory.On("CreateEventJobRepository").Return(mockEventJobRepo)
	mockFactory.On("CreateConditionJobRepository").Return(mockConditionJobRepo)
	mockFactory.On("CreateTaskRepository").Return(mockTaskRepo)
	mockFactory.On("CreateKeeperRepository").Return(mockKeeperRepo)
	mockFactory.On("CreateApiKeyRepository").Return(mockApiKeyRepo)

	// These should compile without errors and execute successfully
	_ = service.GetFactory()
	_ = service.HealthCheck(ctx)
	service.Close()
	_ = service.User()
	_ = service.Job()
	_ = service.TimeJob()
	_ = service.EventJob()
	_ = service.ConditionJob()
	_ = service.Task()
	_ = service.Keeper()
	_ = service.ApiKey()

	mockFactory.AssertExpectations(t)
}

// TestDatastoreService_ConnectionError tests service creation with connection error
func TestDatastoreService_ConnectionError(t *testing.T) {
	// This test would require mocking the connection.NewConnection function
	// Since it's not easily mockable, we'll test the error handling path
	// by using an invalid configuration that will cause connection to fail

	logger := logging.NewNoOpLogger()
	// Use invalid config that will cause connection to fail
	config := &connection.Config{
		Hosts:    []string{}, // Empty hosts will cause validation error
		Keyspace: "",         // Empty keyspace will cause validation error
	}

	service, err := NewService(config, logger)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to create connection")
}

// TestDatastoreService_HealthCheckContext tests HealthCheck with different contexts
func TestDatastoreService_HealthCheckContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockFactory := &mocks.MockRepositoryFactory{}
	logger := logging.NewNoOpLogger()

	service := &datastoreService{
		connection: mockConnection,
		factory:    mockFactory,
		logger:     logger,
	}

	// Test with background context
	ctx1 := context.Background()
	mockConnection.EXPECT().HealthCheck(ctx1).Return(nil)
	err := service.HealthCheck(ctx1)
	assert.NoError(t, err)

	// Test with TODO context
	ctx2 := context.TODO()
	mockConnection.EXPECT().HealthCheck(ctx2).Return(nil)
	err = service.HealthCheck(ctx2)
	assert.NoError(t, err)

	// Test with context with timeout
	ctx3, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to test cancelled context
	mockConnection.EXPECT().HealthCheck(ctx3).Return(context.Canceled)
	err = service.HealthCheck(ctx3)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
