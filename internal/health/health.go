package health

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       logging.Logger
	stateManager *KeeperStateManager
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, stateManager *KeeperStateManager) *Handler {
	return &Handler{
		logger:       logger,
		stateManager: stateManager,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP Request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration,
			"ip", c.ClientIP(),
		)
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine) {
	handler := NewHandler(logging.GetServiceLogger(), GetKeeperStateManager())

	router.GET("/", handler.handleRoot)
	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/status", handler.GetKeeperStatus)
	router.GET("/operators", handler.GetDetailedKeeperStatus)
}

func (h *Handler) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Health Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) HandleCheckInEvent(c *gin.Context) {
	var keeperHealth types.KeeperHealth
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if keeperHealth.Version == "0.0.7" || keeperHealth.Version == "0.0.6" || keeperHealth.Version == "0.0.5" || keeperHealth.Version == "" {
		h.logger.Debug("Obsolete version detected",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.0",
		})
		return
	}

	if keeperHealth.Version != "0.1.0" {
		h.logger.Warn("Unsupported version",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Unsupported Keeper version",
		})
		return
	}

	h.logger.Info("Keeper check-in received",
		"keeper", keeperHealth.KeeperAddress,
		"version", keeperHealth.Version,
		"ip", c.ClientIP(),
	)

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)

	if err := h.stateManager.UpdateKeeperHealth(keeperHealth); err != nil {
		h.logger.Error("Failed to update keeper state",
			"error", err,
			"keeper", keeperHealth.KeeperAddress,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Keeper health check-in received",
		"active":  true,
	})
}

func (h *Handler) GetKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	activeKeepers := h.stateManager.GetAllActiveKeepers()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":      total,
		"active_keepers":     active,
		"active_keeper_list": activeKeepers,
	})
}

func (h *Handler) GetDetailedKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	detailedInfo := h.stateManager.GetDetailedKeeperInfo()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}
