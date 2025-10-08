package interfaces

import (
	"context"

	"github.com/gocql/gocql"
)

// Connection interface for database connection management
type Connection interface {
	GetSession() Sessioner
	GetGocqlxSession() GocqlxSessioner
	Close()
	HealthCheck(ctx context.Context) error
}

// Sessioner interface for database session operations
type Sessioner interface {
	Query(stmt string, values ...interface{}) *gocql.Query
	Close()
}

// Query interface for database query operations
type Query interface {
	WithContext(ctx context.Context) Query
	BindStruct(data interface{}) Query
	Exec() error
	Scan(dest ...interface{}) error
	Iter() Iter
}

// Iter interface for database iteration operations
type Iter interface {
	Scan(dest ...interface{}) bool
	Close() error
}

// GocqlxSessioner interface for gocqlx session operations
type GocqlxSessioner interface {
	Query(stmt string, names []string) GocqlxQueryer
	Close()
}

// GocqlxQueryer interface for gocqlx query operations
type GocqlxQueryer interface {
	WithContext(ctx context.Context) GocqlxQueryer
	BindStruct(data interface{}) GocqlxQueryer
	BindMap(data map[string]interface{}) GocqlxQueryer
	ExecRelease() error
	GetRelease(dest interface{}) error
	Select(dest interface{}) error
	SelectRelease(dest interface{}) error
}
