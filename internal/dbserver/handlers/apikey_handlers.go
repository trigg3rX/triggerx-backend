package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ApiKeyResponse represents an API key response
type ApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"isActive"`
	RateLimit int       `json:"rateLimit"`
	LastUsed  time.Time `json:"lastUsed"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateApiKey creates a new API key
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	var req types.CreateApiKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("[CreateApiKey] Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Owner == "" {
		h.logger.Warnf("[CreateApiKey] Validation failed: Owner is required")
		http.Error(w, "Owner is required", http.StatusBadRequest)
		return
	}

	if req.RateLimit <= 0 {
		h.logger.Infof("[CreateApiKey] RateLimit not provided or invalid for owner %s, defaulting to 60", req.Owner)
		req.RateLimit = 60 // Default rate limit: 60 requests per minute
	}

	h.logger.Infof("[CreateApiKey] Checking for existing API key for owner: %s", req.Owner)

	// First check if user already has an API key
	var existingKey types.ApiKey
	checkQuery := `SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
				  FROM triggerx.apikeys WHERE owner = ? ALLOW FILTERING`

	err := h.db.Session().Query(checkQuery, req.Owner).Scan(
		&existingKey.Key,
		&existingKey.Owner,
		&existingKey.IsActive,
		&existingKey.RateLimit,
		&existingKey.LastUsed,
		&existingKey.CreatedAt,
	)

	if err == nil {
		h.logger.Infof("[CreateApiKey] Existing API key found for owner %s (Key: %s). Proceeding with update.", req.Owner, existingKey.Key)

		// User already has an API key, update it
		updateQuery := `UPDATE triggerx.apikeys 
					   SET isActive = ?, rateLimit = ?, lastUsed = ? 
					   WHERE key = ?`

		if err := h.db.Session().Query(updateQuery,
			true,
			req.RateLimit,
			time.Time{},
			existingKey.Key,
		).Exec(); err != nil {
			h.logger.Errorf("[CreateApiKey] Failed to update existing API key for owner %s (Key: %s): %v", req.Owner, existingKey.Key, err)
			http.Error(w, "Failed to update API key", http.StatusInternalServerError)
			return
		}

		existingKey.IsActive = true
		existingKey.RateLimit = req.RateLimit
		existingKey.LastUsed = time.Time{}

		h.logger.Infof("[CreateApiKey] Successfully updated API key for owner %s (Key: %s)", req.Owner, existingKey.Key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingKey)
		return
	} else if err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateApiKey] Error checking for existing API key for owner %s: %v", req.Owner, err)
		http.Error(w, "Failed to check for existing API key", http.StatusInternalServerError)
		return
	}

	// If no existing key found, create a new one
	apiKey := &types.ApiKey{
		Key:       "trgX_" + uuid.New().String(),
		Owner:     req.Owner,
		IsActive:  true,
		RateLimit: req.RateLimit,
		LastUsed:  time.Time{}, // Zero time
		CreatedAt: time.Now().UTC(),
	}

	// Save to database using CQL
	query := `INSERT INTO triggerx.apikeys (key, owner, isActive, rateLimit, lastUsed, createdAt) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	if err := h.db.Session().Query(query,
		apiKey.Key,
		apiKey.Owner,
		apiKey.IsActive,
		apiKey.RateLimit,
		apiKey.LastUsed,
		apiKey.CreatedAt,
	).Exec(); err != nil {
		h.logger.Errorf("[CreateApiKey] Failed to create API key for owner %s: %v", req.Owner, err)
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateApiKey] Successfully created new API key for owner %s (Key: %s)", req.Owner, apiKey.Key)

	// Return the newly created API key
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiKey)
}

// UpdateApiKey updates an existing API key
func (h *Handler) UpdateApiKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["key"]

	var req struct {
		IsActive  *bool `json:"isActive,omitempty"`
		RateLimit *int  `json:"rateLimit,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the current API key
	var apiKey types.ApiKey
	query := `SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
	          FROM triggerx.apikeys WHERE key = ?`

	if err := h.db.Session().Query(query, keyID).Scan(
		&apiKey.Key,
		&apiKey.Owner,
		&apiKey.IsActive,
		&apiKey.RateLimit,
		&apiKey.LastUsed,
		&apiKey.CreatedAt,
	); err != nil {
		h.logger.Errorf("API key not found: %v", err)
		http.Error(w, "API key not found", http.StatusNotFound)
		return
	}

	// Update fields if provided
	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}

	if req.RateLimit != nil && *req.RateLimit > 0 {
		apiKey.RateLimit = *req.RateLimit
	}

	// Save the updated API key using CQL
	updateQuery := `UPDATE triggerx.apikeys SET isActive = ?, rateLimit = ? WHERE key = ?`
	if err := h.db.Session().Query(updateQuery,
		apiKey.IsActive,
		apiKey.RateLimit,
		apiKey.Key,
	).Exec(); err != nil {
		h.logger.Errorf("Failed to update API key: %v", err)
		http.Error(w, "Failed to update API key", http.StatusInternalServerError)
		return
	}

	// Return the updated API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKey)
}

// DeleteApiKey deactivates an API key
func (h *Handler) DeleteApiKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["key"]

	// Mark the API key as inactive (soft delete) using CQL
	query := `UPDATE apikeys SET isActive = ? WHERE key = ?`
	if err := h.db.Session().Query(query, false, keyID).Exec(); err != nil {
		h.logger.Errorf("Failed to delete API key: %v", err)
		http.Error(w, "Failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
