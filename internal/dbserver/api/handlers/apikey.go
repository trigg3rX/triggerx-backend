package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MaskApiKey masks the API key except for the first 4 and last 4 characters
func MaskApiKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func (h *Handler) CreateApiKey(c *gin.Context) {
	var req types.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "[CreateApiKey] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Owner == "" {
		h.logger.Warn(c.Request.Context(), "[CreateApiKey] Validation failed", observability.String("owner", req.Owner))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required"})
		return
	}
	if req.RateLimit <= 0 {
		req.RateLimit = 60
	}

	apiKey := types.ApiKeyDataEntity{
		Key:       "TGRX-" + uuid.New().String(),
		Owner:     req.Owner,
		IsActive:  true,
		RateLimit: req.RateLimit,
		SuccessCount: 0,
		FailedCount: 0,
		LastUsed:     time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}

	trackDBOp := metrics.TrackDBOperation("create", "apikey_data")
	if err := h.apiKeysRepository.CreateApiKey(&apiKey); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "[CreateApiKey] Failed to insert API key", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusCreated, apiKey)
	h.logger.Info(c.Request.Context(), "[CreateApiKey] Created API key", observability.String("owner", req.Owner))
}

func (h *Handler) UpdateApiKey(c *gin.Context) {
	keyID := c.Param("key")
	var req types.UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	apiKey, err := h.apiKeysRepository.GetApiKeyDataByKey(keyID)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "API key not found", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}
	if req.RateLimit != nil && *req.RateLimit > 0 {
		apiKey.RateLimit = *req.RateLimit
	}

	trackDBOp = metrics.TrackDBOperation("update", "apikey_data")
	if err := h.apiKeysRepository.UpdateApiKey(&req); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "Failed to update API key", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, apiKey)
	h.logger.Info(c.Request.Context(), "[UpdateApiKey] Updated API key", observability.String("key_id", keyID))
}

func (h *Handler) DeleteApiKey(c *gin.Context) {
	keyID := c.Param("key")
	// If the keyID is masked, resolve the real key
	if strings.Contains(keyID, "*") {
		owner := c.Query("owner") // require owner as query param for disambiguation
		if owner == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required when deleting by masked key"})
			return
		}
		apiKeys, err := h.apiKeysRepository.GetApiKeyDataByOwner(owner)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No API keys found for this owner"})
			return
		}
		found := false
		for _, k := range apiKeys {
			if MaskApiKey(k.Key) == keyID {
				keyID = k.Key
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found for this owner"})
			return
		}
	}

	trackDBOp := metrics.TrackDBOperation("update", "apikey_data")
	if err := h.apiKeysRepository.DeleteApiKey(keyID); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "Failed to delete API key", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	trackDBOp(nil)

	c.Status(http.StatusNoContent)
	h.logger.Info(c.Request.Context(), "[DeleteApiKey] Deleted API key", observability.String("key_id", keyID))
}

// GetApiKeysByOwner returns all API keys for a given owner
func (h *Handler) GetApiKeysByOwner(c *gin.Context) {
	owner := c.Param("owner")
	if owner == "" {
		h.logger.Warn(c.Request.Context(), "[GetApiKeysByOwner] Validation failed", observability.String("owner", owner))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required"})
		return
	}

	apiKeys, err := h.apiKeysRepository.GetApiKeyDataByOwner(owner)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetApiKeysByOwner] Failed to retrieve API keys", observability.String("owner", owner), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "No API keys found for this owner"})
		return
	}

	// Mask the API keys before returning
	masked := make([]map[string]interface{}, 0, len(apiKeys))
	for _, k := range apiKeys {
		masked = append(masked, map[string]interface{}{
			"key":           MaskApiKey(k.Key),
			"owner":         k.Owner,
			"is_active":     k.IsActive,
			"rate_limit":    k.RateLimit,
			"success_count": k.SuccessCount,
			"failed_count":  k.FailedCount,
			"last_used":     k.LastUsed,
			"created_at":    k.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, masked)
	h.logger.Debug(c.Request.Context(), "[GetApiKeysByOwner] Retrieved API keys", observability.String("owner", owner), observability.Int("keys_count", len(masked)))
}
