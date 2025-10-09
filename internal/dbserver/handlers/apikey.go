package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[CreateApiKey] trace_id=%s - Creating API key", traceID)
	var req types.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[CreateApiKey] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Owner == "" {
		h.logger.Warnf("[CreateApiKey] Validation failed: Owner is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required"})
		return
	}

	if req.RateLimit <= 0 {
		h.logger.Infof("[CreateApiKey] RateLimit not provided or invalid for owner %s, defaulting to 60", req.Owner)
		req.RateLimit = 60
	}

	ctx := context.Background()

	// No longer check for existing API key for this owner; allow multiple API keys per owner

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
	if err := h.apiKeysRepository.Create(ctx, apiKey); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[CreateApiKey] Failed to insert API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[CreateApiKey] Successfully created new API key for owner %s (Key: %s)", req.Owner, apiKey.Key)
	c.JSON(http.StatusCreated, apiKey)
}

func (h *Handler) UpdateApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateApiKey] trace_id=%s - Updating API key", traceID)
	keyID := c.Param("key")

	var req types.UpdateApiKeyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	apiKey, err := h.apiKeysRepository.GetByNonID(ctx, "key", keyID)
	trackDBOp(err)
	if err != nil || apiKey == nil {
		h.logger.Errorf("API key not found: %v", err)
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
	if err := h.apiKeysRepository.Update(ctx, apiKey); err != nil {
		trackDBOp(err)
		h.logger.Errorf("Failed to update API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, apiKey)
}

func (h *Handler) DeleteApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[DeleteApiKey] trace_id=%s - Deleting API key", traceID)
	keyID := c.Param("key")

	ctx := context.Background()

	// If the keyID is masked, resolve the real key
	if strings.Contains(keyID, "*") {
		owner := c.Query("owner") // require owner as query param for disambiguation
		if owner == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required when deleting by masked key"})
			return
		}
		apiKeys, err := h.apiKeysRepository.GetByField(ctx, "owner", owner)
		if err != nil || len(apiKeys) == 0 {
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

	// Get the API key to mark as inactive
	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	apiKey, err := h.apiKeysRepository.GetByNonID(ctx, "key", keyID)
	trackDBOp(err)
	if err != nil || apiKey == nil {
		h.logger.Errorf("Failed to find API key: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	// Mark as inactive instead of deleting
	apiKey.IsActive = false

	trackDBOp = metrics.TrackDBOperation("update", "apikey_data")
	if err := h.apiKeysRepository.Update(ctx, apiKey); err != nil {
		trackDBOp(err)
		h.logger.Errorf("Failed to delete API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[DeleteApiKey] Successfully deleted API key: %s", keyID)
	c.Status(http.StatusNoContent)
}

// GetApiKeysByOwner returns all API keys for a given owner
func (h *Handler) GetApiKeysByOwner(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetApiKeysByOwner] trace_id=%s - Getting API keys by owner", traceID)
	owner := c.Param("owner")
	if owner == "" {
		h.logger.Warnf("[GetApiKeysByOwner] Owner is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner is required"})
		return
	}

	ctx := context.Background()

	h.logger.Infof("[GetApiKeysByOwner] Fetching API keys for owner: %s", owner)

	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	apiKeys, err := h.apiKeysRepository.GetByField(ctx, "owner", owner)
	trackDBOp(err)
	if err != nil || len(apiKeys) == 0 {
		h.logger.Warnf("[GetApiKeysByOwner] No API keys found for owner %s: %v", owner, err)
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
}
