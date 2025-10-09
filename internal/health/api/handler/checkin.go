package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HandleCheckInEvent processes health check-in requests from keepers
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
	h.handleVersionSpecificResponse(c, &keeperHealth, &response)
}

// handleVersionSpecificResponse returns version-specific responses
func (h *Handler) handleVersionSpecificResponse(c *gin.Context, keeperHealth *types.HealthKeeperCheckInRequest, response *types.HealthKeeperCheckInResponse) {
	switch keeperHealth.Version {
	case "0.1.6", "0.2.0", "0.2.1", "0.2.2", "1.0.0", "1.0.1", "1.0.2":
		// Latest version - return msgData with no warning
		message := h.buildConfigMessage(keeperHealth.IsImua)
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
		message := fmt.Sprintf("%s:%s:%s:%s",
			config.GetEtherscanAPIKey(),
			config.GetAlchemyAPIKey(),
			config.GetPinataHost(),
			config.GetPinataJWT(),
		)
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
		response.Status = true
		response.Data = "UPGRADE TO v0.1.6 for full functionality"
		c.JSON(http.StatusOK, response)
	}
}

// buildConfigMessage constructs the configuration message for keepers
func (h *Handler) buildConfigMessage(isImua bool) string {
	if isImua {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			config.GetEtherscanAPIKey(),
			config.GetAlchemyAPIKey(),
			config.GetPinataHost(),
			config.GetPinataJWT(),
			config.GetManagerSigningAddress(),
			config.GetImuaTaskExecutionAddress(),
		)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		config.GetEtherscanAPIKey(),
		config.GetAlchemyAPIKey(),
		config.GetPinataHost(),
		config.GetPinataJWT(),
		config.GetManagerSigningAddress(),
		config.GetTaskExecutionAddress(),
	)
}
