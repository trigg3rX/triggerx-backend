package websocket

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// WebSocketUpgrader handles WebSocket connection upgrades
type WebSocketUpgrader struct {
	upgrader websocket.Upgrader
	logger   observability.Logger
}

// NewWebSocketUpgrader creates a new WebSocket upgrader
func NewWebSocketUpgrader(logger observability.Logger) *WebSocketUpgrader {
	return &WebSocketUpgrader{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
	}
}

// UpgradeConnection upgrades HTTP connection to WebSocket
func (wsu *WebSocketUpgrader) UpgradeConnection(c *gin.Context) (*websocket.Conn, error) {
	conn, err := wsu.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		wsu.logger.Error(c.Request.Context(), "Failed to upgrade WebSocket connection", observability.Error(err))
		return nil, err
	}
	return conn, nil
}

// WebSocketAuthMiddleware handles WebSocket authentication using existing ApiKeyAuth
type WebSocketAuthMiddleware struct {
	apiKeyAuth *middleware.ApiKeyAuth
	logger     observability.Logger
}

// NewWebSocketAuthMiddleware creates a new WebSocket auth middleware
func NewWebSocketAuthMiddleware(apiKeyAuth *middleware.ApiKeyAuth, logger observability.Logger) *WebSocketAuthMiddleware {
	return &WebSocketAuthMiddleware{
		apiKeyAuth: apiKeyAuth,
		logger:     logger,
	}
}

// AuthenticateWebSocket authenticates WebSocket connection using existing ApiKeyAuth logic
func (wam *WebSocketAuthMiddleware) AuthenticateWebSocket(c *gin.Context) (string, string, error) {
	// Check for API key in query parameters first (WebSocket specific)
	apiKey := c.Query("api_key")
	if apiKey == "" {
		// Fall back to header (standard API key location)
		apiKey = c.GetHeader("X-Api-Key")
	}

	if apiKey == "" {
		wam.logger.Warn(c.Request.Context(), "WebSocket connection attempt without API key")
		return "", "", &WebSocketAuthError{
			Code:    "MISSING_API_KEY",
			Message: "API key is required for WebSocket connection",
		}
	}

	// Use existing ApiKeyAuth logic to validate the API key
	apiKeyData, err := wam.apiKeyAuth.GetApiKey(c.Request.Context(), apiKey)
	if err != nil {
		wam.logger.Error(c.Request.Context(), "Invalid API key for WebSocket connection", observability.Error(err))
		return "", "", &WebSocketAuthError{
			Code:    "INVALID_API_KEY",
			Message: "Invalid or inactive API key",
		}
	}

	if !apiKeyData.IsActive {
		wam.logger.Warn(c.Request.Context(), "Inactive API key used for WebSocket connection", observability.String("api_key", apiKey))
		return "", "", &WebSocketAuthError{
			Code:    "INACTIVE_API_KEY",
			Message: "API key is inactive",
		}
	}

	// Update last used timestamp asynchronously
	go wam.apiKeyAuth.UpdateLastUsed(apiKey)

	// Extract user ID from the API key owner
	userID := wam.extractUserIDFromApiKey(apiKeyData)

	return apiKey, userID, nil
}

// extractUserIDFromApiKey extracts user ID from API key data
func (wam *WebSocketAuthMiddleware) extractUserIDFromApiKey(apiKey *types.ApiKey) string {
	// The API key owner field contains the user information
	// This could be a user ID, email, or other identifier
	// For now, we'll use the owner field directly
	if apiKey.Owner != "" {
		return apiKey.Owner
	}

	// Fallback to a default system user
	return "system"
}

// WebSocketAuthError represents a WebSocket authentication error
type WebSocketAuthError struct {
	Code    string
	Message string
}

func (e *WebSocketAuthError) Error() string {
	return e.Message
}

// WebSocketRateLimiter handles rate limiting for WebSocket connections using existing RateLimiter
type WebSocketRateLimiter struct {
	rateLimiter *middleware.RateLimiter
	maxConns    int
	connections map[string]int
	logger      observability.Logger
}

// NewWebSocketRateLimiter creates a new WebSocket rate limiter
func NewWebSocketRateLimiter(rateLimiter *middleware.RateLimiter, maxConns int, logger observability.Logger) *WebSocketRateLimiter {
	return &WebSocketRateLimiter{
		rateLimiter: rateLimiter,
		maxConns:    maxConns,
		connections: make(map[string]int),
		logger:      logger,
	}
}

