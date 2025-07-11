package database

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/database"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
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

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
)

// retryWithBackoff executes the given function with exponential backoff retry logic
func retryWithBackoff[T any](operation func() (T, error), logger logging.Logger) (T, error) {
	var result T
	var err error
	delay := retryDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt < maxRetries {
			logger.Warnf("Attempt %d failed: %v. Retrying in %v...", attempt, err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return result, fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

func (c *DatabaseClient) Close() {
	c.db.Close()
}