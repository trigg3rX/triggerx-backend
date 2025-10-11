package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetUserDataByAddress(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Param("address"))
	if !types.IsValidEthAddress(userAddress) {
		logger.Debugf("%s: %v", errors.ErrInvalidRequestBody, userAddress)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequestBody,
		})
		return
	}
	logger.Debugf("GET [GetUserDataByAddress] For address: %s", userAddress)

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userData, err := h.userRepository.GetByNonID(c.Request.Context(), "user_address", userAddress)
	trackDBOp(err)
	if err != nil || userData == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.ErrDBRecordNotFound,
		})
		return
	}

	logger.Infof("GET [GetUserDataByAddress] Successful, user: %s", userData.UserAddress)
	c.JSON(http.StatusOK, userData)
}

func (h *Handler) UpdateUserEmail(c *gin.Context) {
	logger := h.getLogger(c)
	var req types.UpdateUserEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	req.UserAddress = strings.ToLower(req.UserAddress)
	logger.Debugf("PUT [UpdateUserEmail] For address: %s", req.UserAddress)

	// Get user
	user, err := h.userRepository.GetByID(c.Request.Context(), req.UserAddress)
	if err != nil || user == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Update email
	unsubscribed := false
	if req.Email != "" {
		user.EmailID = req.Email
	} else {
		user.EmailID = ""
		unsubscribed = true
	}

	trackDBOp := metrics.TrackDBOperation("update", "user_data")
	err = h.userRepository.Update(c.Request.Context(), user)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	logger.Infof("PUT [UpdateUserEmail] Successful, unsubscribed: %t, user: %s", unsubscribed, req.UserAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}
