package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
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

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {

	keeperAddress := c.Query("keeper_address")
	keeperName := c.Query("keeper_name")

	if keeperAddress == "" && keeperName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either keeper_address or keeper_name must be provided",
			"code":  "MISSING_IDENTIFIER",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	keeperEntry, err := h.keeperRepository.GetKeeperLeaderboardByIdentifierInDB(keeperAddress, keeperName)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperByIdentifier] Failed to fetch keeper data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, keeperEntry)
	h.logger.Debug(c.Request.Context(), "[GetKeeperByIdentifier] Retrieved keeper data", observability.String("keeper_address", keeperEntry.KeeperAddress))
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_address must be provided",
			"code":  "MISSING_ADDRESS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	userEntry, err := h.userRepository.GetUserLeaderboardByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetUserLeaderboardByAddress] Failed to fetch user data", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, userEntry)
	h.logger.Debug(c.Request.Context(), "[GetUserLeaderboardByAddress] Retrieved user data", observability.String("user_address", userAddress))
}
