package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	iter := h.db.Session().Query(queries.SelectKeeperLeaderboardQuery).Iter()

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var keeperEntry types.KeeperLeaderboardEntry

	for iter.Scan(
		&keeperEntry.KeeperID,
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.TasksExecuted,
		&keeperEntry.KeeperPoints,
	) {
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetKeeperLeaderboard] Error fetching keeper leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))
	c.JSON(http.StatusOK, keeperLeaderboard)
}

func (h *Handler) GetUserLeaderboard(c *gin.Context) {
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	iter := h.db.Session().Query(queries.SelectUserLeaderboardQuery).Iter()

	var userLeaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	for iter.Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.UserPoints,
	) {
		var totalJobs int
		if err := h.db.Session().Query(queries.SelectUserJobCountQuery, userEntry.UserID).Scan(&totalJobs); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error counting jobs for user %s: %v", userEntry.UserID, err)
			totalJobs = 0
		}
		userEntry.TotalJobs = int64(totalJobs)

		var jobIDs []int64
		if err := h.db.Session().Query(queries.SelectUserJobIDsByIDQuery, userEntry.UserID).Scan(&jobIDs); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error getting job IDs for user %d: %v", userEntry.UserID, err)
			jobIDs = []int64{}
		}

		var tasksCompleted int64
		for _, jobID := range jobIDs {
			var taskCount int
			if err := h.db.Session().Query(queries.SelectTaskCountByJobIDQuery, jobID).Scan(&taskCount); err != nil {
				h.logger.Errorf("[GetUserLeaderboard] Error counting tasks for job %d: %v", jobID, err)
				continue
			}
			tasksCompleted += int64(taskCount)
		}
		userEntry.TasksCompleted = tasksCompleted

		userLeaderboard = append(userLeaderboard, userEntry)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetUserLeaderboard] Successfully retrieved user leaderboard data for %d users", len(userLeaderboard))
	c.JSON(http.StatusOK, userLeaderboard)
}

func (h *Handler) GetKeeperByIdentifier(c *gin.Context) {
	h.logger.Info("[GetKeeperByIdentifier] Fetching keeper data by identifier")

	keeperAddress := c.Query("keeper_address")
	keeperName := c.Query("keeper_name")

	if keeperAddress == "" && keeperName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either keeper_address or keeper_name must be provided"})
		return
	}

	var query string
	var args []interface{}

	if keeperAddress != "" {
		query = queries.SelectKeeperByAddressQuery
		args = append(args, keeperAddress)
	} else {
		query = queries.SelectKeeperByNameQuery
		args = append(args, keeperName)
	}

	var keeperEntry types.KeeperLeaderboardEntry
	if err := h.db.Session().Query(query, args...).Scan(
		&keeperEntry.KeeperID,
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.TasksExecuted,
		&keeperEntry.KeeperPoints,
	); err != nil {
		h.logger.Errorf("[GetKeeperByIdentifier] Error fetching keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Keeper not found"})
		return
	}

	h.logger.Infof("[GetKeeperByIdentifier] Successfully retrieved keeper data for %s", keeperEntry.KeeperAddress)
	c.JSON(http.StatusOK, keeperEntry)
}

func (h *Handler) GetUserByAddress(c *gin.Context) {
	h.logger.Info("[GetUserByAddress] Fetching user data by address")

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address must be provided"})
		return
	}

	var userEntry types.UserLeaderboardEntry
	if err := h.db.Session().Query(queries.SelectUserByAddressQuery, userAddress).Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.UserPoints,
	); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var totalJobs int
	if err := h.db.Session().Query(queries.SelectUserJobCountByAddressQuery, userAddress).Scan(&totalJobs); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error counting jobs for user %s: %v", userAddress, err)
		totalJobs = 0
	}
	userEntry.TotalJobs = int64(totalJobs)

	var tasksCompleted int
	if err := h.db.Session().Query(queries.SelectUserTaskCountByAddressQuery, userAddress).Scan(&tasksCompleted); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error counting tasks for user %s: %v", userAddress, err)
		tasksCompleted = 0
	}
	userEntry.TasksCompleted = int64(tasksCompleted)

	h.logger.Infof("[GetUserByAddress] Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
