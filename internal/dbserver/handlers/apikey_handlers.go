package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"isActive"`
	RateLimit int       `json:"rateLimit"`
	LastUsed  time.Time `json:"lastUsed"`
	CreatedAt time.Time `json:"createdAt"`
}

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

	var existingKey types.ApiKey

	err := h.db.Session().Query(queries.CheckApiKeyQuery, req.Owner).Scan(
		&existingKey.Key,
		&existingKey.Owner,
		&existingKey.IsActive,
		&existingKey.RateLimit,
		&existingKey.LastUsed,
		&existingKey.CreatedAt,
	)

	if err == nil {
		h.logger.Warnf("[CreateApiKey] API key already exists for owner %s", req.Owner)
		c.JSON(http.StatusConflict, gin.H{"error": "API key already exists for this owner"})
		return
	} else if err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateApiKey] Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	apiKey := types.ApiKey{
		Key:       uuid.New().String(),
		Owner:     req.Owner,
		IsActive:  true,
		RateLimit: req.RateLimit,
		LastUsed:  time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	if err := h.db.Session().Query(queries.InsertApiKeyQuery,
		apiKey.Key,
		apiKey.Owner,
		apiKey.IsActive,
		apiKey.RateLimit,
		apiKey.LastUsed,
		apiKey.CreatedAt,
	).Exec(); err != nil {
		h.logger.Errorf("[CreateApiKey] Failed to insert API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	h.logger.Infof("[CreateApiKey] Successfully created new API key for owner %s (Key: %s)", req.Owner, apiKey.Key)
	c.JSON(http.StatusCreated, apiKey)
}

func (h *Handler) UpdateApiKey(c *gin.Context) {
	keyID := c.Param("key")

	var req struct {
		IsActive  *bool `json:"isActive,omitempty"`
		RateLimit *int  `json:"rateLimit,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var apiKey types.ApiKey

	if err := h.db.Session().Query(queries.SelectApiKeyQuery, keyID).Scan(
		&apiKey.Key,
		&apiKey.Owner,
		&apiKey.IsActive,
		&apiKey.RateLimit,
		&apiKey.LastUsed,
		&apiKey.CreatedAt,
	); err != nil {
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

	if err := h.db.Session().Query(queries.UpdateApiKeyQuery,
		apiKey.IsActive,
		apiKey.RateLimit,
		apiKey.Key,
	).Exec(); err != nil {
		h.logger.Errorf("Failed to update API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

func (h *Handler) DeleteApiKey(c *gin.Context) {
	keyID := c.Param("key")

	if err := h.db.Session().Query(queries.DeleteApiKeyQuery, false, keyID).Exec(); err != nil {
		h.logger.Errorf("Failed to delete API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	c.Status(http.StatusNoContent)
}
