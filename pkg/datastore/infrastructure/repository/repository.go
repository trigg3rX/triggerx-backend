package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/query_builder"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// genericRepository implements the GenericRepository interface
type genericRepository[T any] struct {
	db           interfaces.Connection
	queryBuilder interfaces.ScyllaQueryBuilder[T]
	logger       logging.Logger
	tableName    string
	primaryKey   string
}

// NewGenericRepository creates a new generic repository instance
func NewGenericRepository[T any](
	db interfaces.Connection,
	logger logging.Logger,
	tableName string,
	primaryKey string,
) interfaces.GenericRepository[T] {
	// Use gocqlx session for better performance and features
	gocqlxSession := db.GetGocqlxSession()

	queryBuilder := query_builder.NewGocqlxQueryBuilderWithPrimaryKey[T](
		gocqlxSession,
		logger,
		tableName,
		primaryKey,
	)

	return &genericRepository[T]{
		db:           db,
		queryBuilder: queryBuilder,
		logger:       logger,
		tableName:    tableName,
		primaryKey:   primaryKey,
	}
}

// Create creates a new record in the database
func (r *genericRepository[T]) Create(ctx context.Context, data *T) error {
	return r.queryBuilder.Insert(ctx, data)
}

// Update updates an existing record
func (r *genericRepository[T]) Update(ctx context.Context, data *T) error {
	return r.queryBuilder.Update(ctx, data)
}

// GetByID retrieves a record by its primary key
func (r *genericRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	// Create a search entity with the ID
	searchEntity := r.createSearchEntity(id)

	// Use query builder to get the record
	record, err := r.queryBuilder.Get(ctx, searchEntity)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, errors.New("record not found")
	}

	return record, nil
}

// GetByNonID retrieves a record by a non-primary key field
func (r *genericRepository[T]) GetByNonID(ctx context.Context, field string, value interface{}) (*T, error) {
	// Get explicit column list to avoid schema mismatch
	columns := r.getStructColumns()
	columnList := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1 ALLOW FILTERING", columnList, r.tableName, field)

	// WORKAROUND: gocqlx GetRelease has bugs with slices and certain types
	// Use raw gocql scanning instead
	rawSession := r.db.GetSession()
	iter := rawSession.Query(query, value).WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			r.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	var result T
	scanDest := r.getStructScanDestinations(&result, columns)

	if !iter.Scan(scanDest...) {
		if err := iter.Close(); err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.New("record not found")
			}
			return nil, err
		}
		return nil, errors.New("record not found")
	}

	return &result, nil
}

// List retrieves all records
func (r *genericRepository[T]) List(ctx context.Context) ([]*T, error) {
	// Use query builder to get all records
	records, err := r.queryBuilder.SelectAll(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to slice of pointers
	result := make([]*T, len(records))
	for i, record := range records {
		result[i] = &record
	}

	return result, nil
}

// ExecuteQuery executes a custom SELECT query and returns results
func (r *genericRepository[T]) ExecuteQuery(ctx context.Context, query string, values ...interface{}) ([]*T, error) {
	// For custom queries, we'll use the raw gocql session for maximum flexibility
	iter := r.db.GetSession().Query(query, values...).WithContext(ctx).Iter()
	defer func() {
		err := iter.Close()
		if err != nil {
			r.logger.Errorf("Failed to close iterator: %v", err)
		}
	}()

	var results []*T
	var result T
	for iter.Scan(&result) {
		results = append(results, &result)
	}

	return results, nil
}

// ExecuteCustomQuery executes a custom query (INSERT, UPDATE, DELETE, etc.)
func (r *genericRepository[T]) ExecuteCustomQuery(ctx context.Context, query string, values ...interface{}) error {
	return r.db.GetSession().Query(query, values...).WithContext(ctx).Exec()
}

// createSearchEntity creates a search entity with the given ID
func (r *genericRepository[T]) createSearchEntity(id interface{}) *T {
	// Create a new instance of type T
	var entity T

	// Use reflection to set the primary key field
	v := reflect.ValueOf(&entity).Elem()
	t := reflect.TypeOf(entity)

	// Find the primary key field and set its value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cqlTag := field.Tag.Get("cql")

		if cqlTag == r.primaryKey {
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				// Convert the id to the appropriate type
				idValue := reflect.ValueOf(id)
				if idValue.Type().ConvertibleTo(fieldValue.Type()) {
					fieldValue.Set(idValue.Convert(fieldValue.Type()))
				}
			}
			break
		}
	}

	return &entity
}

