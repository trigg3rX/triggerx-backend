package interfaces

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GenericRepository defines the interface for generic database operations
type GenericRepository[T any] interface {
	Create(ctx context.Context, data *T) error
	Update(ctx context.Context, data *T) error
	GetByID(ctx context.Context, id interface{}) (*T, error)
	GetByNonID(ctx context.Context, field string, value interface{}) (*T, error)
	List(ctx context.Context) ([]*T, error)
	ExecuteQuery(ctx context.Context, query string, values ...interface{}) ([]*T, error)
	ExecuteCustomQuery(ctx context.Context, query string, values ...interface{}) error
	BatchCreate(ctx context.Context, data []*T) error
	GetByField(ctx context.Context, field string, value interface{}) ([]*T, error)
	GetByFields(ctx context.Context, conditions map[string]interface{}) ([]*T, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id interface{}) (bool, error)
	ExistsByField(ctx context.Context, field string, value interface{}) (bool, error)
	Close()
}

// RepositoryFactory provides access to all table repositories
type RepositoryFactory interface {
	CreateUserRepository() GenericRepository[types.UserDataEntity]
	CreateJobRepository() GenericRepository[types.JobDataEntity]
	CreateTimeJobRepository() GenericRepository[types.TimeJobDataEntity]
	CreateEventJobRepository() GenericRepository[types.EventJobDataEntity]
	CreateConditionJobRepository() GenericRepository[types.ConditionJobDataEntity]
	CreateTaskRepository() GenericRepository[types.TaskDataEntity]
	CreateKeeperRepository() GenericRepository[types.KeeperDataEntity]
	CreateApiKeyRepository() GenericRepository[types.ApiKeyDataEntity]
}
