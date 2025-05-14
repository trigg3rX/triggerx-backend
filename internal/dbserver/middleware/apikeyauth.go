package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeyAuth struct {
	db          *database.Connection
	logger      logging.Logger
	rateLimiter *RateLimiter
}

func NewApiKeyAuth(db *database.Connection, rateLimiter *RateLimiter, logger logging.Logger) *ApiKeyAuth {
	return &ApiKeyAuth{
		db:          db,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

func (a *ApiKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader := r.Header.Get("X-Api-Key")
		if apiKeyHeader == "" {
			http.Error(w, "API key is required", http.StatusUnauthorized)
			return
		}

		apiKey, err := a.getApiKey(apiKeyHeader)
		if err != nil {
			a.logger.Errorf("Error retrieving API key: %v", err)
			http.Error(w, "Invalid or inactive API key", http.StatusForbidden)
			return
		}

		if !apiKey.IsActive {
			http.Error(w, "API key is inactive", http.StatusForbidden)
			return
		}

		go a.updateLastUsed(apiKeyHeader)

		if a.rateLimiter != nil {
			resp, err := a.rateLimiter.ApplyRateLimit(r, apiKey)
			if err != nil {
				a.logger.Warnf("Rate limit applied: %v", err)

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

		ctx := context.WithValue(r.Context(), "apiKey", apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *ApiKeyAuth) getApiKey(key string) (*types.ApiKey, error) {
	query := `SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
			  FROM triggerx.apikeys WHERE key = ? AND isActive = ? ALLOW FILTERING`

	var apiKey types.ApiKey

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

func (a *ApiKeyAuth) updateLastUsed(key string) {
	query := `UPDATE triggerx.apikeys SET lastUsed = ? WHERE key = ? ALLOW FILTERING`

	if err := a.db.Session().Query(query, time.Now().UTC(), key).Exec(); err != nil {
		a.logger.Errorf("Failed to update last used timestamp: %v", err)
	}
}