// Close closes the repository and its resources
func (r *genericRepository[T]) Close() {
	r.queryBuilder.Close()
}

// GetTableName returns the table name
func (r *genericRepository[T]) GetTableName() string {
	return r.tableName
}

// GetPrimaryKey returns the primary key field name
func (r *genericRepository[T]) GetPrimaryKey() string {
	return r.primaryKey
}

// BatchCreate performs batch insert operations
func (r *genericRepository[T]) BatchCreate(ctx context.Context, data []*T) error {
	if len(data) == 0 {
		return nil
	}
	return r.queryBuilder.BatchInsert(ctx, data)
}

// GetByField retrieves records by any field
func (r *genericRepository[T]) GetByField(ctx context.Context, field string, value interface{}) ([]*T, error) {
	// Get explicit column list to avoid schema mismatch
	columns := r.getStructColumns()
	columnList := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? ALLOW FILTERING", columnList, r.tableName, field)

	// WORKAROUND: gocqlx SelectRelease has bugs with slices and certain types
	// Use raw gocql scanning instead
	rawSession := r.db.GetSession()
	iter := rawSession.Query(query, value).WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			r.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	var results []*T
	for {
		var result T
		scanDest := r.getStructScanDestinations(&result, columns)
		if !iter.Scan(scanDest...) {
			break
		}
		results = append(results, &result)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetByFields retrieves records by multiple fields
func (r *genericRepository[T]) GetByFields(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	if len(conditions) == 0 {
		return r.List(ctx)
	}

	// Get explicit column list to avoid schema mismatch
	columns := r.getStructColumns()
	columnList := strings.Join(columns, ", ")

	var whereClauses []string
	var values []interface{}

	for field, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", field))
		values = append(values, value)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ALLOW FILTERING", columnList, r.tableName, strings.Join(whereClauses, " AND "))

	// WORKAROUND: gocqlx SelectRelease has bugs with slices and certain types
	// Use raw gocql scanning instead
	rawSession := r.db.GetSession()
	iter := rawSession.Query(query, values...).WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			r.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	var results []*T
	for {
		var result T
		scanDest := r.getStructScanDestinations(&result, columns)
		if !iter.Scan(scanDest...) {
			break
		}
		results = append(results, &result)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return results, nil
}

// Count returns the total number of records
func (r *genericRepository[T]) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName)

	// Use gocqlx for better performance
	gocqlxSession := r.db.GetGocqlxSession()
	stmt := gocqlxSession.Query(query, []string{})

	var count int64
	if err := stmt.WithContext(ctx).GetRelease(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if a record exists with the given ID
func (r *genericRepository[T]) Exists(ctx context.Context, id interface{}) (bool, error) {
	_, err := r.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "record not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsByField checks if a record exists with the given field value
func (r *genericRepository[T]) ExistsByField(ctx context.Context, field string, value interface{}) (bool, error) {
	_, err := r.GetByNonID(ctx, field, value)
	if err != nil {
		if err.Error() == "record not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getStructColumns extracts column names from the struct using reflection
func (r *genericRepository[T]) getStructColumns() []string {
	var entity T
	t := reflect.TypeOf(entity)

	// If it's a pointer, get the underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cqlTag := field.Tag.Get("cql")

		// Skip fields without cql tag
		if cqlTag != "" {
			columns = append(columns, cqlTag)
		}
	}

	return columns
}

// getStructScanDestinations returns pointers to struct fields in the order of columns
// This is used to work around gocqlx reflection bugs with slices and certain types
func (r *genericRepository[T]) getStructScanDestinations(entity *T, columns []string) []interface{} {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	// Build a map of cql tag -> field index for quick lookup
	tagToField := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cqlTag := field.Tag.Get("cql")
		if cqlTag != "" {
			tagToField[cqlTag] = i
		}
	}

	// Create scan destinations in the order of columns
	scanDest := make([]interface{}, len(columns))
	for i, col := range columns {
		if fieldIdx, ok := tagToField[col]; ok {
			fieldValue := v.Field(fieldIdx)
			if fieldValue.CanAddr() {
				scanDest[i] = fieldValue.Addr().Interface()
			} else {
				r.logger.Warnf("Field %s (column %s) cannot be addressed", t.Field(fieldIdx).Name, col)
				// Create a dummy destination to avoid nil
				var dummy interface{}
				scanDest[i] = &dummy
			}
		} else {
			r.logger.Warnf("Column %s not found in struct %s", col, t.Name())
			// Create a dummy destination to avoid nil
			var dummy interface{}
			scanDest[i] = &dummy
		}
	}

	return scanDest
}
