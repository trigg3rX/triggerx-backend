package database

import "github.com/gocql/gocql"

// Sessioner defines an interface for gocql.Session for easier mocking.
type Sessioner interface {
	Query(stmt string, values ...interface{}) *gocql.Query
	Close()
}
