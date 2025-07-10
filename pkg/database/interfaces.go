package database

import "github.com/gocql/gocql"

type Sessioner interface {
	Query(string, ...interface{}) *gocql.Query
	ExecuteBatch(*gocql.Batch) error
	Close()
}