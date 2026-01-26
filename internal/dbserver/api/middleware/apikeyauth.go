package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type apiKeyResponseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (rw *apiKeyResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

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

		if a.rateLimiter != nil {
			if err := a.rateLimiter.ApplyGinRateLimit(c, apiKey); err != nil {
				a.logger.Warn(c.Request.Context(), "Rate limit applied", observability.Error(err))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Rate limit exceeded",
					"message": "You have exceeded the rate limit",
				})
				c.Abort()
				// Track as failure
				go a.updateApiKeyUsage(apiKeyHeader, false)
				return
			}
		}

		// Wrap the response writer to track status code
		rw := &apiKeyResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}
		c.Writer = rw

		c.Set("apiKey", apiKey)
		c.Next()

		// Determine success based on status code (2xx = success, others = failure)
		isSuccess := rw.statusCode >= 200 && rw.statusCode < 300
		go a.updateApiKeyUsage(apiKeyHeader, isSuccess)
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

		if a.rateLimiter != nil {
			if err := a.rateLimiter.ApplyGinRateLimit(c, apiKey); err != nil {
				a.logger.Warn(c.Request.Context(), "Rate limit applied", observability.Error(err))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Rate limit exceeded",
					"message": "You have exceeded the rate limit",
				})
				c.Abort()
				// Track as failure
				go a.updateApiKeyUsage(apiKeyHeader, false)
				return
			}
		}

		// Wrap the response writer to track status code
		rw := &apiKeyResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}
		c.Writer = rw

		c.Set("apiKey", apiKey)
		c.Next()

		// Determine success based on status code (2xx = success, others = failure)
		isSuccess := rw.statusCode >= 200 && rw.statusCode < 300
		go a.updateApiKeyUsage(apiKeyHeader, isSuccess)
	}
}

func (a *ApiKeyAuth) getApiKey(ctx context.Context, key string) (*types.ApiKeyDataDTO, error) {
	query := `SELECT key, owner, is_active, rate_limit, last_used, created_at 
			  FROM triggerx.apikeys WHERE key = ? AND is_active = ? ALLOW FILTERING`

	var apiKey types.ApiKeyDataDTO

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

func (a *ApiKeyAuth) updateApiKeyUsage(key string, isSuccess bool) {
	// Get current counters
	var successCount int64
	var failedCount int64
	err := a.db.Session().Query(
		`SELECT success_count, failed_count FROM triggerx.apikeys WHERE key = ?`,
		key,
	).Scan(&successCount, &failedCount)
	
	if err != nil {
		if err == gocql.ErrNotFound {
			// API key not found, skip update
			return
		}
		a.logger.Error(context.Background(), "Failed to get API key counters", observability.Error(err))
		return
	}

	// Increment appropriate counter
	if isSuccess {
		successCount++
	} else {
		failedCount++
	}

	// Update with new counters and timestamp
	err = a.db.Session().Query(
		`UPDATE triggerx.apikeys SET last_used = ?, success_count = ?, failed_count = ? WHERE key = ?`,
		time.Now().UTC(), successCount, failedCount, key,
	).Exec()
	
	if err != nil {
		a.logger.Error(context.Background(), "Failed to update API key usage", observability.Error(err))
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
func (a *ApiKeyAuth) GetApiKey(ctx context.Context, key string) (*types.ApiKeyDataDTO, error) {
	return a.getApiKey(ctx, key)
}

// UpdateApiKeyUsage updates the API key usage (public wrapper for updateApiKeyUsage)
func (a *ApiKeyAuth) UpdateApiKeyUsage(key string, isSuccess bool) {
	a.updateApiKeyUsage(key, true)
}
