package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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
	userData, err := h.userRepository.GetUserData(userAddress)
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
	h.logger.Debug(c.Request.Context(), "[GetUserDataByAddress] Retrieved user data", observability.String("user_address", userAddress))
}

func (h *Handler) GetWalletPoints(c *gin.Context) {
	walletAddress := strings.ToLower(c.Param("address"))

	var userPoints string
	var keeperPoints string

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userPoints, err := h.userRepository.GetUserPoints(walletAddress)
	trackDBOp(err)
	if err != nil {
		userPoints = "0"
	}

	keeperPoints, err = h.keeperRepository.GetKeeperPoints(walletAddress)
	if err != nil {
		keeperPoints = "0"
	}

	totalPoints := types.Add(userPoints, keeperPoints)

	c.JSON(http.StatusOK, gin.H{
		"total_points": totalPoints,
	})
	h.logger.Debug(c.Request.Context(), "[GetWalletPoints] Retrieved wallet points", observability.String("wallet_address", walletAddress), observability.String("total_points", totalPoints))
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
