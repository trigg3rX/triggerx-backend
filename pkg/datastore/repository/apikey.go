package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeysRepository interface {
	CreateApiKey(apiKey *types.ApiKeyData) error
	GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error)
	GetApiKeyDataByKey(key string) (*types.ApiKeyData, error)
	GetApiKeyCounters(key string) (*types.ApiKeyCounters, error)
	GetApiKeyByOwner(owner string) (key string, err error)
	GetApiOwnerByApiKey(key string) (owner string, err error)
	UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error
	UpdateApiKeyStatus(apiKey *types.UpdateApiKeyStatusRequest) error
	UpdateApiKeyLastUsed(key string, isSuccess bool) error
}

type apiKeysRepository struct {
	db connection.ConnectionManager
}

func NewApiKeysRepository(db connection.ConnectionManager) ApiKeysRepository {
	return &apiKeysRepository{
		db: db,
	}
}

func (r *apiKeysRepository) CreateApiKey(apiKey *types.ApiKeyData) error {
	err := r.db.GetSession().Query(queries.CreateApiKeyQuery, apiKey.Key, apiKey.Owner, apiKey.IsActive, apiKey.RateLimit, apiKey.LastUsed, apiKey.CreatedAt).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error) {
	iter := r.db.GetSession().Query(queries.GetApiKeyDataByOwnerQuery, owner).Iter()
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
	var successCount, failedCount int64
	err := r.db.GetSession().Query(queries.GetApiKeyDataByApiKeyQuery, key).Scan(&apiKey.Key, &apiKey.Owner, &apiKey.IsActive, &apiKey.RateLimit, &successCount, &failedCount, &apiKey.LastUsed, &apiKey.CreatedAt)
	apiKey.SuccessCount = successCount
	apiKey.FailedCount = failedCount
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
	err := r.db.GetSession().Query(queries.GetApiKeyCallCountQuery, key).Scan(&callCount.SuccessCount, &callCount.FailedCount)
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
	err = r.db.GetSession().Query(queries.GetApiKeyByOwnerQuery, owner).Scan(&key)
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
	err = r.db.GetSession().Query(queries.GetApiOwnerByApiKeyQuery, key).Scan(&owner)
	if err == gocql.ErrNotFound {
		return "", errors.New("api key not found")
	}
	if err != nil {
		return "", err
	}
	return owner, nil
}

func (r *apiKeysRepository) UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error {
	err := r.db.GetSession().Query(queries.UpdateApiKeyQuery, apiKey.Key, apiKey.IsActive, apiKey.RateLimit).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) UpdateApiKeyStatus(apiKey *types.UpdateApiKeyStatusRequest) error {
	err := r.db.GetSession().Query(queries.UpdateApiKeyStatusQuery, apiKey.IsActive, apiKey.Key).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) UpdateApiKeyLastUsed(key string, isSuccess bool) error {
	if isSuccess {
		err := r.db.GetSession().Query(queries.UpdateApiKeyLastUsedQuery, time.Now(), 1, 0, key).Exec()
		if err != nil {
			return err
		}
	} else {
		err := r.db.GetSession().Query(queries.UpdateApiKeyLastUsedQuery, time.Now(), 0, 1, key).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
