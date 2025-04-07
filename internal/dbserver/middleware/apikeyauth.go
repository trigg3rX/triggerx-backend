package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ApiKeyAuth handles API key authentication
type ApiKeyAuth struct {
	db          *database.Connection
	logger      logging.Logger
	rateLimiter *RateLimiter
}

// NewApiKeyAuth creates a new API key authentication middleware
func NewApiKeyAuth(db *database.Connection, rateLimiter *RateLimiter, logger logging.Logger) *ApiKeyAuth {
	return &ApiKeyAuth{
		db:          db,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

// Middleware validates API keys and applies rate limiting
func (a *ApiKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKeyHeader := r.Header.Get("X-Api-Key")
		if apiKeyHeader == "" {
			http.Error(w, "API key is required", http.StatusUnauthorized)
			return
		}

		// Get API key from database
		apiKey, err := a.getApiKey(r.Context(), apiKeyHeader)
		if err != nil {
			a.logger.Errorf("Error retrieving API key: %v", err)
			http.Error(w, "Invalid or inactive API key", http.StatusForbidden)
			return
		}

		if !apiKey.IsActive {
			http.Error(w, "API key is inactive", http.StatusForbidden)
			return
		}

		// Update last used timestamp
		go a.updateLastUsed(apiKeyHeader)

		// Apply rate limiting
		if a.rateLimiter != nil {
			resp, err := a.rateLimiter.ApplyRateLimit(r, apiKey)
			if err != nil {
				a.logger.Warnf("Rate limit applied: %v", err)

				// Write headers from the rate limit response
				for key, values := range resp.Header {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}

				w.WriteHeader(resp.StatusCode)
				w.Write([]byte(`{"error":"Rate limit exceeded","message":"You have exceeded the rate limit"}`))
				return
			}
		}

		// Store API key in context for handlers
		ctx := context.WithValue(r.Context(), "apiKey", apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getApiKey retrieves an API key from the database
func (a *ApiKeyAuth) getApiKey(ctx context.Context, key string) (*types.ApiKey, error) {
	// CQL query to retrieve the API key
	query := `SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
			  FROM triggerx.apikeys WHERE key = ? AND isActive = ? ALLOW FILTERING`

	var apiKey types.ApiKey

	// Execute the query
	err := a.db.Session().Query(query, key, true).Scan(
		&apiKey.Key,
		&apiKey.Owner,
		&apiKey.IsActive,
		&apiKey.RateLimit,
		&apiKey.LastUsed,
		&apiKey.CreatedAt,
	)

	if err != nil {
		a.logger.Errorf("Failed to retrieve API key for key %s: %v", key, err)
		return nil, err
	}

	return &apiKey, nil
}

// updateLastUsed updates the last used timestamp for an API key
func (a *ApiKeyAuth) updateLastUsed(key string) {
	// CQL query to update the lastUsed field in the database
	query := `UPDATE triggerx.apikeys SET lastUsed = ? WHERE key = ? ALLOW FILTERING`

	// Execute the update query
	if err := a.db.Session().Query(query, time.Now().UTC(), key).Exec(); err != nil {
		a.logger.Errorf("Failed to update last used timestamp: %v", err)
	}
}
