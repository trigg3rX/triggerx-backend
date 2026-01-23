package types

import "time"

// CreateApiKeyRequest represents the request to create a new API key
type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,min=3,max=50"`
	RateLimit int    `json:"rate_limit" validate:"required,min=1,max=1000"`
}

// CreateApiKeyResponse represents the response after creating an API key
type CreateApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	RateLimit int       `json:"rate_limit"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

// GetApiKeyCallCount represents API key call statistics
type GetApiKeyCallCount struct {
	Key          string `json:"key"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
}

// UpdateApiKeyRequest represents the request to update an API key
type UpdateApiKeyRequest struct {
	Key       string `json:"key"`
	IsActive  *bool  `json:"isActive,omitempty"`
	RateLimit *int   `json:"rateLimit,omitempty"`
}

// ApiKeyCounters represents API key call counters
type ApiKeyCounters struct {
	SuccessCount int64 `json:"success_count"`
	FailedCount  int64 `json:"failed_count"`
}
