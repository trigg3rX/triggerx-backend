package repository

import (
	"sync"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// repositoryFactory implements the RepositoryFactory interface
type repositoryFactory struct {
	connection interfaces.Connection
	logger     logging.Logger

	// Mutex for thread-safe repository creation
	mu sync.RWMutex
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(connection interfaces.Connection, logger logging.Logger) interfaces.RepositoryFactory {
	return &repositoryFactory{
		connection: connection,
		logger:     logger,
	}
}

// CreateUserRepository returns the user data repository
func (rf *repositoryFactory) CreateUserRepository() interfaces.GenericRepository[types.UserDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.UserDataEntity](
		rf.connection,
		rf.logger,
		"user_data",
		"user_id",
	)
	
}

// CreateJobRepository returns the job data repository
func (rf *repositoryFactory) CreateJobRepository() interfaces.GenericRepository[types.JobDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.JobDataEntity](
		rf.connection,
		rf.logger,
		"job_data",
		"job_id",
	)
}

// CreateTimeJobRepository returns the time job data repository
func (rf *repositoryFactory) CreateTimeJobRepository() interfaces.GenericRepository[types.TimeJobDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.TimeJobDataEntity](
		rf.connection,
		rf.logger,
		"time_job_data",
		"job_id",
	)
}

// CreateEventJobRepository returns the event job data repository
func (rf *repositoryFactory) CreateEventJobRepository() interfaces.GenericRepository[types.EventJobDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.EventJobDataEntity](
		rf.connection,
		rf.logger,
		"event_job_data",
		"job_id",
	)
}

// CreateConditionJobRepository returns the condition job data repository
func (rf *repositoryFactory) CreateConditionJobRepository() interfaces.GenericRepository[types.ConditionJobDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.ConditionJobDataEntity](
		rf.connection,
		rf.logger,
		"condition_job_data",
		"job_id",
	)
}

// CreateTaskRepository returns the task data repository
func (rf *repositoryFactory) CreateTaskRepository() interfaces.GenericRepository[types.TaskDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.TaskDataEntity](
		rf.connection,
		rf.logger,
		"task_data",
		"task_id",
	)
}

// CreateKeeperRepository returns the keeper data repository
func (rf *repositoryFactory) CreateKeeperRepository() interfaces.GenericRepository[types.KeeperDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.KeeperDataEntity](
		rf.connection,
		rf.logger,
		"keeper_data",
		"keeper_address",
	)
}

// CreateApiKeyRepository returns the API key repository
func (rf *repositoryFactory) CreateApiKeyRepository() interfaces.GenericRepository[types.ApiKeyDataEntity] {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return NewGenericRepository[types.ApiKeyDataEntity](
		rf.connection,
		rf.logger,
		"apikeys",
		"key",
	)
}
