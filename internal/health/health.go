package health

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       logging.Logger
	stateManager *keeper.StateManager
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, stateManager *keeper.StateManager) *Handler {
	return &Handler{
		logger:       logger,
		stateManager: stateManager,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		// Skip logging for metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Record HTTP metrics
		statusCode := fmt.Sprintf("%d", status)
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

		if status >= 500 {
			middlewareLogger.Error("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else if status >= 400 {
			middlewareLogger.Warn("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else {
			middlewareLogger.Info("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		}
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, logger logging.Logger) {
	handler := NewHandler(logger, keeper.GetStateManager())

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector()
	metricsCollector.Start()

	router.GET("/", handler.handleRoot)
	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/status", handler.GetKeeperStatus)
	router.GET("/operators", handler.GetDetailedKeeperStatus)
	router.GET("/metrics", gin.WrapH(metricsCollector.Handler()))
}

func (h *Handler) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Health Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) HandleCheckInEvent(c *gin.Context) {
	var keeperHealth commonTypes.KeeperHealthCheckIn
	var response commonTypes.KeeperHealthCheckInResponse
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error("Failed to parse keeper health check-in request",
			"error", err,
		)
		response.Status = false
		response.Data = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	h.logger.Debug("Received keeper health check-in",
		"keeper", keeperHealth.KeeperAddress,
		"version", keeperHealth.Version,
		"peer_id", keeperHealth.PeerID,
	)

	// Record check-in by version metric
	metrics.CheckinsByVersionTotal.WithLabelValues(keeperHealth.Version).Inc()

	if keeperHealth.Version == "0.1.3" {
		ok, err := cryptography.VerifySignature(keeperHealth.KeeperAddress, keeperHealth.Signature, keeperHealth.ConsensusAddress)
		if !ok {
			h.logger.Error("Invalid keeper signature",
				"keeper", keeperHealth.KeeperAddress,
				"error", err,
			)
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error": "Invalid signature",
			})
			return
		}

		h.logger.Debug("Valid keeper signature verified",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
			"ip", c.ClientIP(),
		)

		keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
		keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

		if err := h.stateManager.UpdateKeeperHealth(keeperHealth); err != nil {
			if errors.Is(err, keeper.ErrKeeperNotVerified) {
				h.logger.Warn("Unverified keeper attempted health check-in",
					"keeper", keeperHealth.KeeperAddress,
				)
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Keeper not verified",
					"code":  "KEEPER_NOT_VERIFIED",
				})
				return
			}

			h.logger.Error("Failed to update keeper state",
				"error", err,
				"keeper", keeperHealth.KeeperAddress,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
			return
		}

		h.logger.Info("Successfully processed keeper health check-in",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)

		message := fmt.Sprintf("%s:%s:%s:%s", config.GetEtherscanAPIKey(), config.GetAlchemyAPIKey(), config.GetPinataHost(), config.GetPinataJWT())
		msgData, err := cryptography.EncryptMessage(keeperHealth.ConsensusPubKey, message)
		if err != nil {
			h.logger.Error("Failed to encrypt message for keeper",
				"error", err,
			)
			response.Status = false
			response.Data = err.Error()
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		response.Status = true
		response.Data = msgData

		c.JSON(http.StatusOK, response)
	} else {
		h.logger.Warn("Rejecting obsolete keeper version",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		response.Status = false
		response.Data = "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.3"
		c.JSON(http.StatusPreconditionFailed, response)
		return
	}
}

func (h *Handler) GetKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	activeKeepers := h.stateManager.GetAllActiveKeepers()

	// Update keeper metrics
	metrics.KeepersTotal.Set(float64(total))
	metrics.KeepersActiveTotal.Set(float64(active))
	metrics.KeepersInactiveTotal.Set(float64(total - active))

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":      total,
		"active_keepers":     active,
		"active_keeper_list": activeKeepers,
	})
}

func (h *Handler) GetDetailedKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	detailedInfo := h.stateManager.GetDetailedKeeperInfo()

	// Update keeper metrics
	metrics.KeepersTotal.Set(float64(total))
	metrics.KeepersActiveTotal.Set(float64(active))
	metrics.KeepersInactiveTotal.Set(float64(total - active))

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}
