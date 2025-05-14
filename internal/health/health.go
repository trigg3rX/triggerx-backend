package health

import (
	"net/http"
	"strings"
	"time"

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
		logger.Debugf("OBSOLETE VERSION !!!  for %s", keeperHealth.KeeperAddress)
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.0"})
	}

	if keeperHealth.Version == "0.1.0" {
		logger.Infof("Check In: %s | %s", keeperHealth.KeeperAddress, c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"message": "Keeper health check-in received",
			"active":  true,
		})
	}

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)

	stateManager := GetKeeperStateManager()
	if err := stateManager.UpdateKeeperHealth(keeperHealth); err != nil {
		logger.Error("Failed to update keeper state", "error", err, "keeper", keeperHealth.KeeperAddress)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
		return
	}
}

func GetKeeperStatus(c *gin.Context) {
	stateManager := GetKeeperStateManager()
	total, active := stateManager.GetKeeperCount()

	activeKeepers := stateManager.GetAllActiveKeepers()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":      total,
		"active_keepers":     active,
		"active_keeper_list": activeKeepers,
	})
}

func GetDetailedKeeperStatus(c *gin.Context) {
	stateManager := GetKeeperStateManager()

	total, active := stateManager.GetKeeperCount()

	detailedInfo := stateManager.GetDetailedKeeperInfo()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}
