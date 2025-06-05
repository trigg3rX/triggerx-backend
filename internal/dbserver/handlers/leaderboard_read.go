package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	keeperLeaderboard, err := h.keeperRepository.GetKeeperLeaderboard()
	if err != nil {
		h.logger.Errorf("[GetKeeperLeaderboard] Error fetching keeper leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))

	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	userLeaderboard, err := h.userRepository.GetUserLeaderboard()
	if err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either keeper_address or keeper_name must be provided"})
		return
	}

	keeperEntry, err := h.keeperRepository.GetKeeperLeaderboardByIdentifierInDB(keeperAddress, keeperName)
	if err != nil {
		h.logger.Errorf("[GetKeeperByIdentifier] Error fetching keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Keeper not found"})
		return
	}

	h.logger.Infof("[GetKeeperByIdentifier] Successfully retrieved keeper data for %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	h.logger.Info("[GetUserLeaderboardByAddress] Fetching user data by address")

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address must be provided"})
		return
	}

	userEntry, err := h.userRepository.GetUserLeaderboardByAddress(userAddress)
	if err != nil {
		h.logger.Errorf("[GetUserLeaderboardByAddress] Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.logger.Infof("[GetUserLeaderboardByAddress] Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
