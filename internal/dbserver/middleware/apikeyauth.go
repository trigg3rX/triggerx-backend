package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func (a *ApiKeyAuth) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyHeader := c.GetHeader("X-Api-Key")
		if apiKeyHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		apiKey, err := a.getApiKey(apiKeyHeader)
		if err != nil {
			a.logger.Errorf("Error retrieving API key: %v", err)
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
				a.logger.Warnf("Rate limit applied: %v", err)
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
