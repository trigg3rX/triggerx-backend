package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"net/http"
)

func (h *Handler) GetKeeperLeaderboard(c *gin.Context) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	query := `SELECT keeper_id, keeper_address, keeper_name, no_exctask, keeper_points 
              FROM triggerx.keeper_data 
              WHERE status = true AND verified = true ALLOW FILTERING`
	iter := h.db.Session().Query(query).Iter()

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

	query := `SELECT user_id, user_address, user_points 
              FROM triggerx.user_data`

	iter := h.db.Session().Query(query).Iter()

	var userLeaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	for iter.Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.UserPoints,
	) {
		jobCountQuery := `SELECT COUNT(*) FROM triggerx.job_data WHERE user_id = ? ALLOW FILTERING`
		var totalJobs int
		if err := h.db.Session().Query(jobCountQuery, userEntry.UserID).Scan(&totalJobs); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error counting jobs for user %s: %v", userEntry.UserID, err)
			totalJobs = 0
		}
		userEntry.TotalJobs = int64(totalJobs)

		jobIDsQuery := `SELECT job_ids FROM triggerx.user_data WHERE user_id = ?`
		var jobIDs []int64
		if err := h.db.Session().Query(jobIDsQuery, userEntry.UserID).Scan(&jobIDs); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error getting job IDs for user %d: %v", userEntry.UserID, err)
			jobIDs = []int64{}
		}

		var tasksCompleted int64
		for _, jobID := range jobIDs {
			var taskCount int
			taskCountQuery := `SELECT COUNT(*) FROM triggerx.task_data WHERE job_id = ? ALLOW FILTERING`
			if err := h.db.Session().Query(taskCountQuery, jobID).Scan(&taskCount); err != nil {
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
		query = `SELECT keeper_id, keeper_address, keeper_name, no_exctask, keeper_points 
                FROM triggerx.keeper_data 
                WHERE status = true AND keeper_address = ? ALLOW FILTERING`
		args = append(args, keeperAddress)
	} else {
		query = `SELECT keeper_id, keeper_address, keeper_name, no_exctask, keeper_points 
                FROM triggerx.keeper_data 
                WHERE status = true AND keeper_name = ? ALLOW FILTERING`
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

	query := `SELECT user_id, user_address, user_points 
              FROM triggerx.user_data 
              WHERE user_address = ? ALLOW FILTERING`

	var userEntry types.UserLeaderboardEntry
	if err := h.db.Session().Query(query, userAddress).Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.UserPoints,
	); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error fetching user data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	jobCountQuery := `SELECT COUNT(*) FROM triggerx.job_data WHERE user_address = ? ALLOW FILTERING`
	var totalJobs int
	if err := h.db.Session().Query(jobCountQuery, userAddress).Scan(&totalJobs); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error counting jobs for user %s: %v", userAddress, err)
		totalJobs = 0
	}
	userEntry.TotalJobs = int64(totalJobs)

	tasksCountQuery := `SELECT COUNT(*) FROM triggerx.task_data WHERE user_address = ? AND execution_timestamp IS NOT NULL ALLOW FILTERING`
	var tasksCompleted int
	if err := h.db.Session().Query(tasksCountQuery, userAddress).Scan(&tasksCompleted); err != nil {
		h.logger.Errorf("[GetUserByAddress] Error counting tasks for user %s: %v", userAddress, err)
		tasksCompleted = 0
	}
	userEntry.TasksCompleted = int64(tasksCompleted)

	h.logger.Infof("[GetUserByAddress] Successfully retrieved user data for %s", userAddress)
	c.JSON(http.StatusOK, userEntry)
}
