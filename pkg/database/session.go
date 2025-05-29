package database

import "github.com/gocql/gocql"

// Sessioner defines the interface for database session operations
type Sessioner interface {
	Query(string, ...interface{}) *gocql.Query
	ExecuteBatch(*gocql.Batch) error
	Close()
}
