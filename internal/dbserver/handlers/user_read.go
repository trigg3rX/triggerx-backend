package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
)

func (h *Handler) GetUserDataByAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("address"))
	if userAddress == "" {
		h.logger.Errorf("[GetUserDataByAddress] Invalid user address: %v", userAddress)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}

	h.logger.Infof("[GetUserDataByAddress] Retrieving user with address: %s", userAddress)

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userID, userData, err := h.userRepository.GetUserDataByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetUserData] Error retrieving user with ID %d: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetUserData] Successfully retrieved user with ID: %d", userID)
	c.JSON(http.StatusOK, userData)
}

func (h *Handler) GetWalletPoints(c *gin.Context) {
	walletAddress := strings.ToLower(c.Param("address"))
	h.logger.Infof("[GetWalletPoints] Retrieving points for wallet address: %s", walletAddress)

	var userPoints float64
	var keeperPoints float64

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userPoints, err := h.userRepository.GetUserPointsByAddress(walletAddress)
	trackDBOp(err)
	if err != nil {
		userPoints = 0
	}

	// keeperPoints, err := h.userRepository.GetKeeperPointsByAddress(walletAddress)
	// if err != nil {
	// 	keeperPoints = 0
	// }

	h.logger.Infof("[GetWalletPoints] Successfully retrieved points for wallet address %s: %.2f + %.2f", walletAddress, userPoints, keeperPoints)

	totalPoints := userPoints + keeperPoints

	c.JSON(http.StatusOK, gin.H{
		"total_points": totalPoints,
	})
}
