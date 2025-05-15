package health

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

	h.logger.Infof("Keeper health check-in received: %+v", keeperHealth)

	if keeperHealth.Version == "0.0.7" || keeperHealth.Version == "0.0.6" || keeperHealth.Version == "0.0.5" || keeperHealth.Version == "" {
		h.logger.Debug("Obsolete Version of Keeper",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.2",
		})
		return
	}
	if keeperHealth.Version == "0.1.0" || keeperHealth.Version == "0.1.1" {
		h.logger.Debug("Older Version of Keeper",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusOK, gin.H{
			"message": "OLDER VERSION of Keeper, UPGRADE TO v0.1.2",
		})
		return
	}
	if keeperHealth.Version == "0.1.2" {
		ok, _ := VerifySignature(keeperHealth.KeeperAddress, keeperHealth.Signature, keeperHealth.ConsensusAddress)
		if !ok {
			h.logger.Error("Invalid signature",
				"keeper", keeperHealth.KeeperAddress,
				"signature", keeperHealth.Signature,
			)
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error": "Invalid signature",
			})
			return
		} else {
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
