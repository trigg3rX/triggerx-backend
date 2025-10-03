package datastore

import (
	"context"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatastoreService is the main service interface for database operations
type DatastoreService interface {
	// Repository factory access
	GetFactory() interfaces.RepositoryFactory

	// Connection management
	HealthCheck(ctx context.Context) error
	Close()

	// Direct repository access (convenience methods)
	User() interfaces.GenericRepository[types.UserDataEntity]
	Job() interfaces.GenericRepository[types.JobDataEntity]
	TimeJob() interfaces.GenericRepository[types.TimeJobDataEntity]
	EventJob() interfaces.GenericRepository[types.EventJobDataEntity]
	ConditionJob() interfaces.GenericRepository[types.ConditionJobDataEntity]
	Task() interfaces.GenericRepository[types.TaskDataEntity]
	Keeper() interfaces.GenericRepository[types.KeeperDataEntity]
	ApiKey() interfaces.GenericRepository[types.ApiKeyDataEntity]
}

// datastoreService implements the DatastoreService interface
type datastoreService struct {
	connection   interfaces.Connection
	factory      interfaces.RepositoryFactory
	logger       logging.Logger
}

// NewService creates a new datastore service
func NewService(config *connection.Config, logger logging.Logger) (DatastoreService, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Create connection
	conn, err := connection.NewConnection(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Create repository factory
	factory := repository.NewRepositoryFactory(conn, logger)

	return &datastoreService{
		connection:   conn,
		factory:      factory,
		logger:       logger,
	}, nil
}

// Repositories returns the repository factory
func (ds *datastoreService) GetFactory() interfaces.RepositoryFactory {
	return ds.factory
}

// HealthCheck performs a health check on the database connection
func (ds *datastoreService) HealthCheck(ctx context.Context) error {
	return ds.connection.HealthCheck(ctx)
}

// Close closes the database connection
func (ds *datastoreService) Close() {
	ds.connection.Close()
}

// User returns the user repository
func (ds *datastoreService) User() interfaces.GenericRepository[types.UserDataEntity] {
	return ds.factory.CreateUserRepository()
}

// Job returns the job repository
func (ds *datastoreService) Job() interfaces.GenericRepository[types.JobDataEntity] {
	return ds.factory.CreateJobRepository()
}

// TimeJob returns the time job repository
func (ds *datastoreService) TimeJob() interfaces.GenericRepository[types.TimeJobDataEntity] {
	return ds.factory.CreateTimeJobRepository()
}

// EventJob returns the event job repository
func (ds *datastoreService) EventJob() interfaces.GenericRepository[types.EventJobDataEntity] {
	return ds.factory.CreateEventJobRepository()
}

// ConditionJob returns the condition job repository
func (ds *datastoreService) ConditionJob() interfaces.GenericRepository[types.ConditionJobDataEntity] {
	return ds.factory.CreateConditionJobRepository()
}

// Task returns the task repository
func (ds *datastoreService) Task() interfaces.GenericRepository[types.TaskDataEntity] {
	return ds.factory.CreateTaskRepository()
}

// Keeper returns the keeper repository
func (ds *datastoreService) Keeper() interfaces.GenericRepository[types.KeeperDataEntity] {
	return ds.factory.CreateKeeperRepository()
}

// ApiKey returns the API key repository
func (ds *datastoreService) ApiKey() interfaces.GenericRepository[types.ApiKeyDataEntity] {
	return ds.factory.CreateApiKeyRepository()
}
