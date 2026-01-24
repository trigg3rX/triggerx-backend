package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeysRepository interface {
	CreateApiKey(apiKey *types.ApiKeyDataEntity) error
	GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyDataDTO, error)
	GetApiKeyDataByKey(key string) (*types.ApiKeyDataDTO, error)
	UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error
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

func (r *apiKeysRepository) CreateApiKey(apiKey *types.ApiKeyDataEntity) error {
	err := r.db.Session().Query(CreateApiKeyQuery,
		apiKey.Key, apiKey.Owner, apiKey.IsActive, apiKey.RateLimit,
		apiKey.SuccessCount, apiKey.FailedCount, apiKey.LastUsed, apiKey.CreatedAt).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyDataDTO, error) {
	iter := r.db.Session().Query(GetApiKeyDataByOwnerQuery, owner).Iter()
	var apiKeys []*types.ApiKeyDataDTO

	var entity types.ApiKeyDataEntity
	for iter.Scan(
		&entity.Key, &entity.Owner, &entity.IsActive, &entity.RateLimit,
		&entity.SuccessCount, &entity.FailedCount, &entity.LastUsed, &entity.CreatedAt) {
		dto := types.ApiKeyDataEntityToDTO(&entity)
		apiKeys = append(apiKeys, dto)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	if len(apiKeys) == 0 {
		return nil, errors.New("owner not found")
	}
	return apiKeys, nil
}

func (r *apiKeysRepository) GetApiKeyDataByKey(key string) (*types.ApiKeyDataDTO, error) {
	var entity types.ApiKeyDataEntity
	err := r.db.Session().Query(GetApiKeyDataByApiKeyQuery, key).Scan(
		&entity.Key, &entity.Owner, &entity.IsActive, &entity.RateLimit,
		&entity.SuccessCount, &entity.FailedCount, &entity.LastUsed, &entity.CreatedAt)

	if err == gocql.ErrNotFound {
		return nil, errors.New("api key not found")
	}
	if err != nil {
		return nil, err
	}

	dto := types.ApiKeyDataEntityToDTO(&entity)
	return dto, nil
}

func (r *apiKeysRepository) UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error {
	// Handle nullable fields - need to get current values first if not provided
	var isActive bool
	var rateLimit int

	if apiKey.IsActive != nil {
		isActive = *apiKey.IsActive
	} else {
		// Get current value
		current, err := r.GetApiKeyDataByKey(apiKey.Key)
		if err != nil {
			return err
		}
		isActive = current.IsActive
	}

	if apiKey.RateLimit != nil {
		rateLimit = *apiKey.RateLimit
	} else {
		// Get current value
		current, err := r.GetApiKeyDataByKey(apiKey.Key)
		if err != nil {
			return err
		}
		rateLimit = current.RateLimit
	}

	err := r.db.Session().Query(UpdateApiKeyQuery,
		isActive, rateLimit, apiKey.Key).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *apiKeysRepository) UpdateApiKeyLastUsed(key string, isSuccess bool) error {
	now := time.Now()
	if isSuccess {
		// Increment success_count, reset failed_count to 0
		err := r.db.Session().Query(UpdateApiKeyLastUsedQuery,
			now, 1, 0, key).Exec()
		if err != nil {
			return err
		}
	} else {
		// Increment failed_count, keep success_count
		err := r.db.Session().Query(UpdateApiKeyLastUsedQuery,
			now, 0, 1, key).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteApiKey physically deletes an API key from the apikeys table
func (r *apiKeysRepository) DeleteApiKey(key string) error {
	err := r.db.Session().Query(DeleteApiKeyQuery, key).Exec()
	if err != nil {
		return err
	}
	return nil
}
