package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	logger := h.getLogger(c)
	host := c.Request.Host
	logger.Debugf("GET [GetKeeperLeaderboard] From domain: %s", host)

	// Get all keepers
	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	allKeepers, err := h.keeperRepository.List(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	// Determine which data to return based on domain
	var filteredKeepers []*types.KeeperDataEntity
	switch host {
	case "app.triggerx.network":
		for _, keeper := range allKeepers {
			if !keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	case "imua.triggerx.network":
		for _, keeper := range allKeepers {
			if keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	default:
		filteredKeepers = allKeepers
	}

	// Convert to leaderboard entries and sort by points
	keeperLeaderboard := make([]types.KeeperLeaderboardEntry, 0, len(filteredKeepers))
	for _, keeper := range filteredKeepers {
		keeperLeaderboard = append(keeperLeaderboard, types.KeeperLeaderboardEntry{
			KeeperAddress:   keeper.KeeperAddress,
			OperatorID:      keeper.OperatorID,
			KeeperName:      keeper.KeeperName,
			KeeperPoints:    keeper.KeeperPoints,
			NoExecutedTasks: keeper.NoExecutedTasks,
			NoAttestedTasks: keeper.NoAttestedTasks,
			OnImua:          keeper.OnImua,
		})
	}

	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		return keeperLeaderboard[i].KeeperPoints > keeperLeaderboard[j].KeeperPoints
	})

	logger.Debugf("GET [GetKeeperLeaderboard] Successful, keepers: %d", len(keeperLeaderboard))
	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetUserLeaderboard] Fetching")

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	allUsers, err := h.userRepository.List(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	// Convert to leaderboard entries and sort by points
	userLeaderboard := make([]types.UserLeaderboardEntry, 0, len(allUsers))
	for _, user := range allUsers {
		userLeaderboard = append(userLeaderboard, types.UserLeaderboardEntry{
			UserAddress: user.UserAddress,
			UserPoints:  user.UserPoints,
			TotalJobs:   user.TotalJobs,
			TotalTasks:  user.TotalTasks,
		})
	}

	sort.Slice(userLeaderboard, func(i, j int) bool {
		return userLeaderboard[i].UserPoints > userLeaderboard[j].UserPoints
	})

	logger.Debugf("GET [GetUserLeaderboard] Successful, users: %d", len(userLeaderboard))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	logger := h.getLogger(c)
	keeperAddress := strings.ToLower(c.Query("keeper_address"))
	keeperName := c.Query("keeper_name")
	if keeperAddress == "" && keeperName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequestBody,
		})
		return
	}
	if keeperAddress != "" && keeperName == "" {
		logger.Debugf("GET [GetKeeperByIdentifier] By address: %s", keeperAddress)
	} else if keeperAddress == "" && keeperName != "" {
		logger.Debugf("GET [GetKeeperByIdentifier] By name: %s", keeperName)
	}

	var keeper *types.KeeperDataEntity
	var err error

	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	if keeperAddress != "" {
		keeper, err = h.keeperRepository.GetByID(c.Request.Context(), strings.ToLower(keeperAddress))
	} else {
		keeper, err = h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_name", keeperName)
	}
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	keeperEntry := types.KeeperLeaderboardEntry{
		KeeperAddress:   keeper.KeeperAddress,
		OperatorID:      keeper.OperatorID,
		KeeperName:      keeper.KeeperName,
		KeeperPoints:    keeper.KeeperPoints,
		NoExecutedTasks: keeper.NoExecutedTasks,
		NoAttestedTasks: keeper.NoAttestedTasks,
		OnImua:          keeper.OnImua,
	}

	logger.Debugf("GET [GetKeeperByIdentifier] Successful, keeper: %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Query("user_address"))
	if !types.IsValidEthAddress(userAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetUserLeaderboardByAddress] For address: %s", userAddress)

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	user, err := h.userRepository.GetByID(c.Request.Context(), strings.ToLower(userAddress))
	trackDBOp(err)
	if err != nil || user == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	userEntry := types.UserLeaderboardEntry{
		UserAddress: user.UserAddress,
		UserPoints:  user.UserPoints,
		TotalJobs:   user.TotalJobs,
		TotalTasks:  user.TotalTasks,
	}

	logger.Debugf("GET [GetUserLeaderboardByAddress] Successful, user: %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
