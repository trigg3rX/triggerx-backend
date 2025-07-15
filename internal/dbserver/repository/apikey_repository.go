package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type ApiKeysRepository interface {
	CreateApiKey(apiKey *types.ApiKeyData) error
	GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error) // changed to return slice
	GetApiKeyDataByKey(key string) (*types.ApiKeyData, error)
	GetApiKeyCounters(key string) (*types.ApiKeyCounters, error)
	GetApiKeyByOwner(owner string) (key string, err error)
	GetApiOwnerByApiKey(key string) (owner string, err error)
	UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error
	UpdateApiKeyStatus(apiKey *types.UpdateApiKeyStatusRequest) error
	UpdateApiKeyLastUsed(key string, isSuccess bool) error
	DeleteApiKey(key string) error
}

type apiKeysRepository struct {
	db *database.Connection
}

func NewApiKeysRepository(db *database.Connection) ApiKeysRepository {
	return &apiKeysRepository{
		db: db,
	}
}

func (r *apiKeysRepository) CreateApiKey(apiKey *types.ApiKeyData) error {
	err := r.db.Session().Query(queries.CreateApiKeyQuery, apiKey.Key, apiKey.Owner, apiKey.IsActive, apiKey.RateLimit, apiKey.LastUsed, apiKey.CreatedAt).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error) {
	iter := r.db.Session().Query(queries.GetApiKeyDataByOwnerQuery, owner).Iter()
	var apiKeys []*types.ApiKeyData
	var key, ownerVal string
	var isActive bool
	var rateLimit int
	var successCount, failedCount int64
	var lastUsed, createdAt time.Time
	for iter.Scan(&key, &ownerVal, &isActive, &rateLimit, &successCount, &failedCount, &lastUsed, &createdAt) {
		apiKeys = append(apiKeys, &types.ApiKeyData{
			Key:          key,
			Owner:        ownerVal,
			IsActive:     isActive,
			RateLimit:    rateLimit,
			SuccessCount: successCount,
			FailedCount:  failedCount,
			LastUsed:     lastUsed,
			CreatedAt:    createdAt,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	if len(apiKeys) == 0 {
		return nil, errors.New("owner not found")
	}
	return apiKeys, nil
}

func (r *apiKeysRepository) GetApiKeyDataByKey(key string) (*types.ApiKeyData, error) {
	apiKey := &types.ApiKeyData{}
	err := r.db.Session().Query(queries.GetApiKeyDataByApiKeyQuery, key).Scan(&apiKey.Key, &apiKey.Owner, &apiKey.IsActive, &apiKey.RateLimit, &apiKey.LastUsed, &apiKey.CreatedAt)
	if err == gocql.ErrNotFound {
		return nil, errors.New("api key not found")
	}
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *apiKeysRepository) GetApiKeyCounters(key string) (*types.ApiKeyCounters, error) {
	callCount := &types.ApiKeyCounters{}
	err := r.db.Session().Query(queries.GetApiKeyCallCountQuery, key).Scan(&callCount.SuccessCount, &callCount.FailedCount)
	if err == gocql.ErrNotFound {
		return nil, errors.New("api key not found")
	}
	if err != nil {
		return nil, err
	}
	return callCount, nil
}

func (r *apiKeysRepository) GetApiKeyByOwner(owner string) (key string, err error) {
	key = ""
	err = r.db.Session().Query(queries.GetApiKeyByOwnerQuery, owner).Scan(&key)
	if err == gocql.ErrNotFound {
		return "", errors.New("owner not found")
	}
	if err != nil {
		return "", err
	}
	return key, nil
}

func (r *apiKeysRepository) GetApiOwnerByApiKey(key string) (owner string, err error) {
	owner = ""
	err = r.db.Session().Query(queries.GetApiOwnerByApiKeyQuery, key).Scan(&owner)
	if err == gocql.ErrNotFound {
		return "", errors.New("api key not found")
	}
	if err != nil {
		return "", err
	}
	return owner, nil
}

func (r *apiKeysRepository) UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error {
	err := r.db.Session().Query(queries.UpdateApiKeyQuery, apiKey.Key, apiKey.IsActive, apiKey.RateLimit).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) UpdateApiKeyStatus(apiKey *types.UpdateApiKeyStatusRequest) error {
	err := r.db.Session().Query(queries.UpdateApiKeyStatusQuery, apiKey.IsActive, apiKey.Key).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) UpdateApiKeyLastUsed(key string, isSuccess bool) error {
	if isSuccess {
		err := r.db.Session().Query(queries.UpdateApiKeyLastUsedQuery, key, time.Now(), 1, 0).Exec()
		if err != nil {
			return err
		}
	} else {
		err := r.db.Session().Query(queries.UpdateApiKeyLastUsedQuery, key, time.Now(), 0, 1).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteApiKey physically deletes an API key from the apikeys table
func (r *apiKeysRepository) DeleteApiKey(key string) error {
	err := r.db.Session().Query(queries.DeleteApiKeyQuery, key).Exec()
	if err != nil {
		return err
	}
	return nil
}
