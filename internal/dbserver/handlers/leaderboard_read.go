package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperLeaderboard] trace_id=%s - Fetching keeper leaderboard data", traceID)

	// Get the domain from the request
	host := c.Request.Host
	h.logger.Infof("[GetKeeperLeaderboard] Request from domain: %s", host)

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var err error

	ctx := context.Background()

	// Get all keepers
	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	allKeepers, err := h.keeperRepository.List(ctx)
	trackDBOp(err)

	// Determine which data to return based on domain
	var filteredKeepers []*types.KeeperDataEntity
	switch host {
	case "app.triggerx.network":
		h.logger.Info("[GetKeeperLeaderboard] Filtering for app.triggerx.network - showing keepers with on_imua = false")
		for _, keeper := range allKeepers {
			if keeper.OnImua != nil && !*keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	case "imua.triggerx.network":
		h.logger.Info("[GetKeeperLeaderboard] Filtering for imua.triggerx.network - showing keepers with on_imua = true")
		for _, keeper := range allKeepers {
			if keeper.OnImua != nil && *keeper.OnImua {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	default:
		h.logger.Info("[GetKeeperLeaderboard] Default domain - showing all keepers")
		filteredKeepers = allKeepers
	}

	// Convert to leaderboard entries and sort by points
	keeperLeaderboard = make([]types.KeeperLeaderboardEntry, 0, len(filteredKeepers))
	for _, keeper := range filteredKeepers {
		points, _ := keeper.KeeperPoints.Float64()
		keeperLeaderboard = append(keeperLeaderboard, types.KeeperLeaderboardEntry{
			KeeperAddress: keeper.KeeperAddress,
			KeeperName:    keeper.KeeperName,
			KeeperPoints:  points,
		})
	}

	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		return keeperLeaderboard[i].KeeperPoints > keeperLeaderboard[j].KeeperPoints
	})

	if err != nil {
		h.logger.Errorf("[GetKeeperLeaderboard] Error fetching keeper leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keeper leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))

	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetUserLeaderboard] trace_id=%s - Fetching user leaderboard data", traceID)
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	allUsers, err := h.userRepository.List(ctx)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user leaderboard",
			"code":  "LEADERBOARD_FETCH_ERROR",
		})
		return
	}

	// Convert to leaderboard entries and sort by points
	userLeaderboard := make([]types.UserLeaderboardEntry, 0, len(allUsers))
	for _, user := range allUsers {
		points, _ := user.OpxConsumed.Float64()
		userLeaderboard = append(userLeaderboard, types.UserLeaderboardEntry{
			UserAddress: user.UserAddress,
			UserPoints:  points,
			TotalTasks:  user.TotalTasks,
		})
	}

	sort.Slice(userLeaderboard, func(i, j int) bool {
		return userLeaderboard[i].UserPoints > userLeaderboard[j].UserPoints
	})

	h.logger.Infof("[GetUserLeaderboard] Successfully retrieved user leaderboard data for %d users", len(userLeaderboard))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperByIdentifier] trace_id=%s - Fetching keeper data by identifier", traceID)
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

	ctx := context.Background()

	var keeper *types.KeeperDataEntity
	var err error

	trackDBOp := metrics.TrackDBOperation("read", "keeper_leaderboard")
	if keeperAddress != "" {
		keeper, err = h.keeperRepository.GetByNonID(ctx, "keeper_address", strings.ToLower(keeperAddress))
	} else {
		keeper, err = h.keeperRepository.GetByNonID(ctx, "keeper_name", keeperName)
	}
	trackDBOp(err)

	if err != nil || keeper == nil {
		h.logger.Errorf("[GetKeeperByIdentifier] Error fetching keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	points, _ := keeper.KeeperPoints.Float64()
	keeperEntry := types.KeeperLeaderboardEntry{
		KeeperAddress: keeper.KeeperAddress,
		KeeperName:    keeper.KeeperName,
		KeeperPoints:  points,
	}

	h.logger.Infof("[GetKeeperByIdentifier] Successfully retrieved keeper data for %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserLeaderboardByAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetUserLeaderboardByAddress] trace_id=%s - Fetching user data by address", traceID)
	h.logger.Info("[GetUserLeaderboardByAddress] Fetching user data by address")

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_address must be provided",
			"code":  "MISSING_ADDRESS",
		})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "user_leaderboard")
	user, err := h.userRepository.GetByNonID(ctx, "user_address", strings.ToLower(userAddress))
	trackDBOp(err)
	if err != nil || user == nil {
		h.logger.Errorf("[GetUserLeaderboardByAddress] Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	points, _ := user.OpxConsumed.Float64()
	userEntry := types.UserLeaderboardEntry{
		UserAddress: user.UserAddress,
		UserPoints:  points,
		TotalTasks:  user.TotalTasks,
	}

	h.logger.Infof("[GetUserLeaderboardByAddress] Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
