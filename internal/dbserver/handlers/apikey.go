package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) CreateApiKey(c *gin.Context) {
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

	h.logger.Infof("[CreateApiKey] Checking for existing API key for owner: %s", req.Owner)

	trackDBOp := metrics.TrackDBOperation("read", "apikey_data")
	existingKey, err := h.apiKeysRepository.GetApiKeyDataByOwner(req.Owner)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateApiKey] Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if existingKey != nil {
		h.logger.Warnf("[CreateApiKey] API key already exists for owner %s", req.Owner)
		c.JSON(http.StatusConflict, gin.H{"error": "API key already exists for this owner"})
		return
	}

	apiKey := types.ApiKeyData{
		Key:       uuid.New().String(),
		Owner:     req.Owner,
		IsActive:  true,
		RateLimit: req.RateLimit,
		LastUsed:  time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	trackDBOp = metrics.TrackDBOperation("create", "apikey_data")
	if err := h.apiKeysRepository.CreateApiKey(&apiKey); err != nil {
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
	if err := h.apiKeysRepository.UpdateApiKey(&req); err != nil {
		trackDBOp(err)
		h.logger.Errorf("Failed to update API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, apiKey)
}

func (h *Handler) DeleteApiKey(c *gin.Context) {
	keyID := c.Param("key")

	trackDBOp := metrics.TrackDBOperation("update", "apikey_data")
	if err := h.apiKeysRepository.UpdateApiKeyStatus(&types.UpdateApiKeyStatusRequest{Key: keyID, IsActive: false}); err != nil {
		trackDBOp(err)
		h.logger.Errorf("Failed to delete API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	trackDBOp(nil)

	c.Status(http.StatusNoContent)
}
