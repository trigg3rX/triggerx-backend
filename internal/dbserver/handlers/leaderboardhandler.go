package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetKeeperLeaderboard retrieves the leaderboard data for all keepers
func (h *Handler) GetKeeperLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	query := `SELECT keeper_id, keeper_address, keeper_name, no_exctask, keeper_points 
              FROM triggerx.keeper_data 
              WHERE status = true ALLOW FILTERING`

	iter := h.db.Session().Query(query).Iter()

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var keeperEntry types.KeeperLeaderboardEntry

	// Scan all rows into keeper leaderboard entries
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for %d keepers", len(keeperLeaderboard))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperLeaderboard)
}

// GetUserLeaderboard retrieves the leaderboard data for all users
func (h *Handler) GetUserLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	// CQL query to get user leaderboard data
	query := `SELECT user_id, user_address, user_points 
              FROM triggerx.user_data 
              WHERE status = true ALLOW FILTERING`

	iter := h.db.Session().Query(query).Iter()

	var userLeaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	// Scan all rows into user leaderboard entries
	for iter.Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.UserPoints,
	) {
		// Count total jobs for the user
		jobCountQuery := `SELECT COUNT(*) FROM triggerx.job_data WHERE user_address = ? ALLOW FILTERING`
		var totalJobs int
		if err := h.db.Session().Query(jobCountQuery, userEntry.UserAddress).Scan(&totalJobs); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error counting jobs for user %s: %v", userEntry.UserAddress, err)
			totalJobs = 0 // Default to 0 if there's an error
		}
		userEntry.TotalJobs = int64(totalJobs)

		// Count tasks completed for the user
		// Refactor this part to avoid joins
		// You may need to fetch tasks separately or redesign your data model
		var tasksCompleted int
		// Example of a separate query to count tasks
		tasksCountQuery := `SELECT COUNT(*) FROM triggerx.task_data WHERE user_address = ? AND execution_timestamp IS NOT NULL ALLOW FILTERING`
		if err := h.db.Session().Query(tasksCountQuery, userEntry.UserAddress).Scan(&tasksCompleted); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error counting tasks for user %s: %v", userEntry.UserAddress, err)
			tasksCompleted = 0 // Default to 0 if there's an error
		}
		userEntry.TasksCompleted = int64(tasksCompleted)

		userLeaderboard = append(userLeaderboard, userEntry)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetUserLeaderboard] Successfully retrieved user leaderboard data for %d users", len(userLeaderboard))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userLeaderboard)
}
