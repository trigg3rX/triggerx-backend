package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

func (h *Handler) GetUserDataByAddress(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Param("address"))
	if userAddress == "" {
		logger.Debugf("Invalid user address: %v", userAddress)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}
	logger.Debugf("GET [GetUserDataByAddress] For address: %s", c.Param("address"))

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userData, err := h.userRepository.GetByNonID(c.Request.Context(), "user_address", userAddress)
	trackDBOp(err)
	if err != nil || userData == nil {
		logger.Errorf("Error retrieving user with address %s: %v", userAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	logger.Debugf("Successfully retrieved user: %s", userData.UserAddress)
	c.JSON(http.StatusOK, userData)
}

func (h *Handler) StoreUserEmail(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("POST [StoreUserEmail] For address: %s", c.Param("address"))

	var req struct {
		UserAddress string `json:"user_address"`
		Email       string `json:"email_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": "INVALID_REQUEST"})
		return
	}
	if req.UserAddress == "" || req.Email == "" {
		logger.Errorf("Missing user_address or email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_address or email", "code": "MISSING_FIELDS"})
		return
	}

	req.UserAddress = strings.ToLower(req.UserAddress)

	// Get user
	user, err := h.userRepository.GetByNonID(c.Request.Context(), "user_address", req.UserAddress)
	if err != nil || user == nil {
		logger.Errorf("User not found for address %s: %v", req.UserAddress, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "code": "USER_NOT_FOUND"})
		return
	}

	// Update email
	user.EmailID = req.Email

	err = h.userRepository.Update(c.Request.Context(), user)
	if err != nil {
		logger.Errorf("Failed to update email %s for address %s: %v", req.Email, req.UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email", "code": "UPDATE_FAILED"})
		return
	}

	logger.Debugf("Successfully updated email for address: %s", req.UserAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}
