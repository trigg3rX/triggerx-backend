package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

func (h *Handler) GetUserDataByAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetUserDataByAddress] trace_id=%s - Retrieving user data by address", traceID)
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetWalletPoints] trace_id=%s - Retrieving wallet points", traceID)
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

func (h *Handler) StoreUserEmail(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[StoreUserEmail] trace_id=%s - Storing user email", traceID)

	var req struct {
		UserAddress string `json:"user_address"`
		Email       string `json:"email_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[StoreUserEmail] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": "INVALID_REQUEST"})
		return
	}
	if req.UserAddress == "" || req.Email == "" {
		h.logger.Errorf("[StoreUserEmail] Missing user_address or email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_address or email", "code": "MISSING_FIELDS"})
		return
	}

	req.UserAddress = strings.ToLower(req.UserAddress)

	err := h.userRepository.UpdateUserEmail(req.UserAddress, req.Email)
	if err != nil {
		h.logger.Errorf("[StoreUserEmail] Failed to update email %s for address %s: %v", req.Email, req.UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email", "code": "UPDATE_FAILED"})
		return
	}

	h.logger.Infof("[StoreUserEmail] Successfully updated email for address: %s", req.UserAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}
