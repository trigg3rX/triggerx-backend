package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	// Get the domain from the request
	host := c.Request.Host

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var err error

	// Determine which data to return based on domain
	var trackDBOp func(error)
	switch host {
	case "app.triggerx.network":
		trackDBOp = metrics.TrackDBOperation("read", "keeper_leaderboard_app")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(false)
		trackDBOp(err)
	case "imua.triggerx.network":
		trackDBOp = metrics.TrackDBOperation("read", "keeper_leaderboard_imua")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(true)
		trackDBOp(err)
	default:
		trackDBOp = metrics.TrackDBOperation("read", "keeper_leaderboard_all")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboard()
		trackDBOp(err)
	}

	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperLeaderboard] Failed to fetch keeper leaderboard", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keeper leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, keeperLeaderboard)
	h.logger.Debug(c.Request.Context(), "[GetKeeperLeaderboard] Retrieved keeper leaderboard", observability.Int("keepers_count", len(keeperLeaderboard)))
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	userLeaderboard, err := h.userRepository.GetUserLeaderboard()
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetUserLeaderboard] Failed to fetch user leaderboard", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, userLeaderboard)
	h.logger.Debug(c.Request.Context(), "[GetUserLeaderboard] Retrieved user leaderboard", observability.Int("users_count", len(userLeaderboard)))
}
