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
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       observability.Logger
	stateManager *keeper.StateManager
}

// NewHandler creates a new instance of Handler
func NewHandler(logger observability.Logger, stateManager *keeper.StateManager) *Handler {
	return &Handler{
		logger:       logger,
		stateManager: stateManager,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger observability.Logger) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ctx := c.Request.Context()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		statusCode := fmt.Sprintf("%d", status)

		// Record HTTP metrics
		metrics.RecordHTTPRequest(ctx, method, path, statusCode, duration)

		middlewareLogger.Debug(ctx, "HTTP Request",
			observability.String("method", method),
			observability.String("path", path),
			observability.Int("status", status),
			observability.Int64("duration_ms", duration.Milliseconds()),
			observability.String("ip", c.ClientIP()),
		)
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, logger observability.Logger) {
	handler := NewHandler(logger, keeper.GetStateManager())

	// Start metrics collection (metrics should already be initialized via InitializeMetrics)
	metrics.StartMetricsCollection()

	// Service status endpoint for Pulsate and nginx
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "health",
			"version":   config.GetVersion(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/operators", handler.GetDetailedKeeperStatus)
}

func (h *Handler) HandleCheckInEvent(c *gin.Context) {
	ctx := c.Request.Context()
	var keeperHealth types.KeeperHealthCheckInRequest
	var response types.KeeperHealthCheckInResponse
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error(ctx, "Failed to parse keeper health check-in request",
			observability.Error(err),
		)
		response.Status = false
		response.Data = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Handle missing fields with defaults
	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}
	if keeperHealth.Version == "" {
		keeperHealth.Version = "0.1.0"
	}
	if keeperHealth.Network == "" {
		keeperHealth.Network = "mainnet"
	}

	// Verify signature for all versions
	ok, _ := cryptography.VerifySignature(keeperHealth.KeeperAddress, keeperHealth.Signature, keeperHealth.ConsensusAddress)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Invalid signature",
		})
		return
	}

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
	keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

	// Update keeper state for all versions
	if err := h.stateManager.UpdateKeeperHealth(ctx, keeperHealth); err != nil {
		if errors.Is(err, keeper.ErrKeeperNotVerified) {
			h.logger.Warn(ctx, "Unverified keeper attempted health check-in",
				observability.String("keeper", keeperHealth.KeeperAddress),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Keeper not verified",
				"code":  "KEEPER_NOT_VERIFIED",
			})
			return
		}

		h.logger.Error(ctx, "Failed to update keeper state",
			observability.Error(err),
			observability.String("keeper", keeperHealth.KeeperAddress),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
		return
	}

	// Update keeper counts metrics after successful check-in
	total, active := h.stateManager.GetKeeperCount(ctx)
	metrics.UpdateKeeperCounts(ctx, total, active)

	// Update keepers online by version metric
	keepersByVersion := h.stateManager.GetKeepersByVersion(ctx)
	metrics.UpdateKeepersOnlineByVersion(ctx, keepersByVersion)

	h.logger.Debug(ctx, "CheckIn Successful",
		observability.String("keeper", keeperHealth.KeeperAddress),
		observability.String("version", keeperHealth.Version),
		observability.String("network", keeperHealth.Network),
	)

	// All versions are allowed to check-in and receive encrypted data
	// Use network field to decide which task execution address to use
	var taskExecutionAddress string
	switch strings.ToLower(keeperHealth.Network) {
	case "imua":
		taskExecutionAddress = config.GetImuaTaskExecutionAddress()
	case "mainnet":
		taskExecutionAddress = config.GetTaskExecutionAddress()
	case "sepolia":
		taskExecutionAddress = config.GetTestTaskExecutionAddress()
	default:
		// Fallback to old logic for backward compatibility
		if keeperHealth.IsImua {
			taskExecutionAddress = config.GetImuaTaskExecutionAddress()
		} else {
			taskExecutionAddress = config.GetTestTaskExecutionAddress()
		}
	}

	message := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		config.GetEtherscanAPIKey(),
		config.GetAlchemyAPIKey(),
		config.GetPinataHost(),
		config.GetPinataJWT(),
		config.GetDispatcherSigningAddress(),
		taskExecutionAddress,
	)
	msgData, err := cryptography.EncryptMessage(keeperHealth.ConsensusPubKey, message)
	if err != nil {
		h.logger.Error(context.Background(), "Failed to encrypt message for keeper",
			observability.Error(err),
		)
		response.Status = false
		response.Data = err.Error()
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response.Status = true
	response.Data = msgData
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetDetailedKeeperStatus(c *gin.Context) {
	ctx := c.Request.Context()
	total, active := h.stateManager.GetKeeperCount(ctx)
	detailedInfo := h.stateManager.GetDetailedKeeperInfo(ctx)

	// Update keeper metrics
	metrics.UpdateKeeperCounts(ctx, total, active)

	// Get keeper uptimes from database and update metrics
	// This uses the cumulative uptime stored in the database, which is more accurate
	// than calculating from last check-in time
	uptimes, err := h.stateManager.GetKeeperUptimes(ctx)
	if err != nil {
		h.logger.Warn(ctx, "Failed to get keeper uptimes from database",
			observability.Error(err),
		)
	} else {
		// Update uptime metrics for all keepers (both active and inactive)
		for _, keeper := range detailedInfo {
			if uptime, exists := uptimes[strings.ToLower(keeper.KeeperAddress)]; exists {
				metrics.UpdateKeeperUptime(ctx, keeper.KeeperAddress, float64(uptime))
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}
