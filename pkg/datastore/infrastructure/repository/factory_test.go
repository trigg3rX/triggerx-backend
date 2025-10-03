package repository

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestRepositoryFactory tests the repository factory functionality
func TestRepositoryFactory(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.MockLogger{}

	// Set up expectations for repository creation
	mockGocqlxSessioner := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSessioner).AnyTimes()

	// Create factory
	factory := NewRepositoryFactory(mockConnection, mockLogger)

	t.Run("CreateUserRepository", func(t *testing.T) {
		// Test successful user repository creation
		userRepo := factory.CreateUserRepository()
		assert.NotNil(t, userRepo)
	})

	t.Run("CreateJobRepository", func(t *testing.T) {
		// Test successful job repository creation
		jobRepo := factory.CreateJobRepository()
		assert.NotNil(t, jobRepo)
	})

	t.Run("CreateTimeJobRepository", func(t *testing.T) {
		// Test successful time job repository creation
		timeJobRepo := factory.CreateTimeJobRepository()
		assert.NotNil(t, timeJobRepo)
	})

	t.Run("CreateEventJobRepository", func(t *testing.T) {
		// Test successful event job repository creation
		eventJobRepo := factory.CreateEventJobRepository()
		assert.NotNil(t, eventJobRepo)
	})

	t.Run("CreateConditionJobRepository", func(t *testing.T) {
		// Test successful condition job repository creation
		conditionJobRepo := factory.CreateConditionJobRepository()
		assert.NotNil(t, conditionJobRepo)
	})

	t.Run("CreateTaskRepository", func(t *testing.T) {
		// Test successful task repository creation
		taskRepo := factory.CreateTaskRepository()
		assert.NotNil(t, taskRepo)
	})

	t.Run("CreateKeeperRepository", func(t *testing.T) {
		// Test successful keeper repository creation
		keeperRepo := factory.CreateKeeperRepository()
		assert.NotNil(t, keeperRepo)
	})

	t.Run("CreateApiKeyRepository", func(t *testing.T) {
		// Test successful API key repository creation
		apiKeyRepo := factory.CreateApiKeyRepository()
		assert.NotNil(t, apiKeyRepo)
	})
}

// TestRepositoryFactoryConcurrency tests concurrent access to the factory
func TestRepositoryFactoryConcurrency(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.MockLogger{}

	// Set up expectations for repository creation
	mockGocqlxSessioner := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSessioner).AnyTimes()

	// Create factory
	factory := NewRepositoryFactory(mockConnection, mockLogger)

	t.Run("ConcurrentRepositoryCreation", func(t *testing.T) {
		// Test concurrent repository creation
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				// Create different types of repositories concurrently
				userRepo := factory.CreateUserRepository()
				jobRepo := factory.CreateJobRepository()
				taskRepo := factory.CreateTaskRepository()

				assert.NotNil(t, userRepo)
				assert.NotNil(t, jobRepo)
				assert.NotNil(t, taskRepo)

				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestRepositoryFactoryThreadSafety tests thread safety of the factory
func TestRepositoryFactoryThreadSafety(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.MockLogger{}

	// Set up expectations for repository creation
	mockGocqlxSessioner := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSessioner).AnyTimes()

	// Create factory
	factory := NewRepositoryFactory(mockConnection, mockLogger)

	t.Run("ThreadSafeRepositoryCreation", func(t *testing.T) {
		// Test that multiple goroutines can safely create repositories
		done := make(chan bool, 20)

		// Create user repositories concurrently
		for i := 0; i < 10; i++ {
			go func() {
				userRepo := factory.CreateUserRepository()
				assert.NotNil(t, userRepo)
				done <- true
			}()
		}

		// Create job repositories concurrently
		for i := 0; i < 10; i++ {
			go func() {
				jobRepo := factory.CreateJobRepository()
				assert.NotNil(t, jobRepo)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 20; i++ {
			<-done
		}
	})
}
