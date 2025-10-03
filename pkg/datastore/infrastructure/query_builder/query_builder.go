package query_builder

import (
	"context"
	"fmt"

	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// GocqlxQueryBuilder implements query builder for ScyllaDB using gocqlx
type GocqlxQueryBuilder[T any] struct {
	session          interfaces.GocqlxSessioner
	logger           logging.Logger
	table            string
	preparedStmts    map[string]interfaces.GocqlxQueryer
	batchSize        int
	consistencyLevel string
}

// NewGocqlxQueryBuilder creates a new gocqlx-based query builder
func NewGocqlxQueryBuilder[T any](session interfaces.GocqlxSessioner, logger logging.Logger, table string) interfaces.ScyllaQueryBuilder[T] {
	return &GocqlxQueryBuilder[T]{
		session:          session,
		logger:           logger,
		table:            table,
		preparedStmts:    make(map[string]interfaces.GocqlxQueryer),
		batchSize:        100,
		consistencyLevel: "QUORUM",
	}
}

// Insert inserts a new record using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Insert(ctx context.Context, data *T) error {
	// Build insert query using qb
	stmt, names := qb.Insert(gqb.table).ToCql()
	query := gqb.getPreparedStatement("insert", stmt, names)

	if err := query.BindStruct(data).WithContext(ctx).ExecRelease(); err != nil {
		gqb.logger.Errorf("gocqlx insert failed: %v", err)
		return fmt.Errorf("gocqlx insert failed: %w", err)
	}

	return nil
}

// Update updates an existing record using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Update(ctx context.Context, data *T) error {
	// Build update query using qb - this is a simplified version
	// In practice, you'd need to know the primary key fields
	stmt, names := qb.Update(gqb.table).ToCql()
	query := gqb.getPreparedStatement("update", stmt, names)

	if err := query.BindStruct(data).WithContext(ctx).ExecRelease(); err != nil {
		gqb.logger.Errorf("gocqlx update failed: %v", err)
		return fmt.Errorf("gocqlx update failed: %w", err)
	}

	return nil
}

// Delete deletes a record using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Delete(ctx context.Context, data *T) error {
	// Build delete query using qb - this is a simplified version
	stmt, names := qb.Delete(gqb.table).ToCql()
	query := gqb.getPreparedStatement("delete", stmt, names)

	if err := query.BindStruct(data).WithContext(ctx).ExecRelease(); err != nil {
		gqb.logger.Errorf("gocqlx delete failed: %v", err)
		return fmt.Errorf("gocqlx delete failed: %w", err)
	}

	return nil
}

// Get retrieves a single record using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Get(ctx context.Context, data *T) (*T, error) {
	// Build select query using qb - this is a simplified version
	stmt, names := qb.Select(gqb.table).ToCql()
	query := gqb.getPreparedStatement("get", stmt, names)

	var result T
	if err := query.BindStruct(data).WithContext(ctx).GetRelease(&result); err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		gqb.logger.Errorf("gocqlx get failed: %v", err)
		return nil, fmt.Errorf("gocqlx get failed: %w", err)
	}

	return &result, nil
}

// Select retrieves multiple records using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Select(ctx context.Context, data *T) ([]T, error) {
	// Build select query using qb
	stmt, names := qb.Select(gqb.table).ToCql()
	query := gqb.getPreparedStatement("select", stmt, names)

	var results []T
	if err := query.BindStruct(data).WithContext(ctx).Select(&results); err != nil {
		gqb.logger.Errorf("gocqlx select failed: %v", err)
		return nil, fmt.Errorf("gocqlx select failed: %w", err)
	}

	return results, nil
}

// SelectAll retrieves all records using gocqlx
func (gqb *GocqlxQueryBuilder[T]) SelectAll(ctx context.Context) ([]T, error) {
	// Build select all query using qb
	stmt, names := qb.Select(gqb.table).ToCql()
	query := gqb.getPreparedStatement("select_all", stmt, names)

	var results []T
	if err := query.WithContext(ctx).SelectRelease(&results); err != nil {
		gqb.logger.Errorf("gocqlx select all failed: %v", err)
		return nil, fmt.Errorf("gocqlx select all failed: %w", err)
	}

	return results, nil
}

// BatchInsert performs batch insert operations using gocqlx
func (gqb *GocqlxQueryBuilder[T]) BatchInsert(ctx context.Context, data []*T) error {
	if len(data) == 0 {
		return nil
	}

	// Build insert query using qb
	stmt, names := qb.Insert(gqb.table).ToCql()
	query := gqb.getPreparedStatement("batch_insert", stmt, names)

	// Convert []*T to []T for gocqlx
	records := make([]T, len(data))
	for i, record := range data {
		records[i] = *record
	}

	if err := query.BindStruct(records).WithContext(ctx).ExecRelease(); err != nil {
		gqb.logger.Errorf("gocqlx batch insert failed: %v", err)
		return fmt.Errorf("gocqlx batch insert failed: %w", err)
	}

	return nil
}

// getPreparedStatement gets or creates a prepared statement
func (gqb *GocqlxQueryBuilder[T]) getPreparedStatement(key string, stmt string, names []string) interfaces.GocqlxQueryer {
	if preparedStmt, exists := gqb.preparedStmts[key]; exists {
		return preparedStmt
	}

	// Create new prepared statement
	query := gqb.session.Query(stmt, names)
	gqb.preparedStmts[key] = query
	return query
}

// SetBatchSize sets the batch size for batch operations
func (gqb *GocqlxQueryBuilder[T]) SetBatchSize(size int) {
	gqb.batchSize = size
}

// SetConsistencyLevel sets the consistency level for queries
func (gqb *GocqlxQueryBuilder[T]) SetConsistencyLevel(level string) {
	gqb.consistencyLevel = level
}

// Close closes the prepared statements
func (gqb *GocqlxQueryBuilder[T]) Close() {
	// Clear prepared statements
	gqb.preparedStmts = make(map[string]interfaces.GocqlxQueryer)
}

// GetTableName returns the table name
func (gqb *GocqlxQueryBuilder[T]) GetTableName() string {
	return gqb.table
}
