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
	h.logger.Info(c.Request.Context(), "[GetKeeperLeaderboard] Request from domain", observability.String("host", host))

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var err error

	// Determine which data to return based on domain
	switch host {
	case "app.triggerx.network":
		h.logger.Info(c.Request.Context(), "[GetKeeperLeaderboard] Filtering for app.triggerx.network - showing keepers with on_imua = false")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_app")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(false)
		trackDBOp(err)
	case "imua.triggerx.network":
		h.logger.Info(c.Request.Context(), "[GetKeeperLeaderboard] Filtering for imua.triggerx.network - showing keepers with on_imua = true")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_imua")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(true)
		trackDBOp(err)
	default:
		h.logger.Info(c.Request.Context(), "[GetKeeperLeaderboard] Default domain - showing all keepers")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_all")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboard()
		trackDBOp(err)
	}

	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperLeaderboard] Error fetching keeper leaderboard data", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keeper leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for keepers", observability.Int("keepers_count", len(keeperLeaderboard)))

	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	h.logger.Info(c.Request.Context(), "[GetUserLeaderboard] Fetching user leaderboard data")

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	userLeaderboard, err := h.userRepository.GetUserLeaderboard()
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetUserLeaderboard] Error fetching user leaderboard data", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetUserLeaderboard] Successfully retrieved user leaderboard data for users", observability.Int("users_count", len(userLeaderboard)))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	h.logger.Info(c.Request.Context(), "[GetKeeperByIdentifier] Fetching keeper data by identifier")

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
		h.logger.Error(c.Request.Context(), "[GetKeeperByIdentifier] Error fetching keeper data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperByIdentifier] Successfully retrieved keeper data for keeper address", observability.String("keeper_address", keeperEntry.KeeperAddress))
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	h.logger.Info(c.Request.Context(), "[GetUserLeaderboardByAddress] Fetching user data by address")

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
		h.logger.Error(c.Request.Context(), "[GetUserLeaderboardByAddress] Error fetching user data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetUserLeaderboardByAddress] Successfully retrieved user data for user address", observability.String("user_address", userAddress))
	c.JSON(http.StatusOK, userEntry)
}
