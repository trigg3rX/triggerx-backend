package repository

import (
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

// ScriptStorageRepository defines the interface for script storage operations.
type ScriptStorageRepository interface {
	// GetStorageByJobID retrieves all storage key-value pairs for a job.
	GetStorageByJobID(jobID *big.Int) (map[string]string, error)
}

type scriptStorageRepository struct {
	db *database.Connection
}

// NewScriptStorageRepository creates a new script storage repository instance.
func NewScriptStorageRepository(db *database.Connection) ScriptStorageRepository {
	return &scriptStorageRepository{
		db: db,
	}
}

// GetStorageByJobID retrieves all storage key-value pairs for a job.
func (r *scriptStorageRepository) GetStorageByJobID(jobID *big.Int) (map[string]string, error) {
	iter := r.db.Session().Query(getStorageByJobIDQuery, jobID).Iter()

	storage := make(map[string]string)
	var key, value string
	var updatedAt time.Time

	for iter.Scan(&key, &value, &updatedAt) {
		storage[key] = value
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return storage, nil
}

// Query constants
const (
	getStorageByJobIDQuery = `
		SELECT storage_key, storage_value, updated_at
		FROM triggerx.script_storage
		WHERE job_id = ?`
)
