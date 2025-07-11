package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	// Get the domain from the request
	host := c.Request.Host
	h.logger.Infof("[GetKeeperLeaderboard] Request from domain: %s", host)

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var err error

	// Determine which data to return based on domain
	switch host {
	case "app.triggerx.network":
		h.logger.Info("[GetKeeperLeaderboard] Filtering for app.triggerx.network - showing keepers with on_imua = false")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_app")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(false)
		trackDBOp(err)
	case "imua.triggerx.network":
		h.logger.Info("[GetKeeperLeaderboard] Filtering for imua.triggerx.network - showing keepers with on_imua = true")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_imua")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboardByOnImua(true)
		trackDBOp(err)
	default:
		h.logger.Info("[GetKeeperLeaderboard] Default domain - showing all keepers")
		trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard_all")
		keeperLeaderboard, err = h.keeperRepository.GetKeeperLeaderboard()
		trackDBOp(err)
	}

	if err != nil {
		h.logger.Errorf("[GetKeeperLeaderboard] Error fetching keeper leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keeper leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))

	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	userLeaderboard, err := h.userRepository.GetUserLeaderboard()
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	h.logger.Infof("[GetUserLeaderboard] Successfully retrieved user leaderboard data for %d users", len(userLeaderboard))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	h.logger.Info("[GetKeeperByIdentifier] Fetching keeper data by identifier")

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
		h.logger.Errorf("[GetKeeperByIdentifier] Error fetching keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetKeeperByIdentifier] Successfully retrieved keeper data for %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	h.logger.Info("[GetUserLeaderboardByAddress] Fetching user data by address")

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
		h.logger.Errorf("[GetUserLeaderboardByAddress] Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetUserLeaderboardByAddress] Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
