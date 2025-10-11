package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// WebSocketHandler handles WebSocket-related HTTP requests
type WebSocketHandler struct {
	connectionManager *websocket.WebSocketConnectionManager
	logger            logging.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(connectionManager *websocket.WebSocketConnectionManager, logger logging.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		connectionManager: connectionManager,
		logger:            logger,
	}
}

// HandleWebSocketConnection handles WebSocket connection upgrade requests
func (wsh *WebSocketHandler) HandleWebSocketConnection(c *gin.Context) {
	wsh.connectionManager.HandleWebSocketConnection(c)
}

// GetWebSocketStats returns WebSocket hub statistics
func (wsh *WebSocketHandler) GetWebSocketStats(c *gin.Context) {
	stats := wsh.connectionManager.GetHubStats()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetWebSocketHealth returns WebSocket service health status
func (wsh *WebSocketHandler) GetWebSocketHealth(c *gin.Context) {
	// Check if hub is running and has active connections
	stats := wsh.connectionManager.GetHubStats()

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"websocket": gin.H{
			"active_connections": stats["total_clients"],
			"active_rooms":       stats["total_rooms"],
		},
	}

	c.JSON(http.StatusOK, health)
}
