package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Owner == "" {
		http.Error(w, "Owner is required", http.StatusBadRequest)
		return
	}

	if req.RateLimit <= 0 {
		req.RateLimit = 60 // Default rate limit: 60 requests per minute
	}

	// Generate a new API key
	apiKey := &types.ApiKey{
		Key:       uuid.New().String(),
		Owner:     req.Owner,
		IsActive:  true,
		RateLimit: req.RateLimit,
		LastUsed:  time.Time{}, // Zero time
		CreatedAt: time.Now(),
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
		h.logger.Errorf("Failed to create API key: %v", err)
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

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
