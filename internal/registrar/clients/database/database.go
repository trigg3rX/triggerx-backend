package database

import (
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// DatabaseClient handles database operations
type DatabaseClient struct {
	logger logging.Logger
	db     *database.Connection
}

// NewDatabaseClient initializes the database manager with a logger
func NewDatabaseClient(logger logging.Logger, connection *database.Connection) *DatabaseClient {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if connection == nil {
		panic("database connection cannot be nil")
	}

	return &DatabaseClient{
		logger: logger.With("component", "database"),
		db:     connection,
	}
}

func (c *DatabaseClient) Close() {
	c.db.Close()
}