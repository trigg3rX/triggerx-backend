package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetUserData(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Infof("[GetUserData] Retrieving user with ID: %s", userID)

	var userData types.UserData

	if err := h.db.Session().Query(queries.SelectUserDataByIDQuery, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs, &userData.AccountBalance); err != nil {
		h.logger.Errorf("[GetUserData] Error retrieving user with ID %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetUserData] Successfully retrieved user with ID: %s", userID)

	response := types.UserData{
		UserID:         userData.UserID,
		UserAddress:    userData.UserAddress,
		JobIDs:         userData.JobIDs,
		AccountBalance: userData.AccountBalance,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetWalletPoints(c *gin.Context) {
	walletAddress := strings.ToLower(c.Param("wallet_address"))
	h.logger.Infof("[GetWalletPoints] Retrieving points for wallet address: %s", walletAddress)

	var userPoints int
	var keeperPoints int

	if err := h.db.Session().Query(queries.SelectUserPointsByAddressQuery, walletAddress).Scan(&userPoints); err != nil {
		userPoints = 0
	}

	if err := h.db.Session().Query(queries.SelectKeeperPointsByAddressQuery, walletAddress).Scan(&keeperPoints); err != nil {
		keeperPoints = 0
	}

	h.logger.Infof("[GetWalletPoints] Successfully retrieved points for wallet address %s: %d + %d", walletAddress, userPoints, keeperPoints)

	totalPoints := userPoints + keeperPoints

	c.JSON(http.StatusOK, gin.H{
		"total_points": totalPoints,
	})
}
