package interfaces

import (
	"context"
)

// QueryBuilder interface for generic database operations
type ScyllaQueryBuilder[T any] interface {
	Insert(ctx context.Context, data *T) error
	Update(ctx context.Context, data *T) error
	Delete(ctx context.Context, data *T) error
	Get(ctx context.Context, data *T) (*T, error)
	Select(ctx context.Context, data *T) ([]T, error)
	SelectAll(ctx context.Context) ([]T, error)
	BatchInsert(ctx context.Context, data []*T) error
	SetBatchSize(size int)
	SetConsistencyLevel(level string)
	Close()
	GetTableName() string
}
