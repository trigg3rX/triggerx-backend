package query_builder

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gocql/gocql"
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
	// Get column names from struct to build proper INSERT query
	columns := gqb.getStructColumns()

	// WORKAROUND: gocqlx BindStruct has bugs with collections (set, list) and big.Int
	// Use raw gocql with manual value extraction
	values := gqb.getStructValues(data, columns)

	// Build insert query - create placeholders
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		gqb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// Use raw session for insert
	rawSession := gqb.session.RawSession()
	if err := rawSession.Query(stmt, values...).WithContext(ctx).Exec(); err != nil {
		gqb.logger.Errorf("gocql insert failed: %v", err)
		return fmt.Errorf("gocql insert failed: %w", err)
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

	// WORKAROUND: gocqlx BindMap has bugs with collections - use raw gocql
	// Extract values in the order of fieldNames
	queryValues := make([]interface{}, len(fieldNames))
	for i, fieldName := range fieldNames {
		queryValues[i] = valuesMap[fieldName]
	}

	// Use raw session for update
	rawSession := gqb.session.RawSession()
	if err := rawSession.Query(query, queryValues...).WithContext(ctx).Exec(); err != nil {
		gqb.logger.Errorf("gocql partial update failed: %v", err)
		return fmt.Errorf("gocql partial update failed: %w", err)
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
			isPrimaryKey = strings.HasSuffix(cqlTag, "_id") || cqlTag == "key" || strings.HasSuffix(cqlTag, "_address")
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
	// WORKAROUND: gocqlx BindStruct has bugs - use raw gocql
	// Build delete query - we need to find the primary key
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var primaryKeyField string
	var primaryKeyValue interface{}

	// Find the primary key
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
			isPrimaryKey = strings.HasSuffix(cqlTag, "_id") || cqlTag == "key" || strings.HasSuffix(cqlTag, "_address")
		}

		if isPrimaryKey && !fieldValue.IsZero() {
			primaryKeyField = cqlTag
			primaryKeyValue = fieldValue.Interface()
			break
		}
	}

	if primaryKeyField == "" {
		return fmt.Errorf("no primary key field found for delete")
	}

	// Build delete query
	stmt := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", gqb.table, primaryKeyField)

	// Use raw session for delete
	rawSession := gqb.session.RawSession()
	if err := rawSession.Query(stmt, primaryKeyValue).WithContext(ctx).Exec(); err != nil {
		gqb.logger.Errorf("gocql delete failed: %v", err)
		return fmt.Errorf("gocql delete failed: %w", err)
	}

	return nil
}

// Get retrieves a single record using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Get(ctx context.Context, data *T) (*T, error) {
	// Get column names from struct to avoid schema mismatch issues
	columns := gqb.getStructColumns()

	// WORKAROUND: gocqlx has bugs with reflection for set<varchar> -> []string and varint -> *big.Int
	// Use direct gocql.Scan() instead of gocqlx's GetRelease()
	// Build the query parameters from the data struct
	var queryArgs []interface{}
	var whereClauses []string
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	// Find primary key fields and other non-zero fields to build WHERE clause
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		cqlTag := field.Tag.Get("cql")

		if cqlTag == "" {
			continue
		}

		// Only use non-zero fields for WHERE clause
		if !fieldValue.IsZero() {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", cqlTag))
			queryArgs = append(queryArgs, fieldValue.Interface())
		}
	}

	if len(whereClauses) == 0 {
		return nil, fmt.Errorf("no query parameters provided for Get operation")
	}

	// Build select query with explicit columns and WHERE clause
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(columns, ", "),
		gqb.table,
		strings.Join(whereClauses, " AND "))

	// Get the underlying raw session and create a query
	rawSession := gqb.session.RawSession()
	rawQuery := rawSession.Query(stmt, queryArgs...)

	var result T
	scanDest := gqb.getStructScanDestinations(&result, columns)

	iter := rawQuery.WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			gqb.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	if !iter.Scan(scanDest...) {
		if err := iter.Close(); err != nil {
			if err == gocql.ErrNotFound {
				return nil, nil
			}
			gqb.logger.Errorf("gocql scan failed: %v", err)
			return nil, fmt.Errorf("gocql scan failed: %w", err)
		}
		return nil, nil // No results found
	}

	return &result, nil
}

