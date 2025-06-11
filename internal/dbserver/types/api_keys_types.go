package types

import "time"

type ApiKeyData struct {
	Key          string    `json:"key"`
	Owner        string    `json:"owner"`
	IsActive     bool      `json:"is_active"`
	SuccessCount int64     `json:"success_count"`
	FailedCount  int64     `json:"failed_count"`
	RateLimit    int       `json:"rate_limit"`
	LastUsed     time.Time `json:"last_used"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,min=3,max=50"`
	RateLimit int    `json:"rate_limit" validate:"required,min=1,max=1000"`
}

type CreateApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	RateLimit int       `json:"rate_limit"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

type GetApiKeyCallCount struct {
	Key          string `json:"key"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
}

type UpdateApiKeyRequest struct {
	Key       string `json:"key"`
	IsActive  *bool  `json:"isActive,omitempty"`
	RateLimit *int   `json:"rateLimit,omitempty"`
}

type UpdateApiKeyStatusRequest struct {
	Key      string `json:"key"`
	IsActive bool   `json:"is_active"`
}

type ApiKeyCounters struct {
	SuccessCount int64 `json:"success_count"`
	FailedCount  int64 `json:"failed_count"`
}
