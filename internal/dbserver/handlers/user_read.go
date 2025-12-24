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

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userID, userData, err := h.userRepository.GetUserDataByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetUserDataByAddress] Failed to retrieve user", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, userData)
	h.logger.Debug(c.Request.Context(), "[GetUserDataByAddress] Retrieved user data", observability.Int64("user_id", userID), observability.String("user_address", userAddress))
}

func (h *Handler) GetWalletPoints(c *gin.Context) {
	walletAddress := strings.ToLower(c.Param("address"))

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

	totalPoints := userPoints + keeperPoints

	c.JSON(http.StatusOK, gin.H{
		"total_points": totalPoints,
	})
	h.logger.Debug(c.Request.Context(), "[GetWalletPoints] Retrieved wallet points", observability.String("wallet_address", walletAddress), observability.Float64("total_points", totalPoints))
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

	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
	h.logger.Info(c.Request.Context(), "[StoreUserEmail] Updated user email", observability.String("user_address", req.UserAddress))
}
