package types

import "time"

// CreateApiKeyRequest is the request to create a new API key
type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,eth_addr"`
	RateLimit int    `json:"rate_limit" validate:"required,min=1,max=1000"`
}

// CreateApiKeyResponse is the response to create a new API key
type CreateApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

// DeleteApiKeyRequest is the request to delete an API key
type DeleteApiKeyRequest struct {
	Key string `json:"key" validate:"required"`
	Owner string `json:"owner" validate:"required,eth_addr"`
}

// GetApiKeyDataResponse is the response to get the data of an API key
type GetApiKeyDataResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}
