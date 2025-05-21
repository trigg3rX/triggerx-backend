package health

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       logging.Logger
	stateManager *keeper.StateManager
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, stateManager *keeper.StateManager) *Handler {
	return &Handler{
		logger:       logger.With("component", "health_handler"),
		stateManager: stateManager,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	middlewareLogger := logger.With("component", "http_middleware")
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

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
func RegisterRoutes(router *gin.Engine) {
	handler := NewHandler(logging.GetServiceLogger(), keeper.GetStateManager())

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
	var keeperHealth types.KeeperHealthCheckIn
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error("Failed to parse keeper health check-in request",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Received keeper health check-in",
		"keeper", keeperHealth.KeeperAddress,
		"version", keeperHealth.Version,
		"peer_id", keeperHealth.PeerID,
	)

	if keeperHealth.Version == "0.0.7" || keeperHealth.Version == "0.0.6" || keeperHealth.Version == "0.0.5" || keeperHealth.Version == "" {
		h.logger.Warn("Rejecting obsolete keeper version",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.2",
		})
		return
	}

	if keeperHealth.Version == "0.1.0" || keeperHealth.Version == "0.1.1" {
		h.logger.Warn("Older keeper version detected",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusOK, gin.H{
			"message": "OLDER VERSION of Keeper, UPGRADE TO v0.1.2",
		})
		return
	}

	if keeperHealth.Version == "0.1.2" {
		ok, err := VerifySignature(keeperHealth.KeeperAddress, keeperHealth.Signature, keeperHealth.ConsensusAddress)
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

		c.JSON(http.StatusOK, gin.H{
			"message": "Keeper health check-in received",
			"active":  true,
		})
	}
}

func VerifySignature(message string, signatureHex string, expectedAddress string) (bool, error) {
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length")
	}

	if signature[64] >= 27 {
		signature[64] -= 27
	}

	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	pubKeyRaw, err := crypto.Ecrecover(messageHash.Bytes(), signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	checksumAddr := common.HexToAddress(expectedAddress)

	return checksumAddr == recoveredAddr, nil
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
