package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateApiKey(c *gin.Context) {
	logger := h.getLogger(c)
	var req types.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("POST [CreateApiKey] For owner: %s", req.Owner)

	apiKey := &types.ApiKeyDataEntity{
		Key:          "TGRX-" + uuid.New().String(),
		Owner:        req.Owner,
		IsActive:     true,
		RateLimit:    req.RateLimit,
		SuccessCount: 0,
		FailedCount:  0,
		LastUsed:     time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}

	trackDBOp := metrics.TrackDBOperation("create", "apikey_data")
	if err := h.apiKeysRepository.Create(c.Request.Context(), apiKey); err != nil {
		trackDBOp(err)
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}
	trackDBOp(nil)

	var resp = &types.CreateApiKeyResponse{
		Key:       apiKey.Key,
		Owner:     apiKey.Owner,
		IsActive:  apiKey.IsActive,
		LastUsed:  apiKey.LastUsed,
		CreatedAt: apiKey.CreatedAt,
	}

	logger.Infof("POST [CreateApiKey] Successful, owner: %s, key: %s", req.Owner, apiKey.Key)
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) DeleteApiKey(c *gin.Context) {
	logger := h.getLogger(c)
	var req types.DeleteApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("PUT [DeleteApiKey] For key: %s", req.Key)

	var apiKeyData *types.ApiKeyDataEntity
	// If the keyID is masked, resolve the real key
	if strings.Contains(req.Key, "*") {
		trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
		apiKeys, err := h.apiKeysRepository.GetByField(c.Request.Context(), "owner", req.Owner)
		trackDBOp(err)
		if err != nil || len(apiKeys) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No API keys found for this owner"})
			return
		}
		found := false
		for _, k := range apiKeys {
			if maskApiKey(k.Key) == req.Key {
				apiKeyData = k
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found for this owner"})
			return
		}
	} else {
		// If the key is not masked, fetch it directly by key
		trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
		existingKey, err := h.apiKeysRepository.GetByID(c.Request.Context(), req.Key)
		trackDBOp(err)
		if err != nil || existingKey == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		apiKeyData = existingKey
	}

	// Mark as inactive instead of deleting
	apiKeyData.IsActive = false

	trackDBOp := metrics.TrackDBOperation("update", "apikey_data")
	if err := h.apiKeysRepository.Update(c.Request.Context(), apiKeyData); err != nil {
		trackDBOp(err)
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}
	trackDBOp(nil)

	logger.Infof("PUT [DeleteApiKey] Successful, key: %s", req.Key)
	c.Status(http.StatusOK)
}

// GetApiKeysByOwner returns all API keys for a given owner
func (h *Handler) GetApiKeysByOwner(c *gin.Context) {
	logger := h.getLogger(c)
	owner := c.Param("owner")
	if owner == "" {
		logger.Debugf("Owner is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner is required"})
		return
	}
	logger.Debugf("GET [GetApiKeysByOwner] For owner: %s", c.Param("owner"))

	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	apiKeys, err := h.apiKeysRepository.GetByField(c.Request.Context(), "owner", owner)
	trackDBOp(err)
	if err != nil || len(apiKeys) == 0 {
		logger.Debugf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Mask the API keys before returning
	var resp []*types.GetApiKeyDataResponse
	for _, k := range apiKeys {
		resp = append(resp, &types.GetApiKeyDataResponse{
			Key:          maskApiKey(k.Key),
			Owner:        k.Owner,
			IsActive:     k.IsActive,
			SuccessCount: k.SuccessCount,
			FailedCount:  k.FailedCount,
			LastUsed:     k.LastUsed,
			CreatedAt:    k.CreatedAt,
		})
	}

	logger.Infof("GET [GetApiKeysByOwner] Successful, owner: %s, keys: %d", owner, len(apiKeys))
	c.JSON(http.StatusOK, resp)
}

// Helper Functions
// MaskApiKey masks the API key except for the first 4 and last 4 characters
func maskApiKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
