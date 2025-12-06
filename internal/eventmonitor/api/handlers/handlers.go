package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/registry"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// HandleHealth handles health check requests
func HandleHealth(logger logging.Logger, rm *registry.RegistryManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.HealthResponse{
			Status:          "healthy",
			Version:         "0.1.0-mvp",
			ActiveMonitors:  rm.GetActiveMonitorCount(),
			ChainsSupported: rm.GetChainsSupported(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// HandleRegister handles register monitoring request
func HandleRegister(logger logging.Logger, svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.MonitoringRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Invalid register request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Validate expiration time
		if req.ExpiresAt.Before(time.Now()) {
			logger.Warn("Invalid expiration time", "expires_at", req.ExpiresAt)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "expires_at must be in the future",
			})
			return
		}

		// Register the request (service handles worker creation)
		if err := svc.Register(&req); err != nil {
			logger.Error("Failed to register monitoring request", "error", err, "request_id", req.RequestID)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		response := types.RegisterResponse{
			Success:   true,
			RequestID: req.RequestID,
			Status:    "registered",
			Message:   "Monitoring request registered successfully",
		}

		logger.Info("Monitoring request registered", "request_id", req.RequestID)
		c.JSON(http.StatusOK, response)
	}
}

// HandleUnregister handles unregister monitoring request
func HandleUnregister(logger logging.Logger, svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.UnregisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Invalid unregister request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Unregister the request (service handles worker cleanup)
		if err := svc.Unregister(req.RequestID); err != nil {
			logger.Warn("Failed to unregister monitoring request", "error", err, "request_id", req.RequestID)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		response := types.UnregisterResponse{
			Success:   true,
			RequestID: req.RequestID,
			Status:    "unregistered",
			Message:   "Monitoring request unregistered successfully",
		}

		logger.Info("Monitoring request unregistered", "request_id", req.RequestID)
		c.JSON(http.StatusOK, response)
	}
}

// HandleStatus handles status check request
func HandleStatus(logger logging.Logger, rm *registry.RegistryManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Param("request_id")
		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "request_id is required",
			})
			return
		}

		// Get entry by request ID
		entry, _, exists := rm.GetEntryByRequestID(requestID)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "request_id not found",
			})
			return
		}

		entry.Mu.RLock()
		subscriber, exists := entry.Subscribers[requestID]
		if !exists {
			entry.Mu.RUnlock()
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "subscriber not found",
			})
			return
		}

		response := types.StatusResponse{
			RequestID:          requestID,
			Status:             "active",
			ChainID:            entry.ChainID,
			ContractAddress:    entry.ContractAddr.Hex(),
			LastBlockProcessed: entry.LastBlock,
			EventsFound:        0, // TODO: Track event count if needed
			ExpiresAt:          subscriber.ExpiresAt,
		}
		entry.Mu.RUnlock()

		c.JSON(http.StatusOK, response)
	}
}
