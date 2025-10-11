package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ApiKeyAuth struct {
	apiKeyRepo  interfaces.GenericRepository[types.ApiKeyDataEntity]
	logger      logging.Logger
	rateLimiter *RateLimiter
}

func NewApiKeyAuth(apiKeyRepo interfaces.GenericRepository[types.ApiKeyDataEntity], rateLimiter *RateLimiter, logger logging.Logger) *ApiKeyAuth {
	return &ApiKeyAuth{
		apiKeyRepo:  apiKeyRepo,
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

func (a *ApiKeyAuth) KeeperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if the request has a valid API key
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

		// Check if the API key belongs to a keeper
		// isKeeper, err := a.isKeeperApiKey(apiKey.Key)
		// if err != nil {
		// 	a.logger.Errorf("Error checking keeper status: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		// 	c.Abort()
		// 	return
		// }

		// if !isKeeper {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Keeper authorization required"})
		// 	c.Abort()
		// 	return
		// }

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

func (a *ApiKeyAuth) getApiKey(key string) (*types.ApiKeyDataEntity, error) {
	ctx := context.Background()

	// Use repository to get API key by key field
	apiKey, err := a.apiKeyRepo.GetByNonID(ctx, "key", key)
	if err != nil {
		a.logger.Errorf("Failed to retrieve API key for key %s: %v", key, err)
		return nil, err
	}

	// Check if the API key is active
	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	return apiKey, nil
}

func (a *ApiKeyAuth) updateLastUsed(key string) {
	ctx := context.Background()

	// First get the current API key
	apiKey, err := a.apiKeyRepo.GetByNonID(ctx, "key", key)
	if err != nil {
		a.logger.Errorf("Failed to retrieve API key for update: %v", err)
		return
	}

	// Update the last used timestamp
	apiKey.LastUsed = time.Now().UTC()

	// Use repository to update the record
	if err := a.apiKeyRepo.Update(ctx, apiKey); err != nil {
		a.logger.Errorf("Failed to update last used timestamp: %v", err)
	}
}

// func (a *ApiKeyAuth) isKeeperApiKey(key string) (bool, error) {
// 	ctx := context.Background()

// 	// Use repository to get API key by key field
// 	apiKey, err := a.apiKeyRepo.GetByNonID(ctx, "key", key)
// 	if err != nil {
// 		return false, err
// 	}

// 	return apiKey.IsKeeper, nil
// }

// Public wrapper methods for WebSocket authentication

// GetApiKey validates and retrieves API key data (public wrapper for getApiKey)
func (a *ApiKeyAuth) GetApiKey(key string) (*types.ApiKeyDataEntity, error) {
	return a.getApiKey(key)
}

// UpdateLastUsed updates the last used timestamp for an API key (public wrapper for updateLastUsed)
func (a *ApiKeyAuth) UpdateLastUsed(key string) {
	a.updateLastUsed(key)
}