// Select retrieves multiple records using gocqlx
func (gqb *GocqlxQueryBuilder[T]) Select(ctx context.Context, data *T) ([]T, error) {
	// Get column names from struct to avoid schema mismatch issues
	columns := gqb.getStructColumns()

	// WORKAROUND: gocqlx Select has bugs with slices and certain types
	// Use direct gocql.Scan() instead of gocqlx's Select()
	// Build the query parameters from the data struct
	var queryArgs []interface{}
	var whereClauses []string
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	// Find non-zero fields to build WHERE clause
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		cqlTag := field.Tag.Get("cql")

		if cqlTag == "" {
			continue
		}

		// Only use non-zero fields for WHERE clause
		if !fieldValue.IsZero() {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", cqlTag))
			queryArgs = append(queryArgs, fieldValue.Interface())
		}
	}

	if len(whereClauses) == 0 {
		return gqb.SelectAll(ctx)
	}

	// Build select query with explicit columns and WHERE clause
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE %s ALLOW FILTERING",
		strings.Join(columns, ", "),
		gqb.table,
		strings.Join(whereClauses, " AND "))

	// Get the underlying raw session and create a query
	rawSession := gqb.session.RawSession()
	iter := rawSession.Query(stmt, queryArgs...).WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			gqb.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	var results []T
	for {
		var result T
		scanDest := gqb.getStructScanDestinations(&result, columns)
		if !iter.Scan(scanDest...) {
			break
		}
		results = append(results, result)
	}

	if err := iter.Close(); err != nil {
		gqb.logger.Errorf("gocql select failed: %v", err)
		return nil, fmt.Errorf("gocql select failed: %w", err)
	}

	return results, nil
}

// SelectAll retrieves all records using gocqlx
func (gqb *GocqlxQueryBuilder[T]) SelectAll(ctx context.Context) ([]T, error) {
	// Get column names from struct to avoid schema mismatch issues
	columns := gqb.getStructColumns()

	// Build select query with explicit columns
	stmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), gqb.table)

	// WORKAROUND: gocqlx SelectRelease has bugs with slices and certain types
	// Use raw gocql scanning instead
	rawSession := gqb.session.RawSession()
	iter := rawSession.Query(stmt).WithContext(ctx).Iter()
	defer func() {
		if err := iter.Close(); err != nil {
			gqb.logger.Errorf("gocql close failed: %v", err)
		}
	}()

	var results []T
	for {
		var result T
		scanDest := gqb.getStructScanDestinations(&result, columns)
		if !iter.Scan(scanDest...) {
			break
		}
		results = append(results, result)
	}

	if err := iter.Close(); err != nil {
		gqb.logger.Errorf("gocql select all failed: %v", err)
		return nil, fmt.Errorf("gocql select all failed: %w", err)
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

// getStructColumns extracts column names from the struct using reflection
func (gqb *GocqlxQueryBuilder[T]) getStructColumns() []string {
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
// This is used to work around gocqlx reflection bugs with set<varchar> -> []string and varint -> *big.Int
func (gqb *GocqlxQueryBuilder[T]) getStructScanDestinations(entity *T, columns []string) []interface{} {
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
				gqb.logger.Warnf("Field %s (column %s) cannot be addressed", t.Field(fieldIdx).Name, col)
				// Create a dummy destination to avoid nil
				var dummy interface{}
				scanDest[i] = &dummy
			}
		} else {
			gqb.logger.Warnf("Column %s not found in struct %s", col, t.Name())
			// Create a dummy destination to avoid nil
			var dummy interface{}
			scanDest[i] = &dummy
		}
	}

	return scanDest
}

// getStructValues returns field values from struct in the order of columns
// This is used for INSERT operations to work around gocqlx BindStruct bugs
func (gqb *GocqlxQueryBuilder[T]) getStructValues(entity *T, columns []string) []interface{} {
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

	// Create values in the order of columns
	values := make([]interface{}, len(columns))
	for i, col := range columns {
		if fieldIdx, ok := tagToField[col]; ok {
			fieldValue := v.Field(fieldIdx)
			values[i] = fieldValue.Interface()
		}
	}

	return values
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
