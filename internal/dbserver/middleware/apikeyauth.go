package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeyAuth struct {
	db          *database.Connection
	logger      observability.Logger
	rateLimiter *RateLimiter
}

func NewApiKeyAuth(db *database.Connection, rateLimiter *RateLimiter, logger observability.Logger) *ApiKeyAuth {
	return &ApiKeyAuth{
		db:          db,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

func (a *ApiKeyAuth) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyHeader := c.GetHeader("X-Api-Key")
		if apiKeyHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		apiKey, err := a.getApiKey(c.Request.Context(), apiKeyHeader)
		if err != nil {
			a.logger.Error(c.Request.Context(), "Error retrieving API key", observability.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive API key"})
			c.Abort()
			return
		}

		if !apiKey.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key is inactive"})
			c.Abort()
			return
		}

		go a.updateLastUsed(apiKeyHeader)

		if a.rateLimiter != nil {
			if err := a.rateLimiter.ApplyGinRateLimit(c, apiKey); err != nil {
				a.logger.Warn(c.Request.Context(), "Rate limit applied", observability.Error(err))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Rate limit exceeded",
					"message": "You have exceeded the rate limit",
				})
				c.Abort()
				return
			}
		}

		c.Set("apiKey", apiKey)
		c.Next()
	}
}

func (a *ApiKeyAuth) KeeperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if the request has a valid API key
		apiKeyHeader := c.GetHeader("X-Api-Key")
		if apiKeyHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		apiKey, err := a.getApiKey(c.Request.Context(), apiKeyHeader)
		if err != nil {
			a.logger.Error(c.Request.Context(), "Error retrieving API key", observability.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive API key"})
			c.Abort()
			return
		}

		if !apiKey.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key is inactive"})
			c.Abort()
			return
		}

		// Check if the API key belongs to a keeper
		isKeeper, err := a.isKeeperApiKey(apiKey.Key)
		if err != nil {
			a.logger.Error(c.Request.Context(), "Error checking keeper status", observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !isKeeper {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Keeper authorization required"})
			c.Abort()
			return
		}

		go a.updateLastUsed(apiKeyHeader)

		if a.rateLimiter != nil {
			if err := a.rateLimiter.ApplyGinRateLimit(c, apiKey); err != nil {
				a.logger.Warn(c.Request.Context(), "Rate limit applied", observability.Error(err))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Rate limit exceeded",
					"message": "You have exceeded the rate limit",
				})
				c.Abort()
				return
			}
		}

		c.Set("apiKey", apiKey)
		c.Next()
	}
}

func (a *ApiKeyAuth) getApiKey(ctx context.Context, key string) (*types.ApiKey, error) {
	query := `SELECT key, owner, is_active, rate_limit, last_used, created_at 
			  FROM triggerx.apikeys WHERE key = ? AND is_active = ? ALLOW FILTERING`

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
		a.logger.Error(ctx, "Failed to retrieve API key for key", observability.String("key", key), observability.Error(err))
		return nil, err
	}

	return &apiKey, nil
}

func (a *ApiKeyAuth) updateLastUsed(key string) {
	query := `UPDATE triggerx.apikeys SET last_used = ? WHERE key = ?`

	if err := a.db.Session().Query(query, time.Now().UTC(), key).Exec(); err != nil {
		a.logger.Error(context.Background(), "Failed to update last used timestamp", observability.Error(err))
	}
}

func (a *ApiKeyAuth) isKeeperApiKey(key string) (bool, error) {
	query := `SELECT isKeeper FROM triggerx.apikeys WHERE key = ? ALLOW FILTERING`

	var isKeeper bool
	err := a.db.Session().Query(query, key).Scan(&isKeeper)
	if err != nil {
		return false, err
	}

	return isKeeper, nil
}

// Public wrapper methods for WebSocket authentication

// GetApiKey validates and retrieves API key data (public wrapper for getApiKey)
func (a *ApiKeyAuth) GetApiKey(ctx context.Context, key string) (*types.ApiKey, error) {
	return a.getApiKey(ctx, key)
}

// UpdateLastUsed updates the last used timestamp for an API key (public wrapper for updateLastUsed)
func (a *ApiKeyAuth) UpdateLastUsed(key string) {
	a.updateLastUsed(key)
}
