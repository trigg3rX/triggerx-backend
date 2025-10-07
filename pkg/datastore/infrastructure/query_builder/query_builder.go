package query_builder

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// GocqlxQueryBuilder implements query builder for ScyllaDB using gocqlx
type GocqlxQueryBuilder[T any] struct {
	session          interfaces.GocqlxSessioner
	logger           logging.Logger
	table            string
	primaryKey       string
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
		primaryKey:       "", // Will be detected automatically
		preparedStmts:    make(map[string]interfaces.GocqlxQueryer),
		batchSize:        100,
		consistencyLevel: "QUORUM",
	}
}

// NewGocqlxQueryBuilderWithPrimaryKey creates a new gocqlx-based query builder with explicit primary key
func NewGocqlxQueryBuilderWithPrimaryKey[T any](session interfaces.GocqlxSessioner, logger logging.Logger, table string, primaryKey string) interfaces.ScyllaQueryBuilder[T] {
	return &GocqlxQueryBuilder[T]{
		session:          session,
		logger:           logger,
		table:            table,
		primaryKey:       primaryKey,
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

// Update updates an existing record using gocqlx with partial update support
// Only non-zero fields will be updated, preventing overwriting of existing data with empty values
func (gqb *GocqlxQueryBuilder[T]) Update(ctx context.Context, data *T) error {
	// Extract non-zero fields and primary key using reflection
	setClauses, whereClause, fieldNames, valuesMap, err := gqb.buildPartialUpdateQuery(data)
	if err != nil {
		return fmt.Errorf("failed to build partial update query: %w", err)
	}

	if len(setClauses) == 0 {
		gqb.logger.Warn("No fields to update (all fields are zero values)")
		return nil
	}

	// Build the UPDATE query with placeholders
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		gqb.table,
		strings.Join(setClauses, ", "),
		whereClause)

	// Convert values map to ordered array matching fieldNames
	values := make([]interface{}, len(fieldNames))
	for i, name := range fieldNames {
		values[i] = valuesMap[name]
	}

	// Use gocqlx query with named binding
	stmt := gqb.session.Query(query, fieldNames)

	// Bind the values as a slice
	if err := stmt.BindStruct(values).WithContext(ctx).ExecRelease(); err != nil {
		gqb.logger.Errorf("gocqlx partial update failed: %v", err)
		return fmt.Errorf("gocqlx partial update failed: %w", err)
	}

	return nil
}

// buildPartialUpdateQuery builds a partial update query by extracting non-zero fields
func (gqb *GocqlxQueryBuilder[T]) buildPartialUpdateQuery(data *T) ([]string, string, []string, map[string]interface{}, error) {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var setClauses []string
	var fieldNames []string
	values := make(map[string]interface{})
	var primaryKeyField string
	var primaryKeyValue interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		cqlTag := field.Tag.Get("cql")

		if cqlTag == "" {
			continue
		}

		// Check if this is the configured primary key
		isPrimaryKey := false
		if gqb.primaryKey != "" {
			isPrimaryKey = (cqlTag == gqb.primaryKey)
		} else {
			// Fallback: detect primary key by common naming patterns
			isPrimaryKey = strings.HasSuffix(cqlTag, "_id") || cqlTag == "key"
		}

		// Store primary key for WHERE clause
		if isPrimaryKey && primaryKeyField == "" {
			primaryKeyField = cqlTag
			primaryKeyValue = fieldValue.Interface()
			continue
		}

		// Check if field has a non-zero value
		if !gqb.isZeroValue(fieldValue) {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", cqlTag))
			fieldNames = append(fieldNames, cqlTag)
			values[cqlTag] = fieldValue.Interface()
		}
	}

	if primaryKeyField == "" {
		return nil, "", nil, nil, fmt.Errorf("no primary key field found")
	}

	// Build WHERE clause
	whereClause := fmt.Sprintf("%s = ?", primaryKeyField)
	fieldNames = append(fieldNames, primaryKeyField)
	values[primaryKeyField] = primaryKeyValue

	return setClauses, whereClause, fieldNames, values, nil
}

// isZeroValue checks if a reflect.Value is a zero value
func (gqb *GocqlxQueryBuilder[T]) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		// For booleans, we consider false as a valid value to update
		// So we'll allow boolean updates regardless of value
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		// Special handling for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		// For other structs (like big.Int), check if all fields are zero
		// For now, we'll consider non-zero if the struct is not the zero value
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
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
