package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleGetRewardsHealth returns the health status of the rewards system
func (h *Handler) HandleGetRewardsHealth(c *gin.Context) {
	if h.rewardsService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "Rewards service not initialized",
			"status": "unavailable",
		})
		return
	}

	healthStatus := h.rewardsService.GetRewardsHealth()

	// Determine HTTP status based on rewards health
	status := http.StatusOK
	if healthStatus["status"] == "degraded" {
		status = http.StatusServiceUnavailable
	} else if healthStatus["status"] == "overdue" {
		status = http.StatusInternalServerError
	}

	c.JSON(status, healthStatus)
}

// HandleGetKeeperDailyUptime returns the daily uptime for a specific keeper
func (h *Handler) HandleGetKeeperDailyUptime(c *gin.Context) {
	keeperAddress := c.Query("keeper_address")
	if keeperAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "keeper_address query parameter is required",
		})
		return
	}

	if h.rewardsService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Rewards service not initialized",
		})
		return
	}

	uptime, err := h.rewardsService.GetKeeperDailyUptime(keeperAddress)
	if err != nil {
		h.logger.Error("Failed to get keeper daily uptime",
			"keeper", keeperAddress,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve daily uptime",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keeper_address":       keeperAddress,
		"daily_uptime_seconds": int64(uptime.Seconds()),
		"daily_uptime_hours":   uptime.Hours(),
	})
}
