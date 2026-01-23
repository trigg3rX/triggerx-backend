package repository

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

// ScriptStorageRepository handles script storage (persistent key-value pairs)
type ScriptStorageRepository interface {
	GetStorageByJobID(jobID string) (map[string]string, error)
	GetStorageValue(jobID string, key string) (string, error)
	UpsertStorage(jobID string, key string, value string) error
	DeleteStorageKey(jobID string, key string) error
	DeleteAllStorageForJob(jobID string) error
}

type scriptStorageRepository struct {
	db *database.Connection
}

// NewScriptStorageRepository creates a new script storage repository
func NewScriptStorageRepository(db *database.Connection) ScriptStorageRepository {
	return &scriptStorageRepository{
		db: db,
	}
}

func (r *scriptStorageRepository) GetStorageByJobID(jobID string) (map[string]string, error) {
	iter := r.db.Session().Query(queries.GetStorageByJobIDQuery, jobID).Iter()

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

func (r *scriptStorageRepository) GetStorageValue(jobID string, key string) (string, error) {
	var value string

	err := r.db.Session().Query(queries.GetStorageValueQuery, jobID, key).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (r *scriptStorageRepository) UpsertStorage(jobID string, key string, value string) error {
	return r.db.Session().Query(queries.UpsertStorageQuery,
		jobID,
		key,
		value,
		time.Now(),
	).Exec()
}

func (r *scriptStorageRepository) DeleteStorageKey(jobID string, key string) error {
	return r.db.Session().Query(queries.DeleteStorageKeyQuery, jobID, key).Exec()
}

func (r *scriptStorageRepository) DeleteAllStorageForJob(jobID string) error {
	return r.db.Session().Query(queries.DeleteAllStorageForJobQuery, jobID).Exec()
}

// GetStorageSnapshot returns storage as JSON string for execution context
func (r *scriptStorageRepository) GetStorageSnapshot(jobID string) (string, error) {
	storage, err := r.GetStorageByJobID(jobID)
	if err != nil {
		return "{}", err
	}

	// Convert map to JSON
	if len(storage) == 0 {
		return "{}", nil
	}

	// Simple JSON encoding (for production, use json.Marshal)
	jsonStr := "{"
	first := true
	for k, v := range storage {
		if !first {
			jsonStr += ","
		}
		jsonStr += `"` + k + `":"` + v + `"`
		first = false
	}
	jsonStr += "}"

	return jsonStr, nil
}

// ParseStorageUpdates parses STORAGE_SET commands from stderr
func ParseStorageUpdates(stderr string) map[string]string {
	updates := make(map[string]string)

	// Parse lines like: STORAGE_SET:key=value
	// Implementation depends on actual format
	// For now, return empty map

	return updates
}
