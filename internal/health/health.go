package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	// "github.com/trigg3rX/triggerx-backend/pkg/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.HealthProcess)

func HandleCheckInEvent(c *gin.Context) {
	var keeperHealth types.KeeperHealth
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature
	// message := keeperHealth.KeeperAddress
	// isValid, err := crypto.VerifySignature(message, keeperHealth.Signature, keeperHealth.KeeperAddress)
	// if err != nil {
	// 	logger.Error("Failed to verify signature", "error", err, "keeper", keeperHealth.KeeperAddress)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
	// 	return
	// }

	// if !isValid {
	// 	logger.Error("Invalid signature for keeper", "keeper", keeperHealth.KeeperAddress)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature, authorization failed"})
	// 	return
	// }

	if keeperHealth.Version == "0.0.7" || keeperHealth.Version == "0.0.6" || keeperHealth.Version == "0.0.5" || keeperHealth.Version == "" {
		} else {
		logger.Debugf("OBSOLETE VERSION !!!  for %s", keeperHealth.KeeperAddress)
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.0"})
		return
	}

	if keeperHealth.Version == "0.1.0" {
		logger.Infof("Check In: %s | %s", keeperHealth.KeeperAddress, c.ClientIP())
	}

	// Update the keeper state in our in-memory store
	stateManager := GetKeeperStateManager()
	if err := stateManager.UpdateKeeperHealth(keeperHealth); err != nil {
		logger.Error("Failed to update keeper state", "error", err, "keeper", keeperHealth.KeeperAddress)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
		return
	}

	// Only send to database when keeper status changes (handled by state manager)
	c.JSON(http.StatusOK, gin.H{
		"message": "Keeper health check-in received",
		"active":  true,
	})
}

// GetKeeperStatus returns the status of keepers
func GetKeeperStatus(c *gin.Context) {
	// Get the state manager and query for counts
	stateManager := GetKeeperStateManager()
	total, active := stateManager.GetKeeperCount()

	// Get list of active keepers
	activeKeepers := stateManager.GetAllActiveKeepers()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":      total,
		"active_keepers":     active,
		"active_keeper_list": activeKeepers,
	})
}