// CheckRateLimit checks if the client has exceeded rate limits
func (wrl *WebSocketRateLimiter) CheckRateLimit(ctx context.Context, clientIP string, apiKey *types.ApiKey) bool {
	// Check connection limit per IP (simple in-memory tracking)
	if wrl.connections[clientIP] >= wrl.maxConns {
		wrl.logger.Warn(ctx, "Connection limit exceeded for IP", observability.String("client_ip", clientIP))
		return false
	}

	// If we have a rate limiter and API key, use the Redis-backed rate limiting
	if wrl.rateLimiter != nil && apiKey != nil {
		if err := wrl.rateLimiter.CheckRateLimitForKey(apiKey); err != nil {
			wrl.logger.Warn(ctx, "Rate limit exceeded for API key", observability.String("api_key", apiKey.Key))
			return false
		}
	}

	wrl.connections[clientIP]++
	return true
}

// ReleaseConnection releases a connection for rate limiting
func (wrl *WebSocketRateLimiter) ReleaseConnection(clientIP string) {
	if wrl.connections[clientIP] > 0 {
		wrl.connections[clientIP]--
		if wrl.connections[clientIP] == 0 {
			delete(wrl.connections, clientIP)
		}
	}
}

// WebSocketConnectionManager manages WebSocket connections
type WebSocketConnectionManager struct {
	upgrader  *WebSocketUpgrader
	auth      *WebSocketAuthMiddleware
	rateLimit *WebSocketRateLimiter
	hub       *Hub
	logger    observability.Logger
}

// NewWebSocketConnectionManager creates a new WebSocket connection manager
func NewWebSocketConnectionManager(
	upgrader *WebSocketUpgrader,
	auth *WebSocketAuthMiddleware,
	rateLimit *WebSocketRateLimiter,
	hub *Hub,
	logger observability.Logger,
) *WebSocketConnectionManager {
	return &WebSocketConnectionManager{
		upgrader:  upgrader,
		auth:      auth,
		rateLimit: rateLimit,
		hub:       hub,
		logger:    logger,
	}
}

// HandleWebSocketConnection handles incoming WebSocket connections
func (wcm *WebSocketConnectionManager) HandleWebSocketConnection(c *gin.Context) {
	// Get client IP for rate limiting
	clientIP := c.ClientIP()

	// Authenticate the connection first to get API key data
	apiKey, userID, err := wcm.auth.AuthenticateWebSocket(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "AUTHENTICATION_FAILED",
		})
		return
	}

	// Get API key data for rate limiting
	apiKeyData, err := wcm.auth.apiKeyAuth.GetApiKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
			"code":  "INVALID_API_KEY",
		})
		return
	}

	// Check rate limit with API key data
	if !wcm.rateLimit.CheckRateLimit(c.Request.Context(), clientIP, apiKeyData) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Rate limit exceeded",
			"code":  "RATE_LIMIT_EXCEEDED",
		})
		return
	}

	// Check if the request context is still valid before attempting upgrade
	select {
	case <-c.Request.Context().Done():
		wcm.rateLimit.ReleaseConnection(clientIP)
		wcm.logger.Warn(c.Request.Context(), "Request context cancelled before WebSocket upgrade for client", observability.String("client_ip", clientIP))
		return
	default:
		// Context is still valid, proceed with upgrade
	}

	// Upgrade to WebSocket
	conn, err := wcm.upgrader.UpgradeConnection(c)
	if err != nil {
		wcm.rateLimit.ReleaseConnection(clientIP)
		// Don't try to send JSON response after failed upgrade attempt
		// as the connection may already be hijacked or closed
		wcm.logger.Error(c.Request.Context(), "Failed to upgrade WebSocket connection for client", observability.String("client_ip", clientIP), observability.Error(err))
		return
	}

	// Create client
	clientID := generateClientID()
	client := NewClient(clientID, conn, wcm.hub, wcm.logger)
	client.UserID = userID
	client.APIKey = apiKey
	// Ensure per-IP connection is released on client disconnect
	client.OnClose = func() {
		wcm.rateLimit.ReleaseConnection(clientIP)
	}

	// Register client with hub
	wcm.hub.register <- client

	// Start client pumps
	go client.WritePump(c.Request.Context())
	go client.ReadPump(c.Request.Context())

	wcm.logger.Info(c.Request.Context(), "WebSocket client connected", observability.String("client_ip", clientIP))
}

// GetHubStats returns hub statistics
func (wcm *WebSocketConnectionManager) GetHubStats() map[string]interface{} {
	return wcm.hub.GetStats()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	// Simple ID generation - in production, use UUID or similar
	return "client_" + strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "")
}
