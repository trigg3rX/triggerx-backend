package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) GetUserDataByAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("address"))
	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetUserDataByAddress] Invalid user address", observability.String("user_address", userAddress))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetUserDataByAddress] Retrieving user with address", observability.String("user_address", userAddress))

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userID, userData, err := h.userRepository.GetUserDataByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetUserData] Error retrieving user with ID", observability.Int64("user_id", userID), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetUserData] Successfully retrieved user with ID", observability.Int64("user_id", userID))
	c.JSON(http.StatusOK, userData)
}

func (h *Handler) GetWalletPoints(c *gin.Context) {
	walletAddress := strings.ToLower(c.Param("address"))
	h.logger.Info(c.Request.Context(), "[GetWalletPoints] Retrieving points for wallet address", observability.String("wallet_address", walletAddress))

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

	h.logger.Info(c.Request.Context(), "[GetWalletPoints] Successfully retrieved points for wallet address", observability.String("wallet_address", walletAddress), observability.Float64("user_points", userPoints), observability.Float64("keeper_points", keeperPoints))

	totalPoints := userPoints + keeperPoints

	c.JSON(http.StatusOK, gin.H{
		"total_points": totalPoints,
	})
}

func (h *Handler) StoreUserEmail(c *gin.Context) {
	var req struct {
		UserAddress string `json:"user_address"`
		Email       string `json:"email_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "[StoreUserEmail] Invalid request", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": "INVALID_REQUEST"})
		return
	}
	if req.UserAddress == "" || req.Email == "" {
		h.logger.Error(c.Request.Context(), "[StoreUserEmail] Missing user_address or email", observability.String("user_address", req.UserAddress), observability.String("email", req.Email))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_address or email", "code": "MISSING_FIELDS"})
		return
	}

	req.UserAddress = strings.ToLower(req.UserAddress)

	err := h.userRepository.UpdateUserEmail(req.UserAddress, req.Email)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[StoreUserEmail] Failed to update email for address", observability.String("email", req.Email), observability.String("user_address", req.UserAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email", "code": "UPDATE_FAILED"})
		return
	}

	h.logger.Info(c.Request.Context(), "[StoreUserEmail] Successfully updated email for address", observability.String("user_address", req.UserAddress))
	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}
