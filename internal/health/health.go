package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	// "strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger        logging.Logger
	stateManager  *keeper.StateManager
	healthMetrics *HealthMetrics
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, stateManager *keeper.StateManager, healthMetrics *HealthMetrics) *Handler {
	return &Handler{
		logger:        logger,
		stateManager:  stateManager,
		healthMetrics: healthMetrics,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger, healthMetrics *HealthMetrics) gin.HandlerFunc {
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
		healthMetrics.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		healthMetrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

		middlewareLogger.Debug("HTTP Request",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, logger logging.Logger) {
	// Initialize metrics collector
	metricsCollector := metrics.NewCollector("health")
	metricsCollector.Start()

	// Initialize health-specific metrics
	healthMetrics := NewHealthMetrics(metricsCollector)

	// Apply logger middleware with metrics
	router.Use(LoggerMiddleware(logger, healthMetrics))

	// Create handler with metrics
	handler := NewHandler(logger, keeper.GetStateManager(), healthMetrics)

	router.GET("/", handler.handleRoot)
	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/status", handler.GetKeeperStatus)
	router.GET("/operators", handler.GetDetailedKeeperStatus)
	router.GET("/performers", handler.GetActivePerformers) // New endpoint for taskmanager
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
	var keeperHealth types.HealthKeeperCheckInRequest
	var response types.HealthKeeperCheckInResponse
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error("Failed to parse keeper health check-in request",
			"error", err,
		)
		response.Status = false
		response.Data = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Handle missing fields with defaults
	if keeperHealth.Version == "" {
		keeperHealth.Version = "0.1.0"
	}

	// h.logger.Debug("Received keeper health check-in",
	// 	"keeper", keeperHealth.KeeperAddress,
	// 	"version", keeperHealth.Version,
	// 	"peer_id", keeperHealth.PeerID,
	// )

	// Record check-in by version metric
	h.healthMetrics.CheckinsByVersionTotal.WithLabelValues(keeperHealth.Version).Inc()

	// Verify signature for all versions
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

	// h.logger.Debug("Valid keeper signature verified",
	// 	"keeper", keeperHealth.KeeperAddress,
	// 	"version", keeperHealth.Version,
	// 	"ip", c.ClientIP(),
	// )

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
	keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

	// Convert request to HealthKeeperInfo for state manager
	keeperInfo := types.HealthKeeperInfo{
		KeeperAddress:    keeperHealth.KeeperAddress,
		ConsensusAddress: keeperHealth.ConsensusAddress,
		Version:          keeperHealth.Version,
		IsImua:           keeperHealth.IsImua,
	}

	// Update keeper state for all versions
	if err := h.stateManager.UpdateKeeperStatus(context.Background(), keeperInfo); err != nil {
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

	h.logger.Infof("CheckIn Successful: %s | %s", keeperHealth.KeeperAddress, keeperHealth.Version)

	// Handle different versions according to requirements
	switch keeperHealth.Version {
	case "0.1.6", "0.2.0", "0.2.1", "0.2.2", "1.0.0", "1.0.1":
		// Latest version - return msgData with no warning
		var message string
		if keeperHealth.IsImua {
			message = fmt.Sprintf("%s:%s:%s:%s:%s:%s",
				config.GetEtherscanAPIKey(),
				config.GetAlchemyAPIKey(),
				config.GetPinataHost(),
				config.GetPinataJWT(),
				config.GetManagerSigningAddress(),
				config.GetImuaTaskExecutionAddress(),
			)
		} else {
			message = fmt.Sprintf("%s:%s:%s:%s:%s:%s",
				config.GetEtherscanAPIKey(),
				config.GetAlchemyAPIKey(),
				config.GetPinataHost(),
				config.GetPinataJWT(),
				config.GetManagerSigningAddress(),
				config.GetTaskExecutionAddress(),
			)
		}
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

	case "0.1.5", "0.1.4", "0.1.3":
		// Old versions that can handle msgData - return msgData with warning
		// h.logger.Warn("Keeper using outdated version, recommend upgrade to latest",
		// 	"keeper", keeperHealth.KeeperAddress,
		// 	"version", keeperHealth.Version,
		// 	"recommended_version", "0.1.6",
		// )

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

	default:
		// Oldest versions (0.1.0-0.1.2) - return warning only, no msgData
		// h.logger.Warn("Keeper using very outdated version, recommend upgrade to latest",
		// 	"keeper", keeperHealth.KeeperAddress,
		// 	"version", keeperHealth.Version,
		// 	"recommended_version", "0.1.6",
		// )

		response.Status = true
		response.Data = "UPGRADE TO v0.1.6 for full functionality"
		c.JSON(http.StatusOK, response)
	}
}

func (h *Handler) GetKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	activeKeepers := h.stateManager.GetAllActiveKeepers()

	// Update keeper metrics
	h.healthMetrics.KeepersTotal.Set(float64(total))
	h.healthMetrics.KeepersActiveTotal.Set(float64(active))
	h.healthMetrics.KeepersInactiveTotal.Set(float64(total - active))

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
	h.healthMetrics.KeepersTotal.Set(float64(total))
	h.healthMetrics.KeepersActiveTotal.Set(float64(active))
	h.healthMetrics.KeepersInactiveTotal.Set(float64(total - active))

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

// GetActivePerformers returns active performers for taskmanager service
func (h *Handler) GetActivePerformers(c *gin.Context) {
	// detailedKeepers := h.stateManager.GetDetailedKeeperInfo()

	// Filter for active keepers only and convert to performer format
	// performers := make([]map[string]interface{}, 0)
	// for _, keeper := range detailedKeepers {
	// 	if keeper.IsActive {
	// 		// Parse operator_id from string to int64
	// 		operatorID, err := strconv.ParseInt(keeper.OperatorID, 10, 64)
	// 		if err != nil {
	// 			h.logger.Error("Failed to parse operator_id", "operator_id", keeper.OperatorID, "error", err)
	// 			continue // Skip this keeper if we can't parse the operator_id
	// 		}

	// 		performer := map[string]interface{}{
	// 			"operator_id":    operatorID,
	// 			"keeper_address": keeper.KeeperAddress,
	// 			"is_imua":        keeper.IsImua,
	// 			"last_seen":      keeper.LastCheckedIn,
	// 		}
	// 		performers = append(performers, performer)
	// 	}
	// }

	// Temporary fix for taskmanager
	fallbackPerformers := []types.PerformerData{
		{
			OperatorID:    4,
			KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
			IsImua:        false,
		},
		{
			OperatorID:    1,
			KeeperAddress: "0xcacce39134e3b9d5d9220d87fc546c6f0fb9cc37",
			IsImua:        true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"performers": fallbackPerformers,
		"count":      len(fallbackPerformers),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}
