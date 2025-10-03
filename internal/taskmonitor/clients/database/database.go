package database

import (
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseClient handles database operations
type DatabaseClient struct {
	logger logging.Logger
	taskRepo     interfaces.GenericRepository[types.TaskDataEntity]
	keeperRepo   interfaces.GenericRepository[types.KeeperDataEntity]
	userRepo     interfaces.GenericRepository[types.UserDataEntity]
	jobRepo      interfaces.GenericRepository[types.JobDataEntity]
}

// NewDatabaseClient initializes the database manager with a logger
func NewDatabaseClient(
	logger logging.Logger,
	taskRepo interfaces.GenericRepository[types.TaskDataEntity],
	keeperRepo interfaces.GenericRepository[types.KeeperDataEntity],
	userRepo interfaces.GenericRepository[types.UserDataEntity],
	jobRepo interfaces.GenericRepository[types.JobDataEntity],
) *DatabaseClient {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if taskRepo == nil {
		panic("task repository cannot be nil")
	}
	if keeperRepo == nil {
		panic("keeper repository cannot be nil")
	}
	if userRepo == nil {
		panic("user repository cannot be nil")
	}
	if jobRepo == nil {
		panic("job repository cannot be nil")
	}

	return &DatabaseClient{
		logger: logger.With("component", "database"),
		taskRepo: taskRepo,
		keeperRepo: keeperRepo,
		userRepo: userRepo,
		jobRepo: jobRepo,
	}
}
