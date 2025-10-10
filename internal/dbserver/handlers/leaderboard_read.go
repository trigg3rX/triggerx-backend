package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	logger := h.getLogger(c)

	// Get the domain from the request
	host := c.Request.Host
	logger.Debugf("GET [GetKeeperLeaderboard] Request from domain: %s", host)

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var err error

	// Get all keepers
	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	allKeepers, err := h.keeperRepository.List(c.Request.Context())
	trackDBOp(err)

	// Determine which data to return based on domain
	var filteredKeepers []*types.KeeperDataEntity
	switch host {
	case "app.triggerx.network":
		logger.Debugf("GET [GetKeeperLeaderboard] Filtering for app.triggerx.network - showing keepers with on_imua = false")
		for _, keeper := range allKeepers {
			if !keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	case "imua.triggerx.network":
		logger.Debugf("GET [GetKeeperLeaderboard] Filtering for imua.triggerx.network - showing keepers with on_imua = true")
		for _, keeper := range allKeepers {
			if keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	default:
		logger.Debugf("GET [GetKeeperLeaderboard] Default domain - showing all keepers")
		filteredKeepers = allKeepers
	}

	// Convert to leaderboard entries and sort by points
	keeperLeaderboard = make([]types.KeeperLeaderboardEntry, 0, len(filteredKeepers))
	for _, keeper := range filteredKeepers {
		keeperLeaderboard = append(keeperLeaderboard, types.KeeperLeaderboardEntry{
			KeeperAddress: keeper.KeeperAddress,
			KeeperName:    keeper.KeeperName,
			KeeperPoints:  keeper.KeeperPoints,
		})
	}

	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		return keeperLeaderboard[i].KeeperPoints > keeperLeaderboard[j].KeeperPoints
	})

	if err != nil {
		logger.Errorf("Error fetching keeper leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keeper leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	logger.Debugf("Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))

	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetUserLeaderboard] Fetching user leaderboard data")

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	allUsers, err := h.userRepository.List(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error fetching user leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	// Convert to leaderboard entries and sort by points
	userLeaderboard := make([]types.UserLeaderboardEntry, 0, len(allUsers))
	for _, user := range allUsers {
		userLeaderboard = append(userLeaderboard, types.UserLeaderboardEntry{
			UserAddress: user.UserAddress,
			UserPoints:  user.OpxConsumed,
			TotalTasks:  user.TotalTasks,
		})
	}

	sort.Slice(userLeaderboard, func(i, j int) bool {
		return userLeaderboard[i].UserPoints > userLeaderboard[j].UserPoints
	})

	logger.Debugf("Successfully retrieved user leaderboard data for %d users", len(userLeaderboard))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetKeeperByIdentifier] Fetching keeper data by identifier")
	h.logger.Info("[GetKeeperByIdentifier] Fetching keeper data by identifier")

	keeperAddress := c.Query("keeper_address")
	keeperName := c.Query("keeper_name")

	if keeperAddress == "" && keeperName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either keeper_address or keeper_name must be provided",
			"code":  "MISSING_IDENTIFIER",
		})
		return
	}

	var keeper *types.KeeperDataEntity
	var err error

	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	if keeperAddress != "" {
		keeper, err = h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_address", strings.ToLower(keeperAddress))
	} else {
		keeper, err = h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_name", keeperName)
	}
	trackDBOp(err)

	if err != nil || keeper == nil {
		logger.Errorf("Error fetching keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	keeperEntry := types.KeeperLeaderboardEntry{
		KeeperAddress: keeper.KeeperAddress,
		KeeperName:    keeper.KeeperName,
		KeeperPoints:  keeper.KeeperPoints,
	}

	logger.Debugf("Successfully retrieved keeper data for %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetUserLeaderboardByAddress] Fetching user data by address")

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_address must be provided",
			"code":  "MISSING_ADDRESS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	user, err := h.userRepository.GetByNonID(c.Request.Context(), "user_address", strings.ToLower(userAddress))
	trackDBOp(err)
	if err != nil || user == nil {
		logger.Errorf("Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	userEntry := types.UserLeaderboardEntry{
		UserAddress: user.UserAddress,
		UserPoints:  user.OpxConsumed,
		TotalTasks:  user.TotalTasks,
	}

	logger.Debugf("Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
