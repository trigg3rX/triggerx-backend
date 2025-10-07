package database

import (
	dbinterfaces "github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManager handles database operations
type DatabaseManager struct {
	logger     logging.Logger
	keeperRepo interfaces.GenericRepository[types.KeeperDataEntity]
}

// InitDatabaseManager initializes the database manager with a logger
func InitDatabaseManager(
	logger logging.Logger,
	keeperRepo interfaces.GenericRepository[types.KeeperDataEntity],
) dbinterfaces.DatabaseManagerInterface {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if keeperRepo == nil {
		panic("keeper repository cannot be nil")
	}

	// Create a new logger with component field and proper level
	dbLogger := logger.With("component", "database")

	instance := &DatabaseManager{
		logger:     dbLogger,
		keeperRepo: keeperRepo,
	}

	return instance
}
