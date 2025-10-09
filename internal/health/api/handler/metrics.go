package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleRoot returns basic service information
func (h *Handler) HandleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Health Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// GetKeeperStatus returns summary statistics about keepers
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

// GetDetailedKeeperStatus returns detailed information about all keepers
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
