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
             WHERE status = true 
             ORDER BY keeper_points DESC`

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

	// Updated query to include tasks completed count   (user points was not there for now so took static 0 points)
	query := `SELECT 
                u.user_id, 
                u.user_address,
                (SELECT COUNT(*) FROM triggerx.job_data WHERE user_address = u.user_address) as total_jobs,
                (SELECT COUNT(*) 
                 FROM triggerx.task_data t 
                 JOIN triggerx.job_data j ON t.job_id = j.job_id 
                 WHERE j.user_address = u.user_address 
                 AND t.execution_timestamp IS NOT NULL) as tasks_completed,
                COALESCE(u.user_points, 0) as user_points 
              FROM triggerx.user_data u
              WHERE u.status = true
              ORDER BY u.user_points DESC`

	iter := h.db.Session().Query(query).Iter()

	var userLeaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	// Scan all rows into user leaderboard entries
	for iter.Scan(
		&userEntry.UserID,
		&userEntry.UserAddress,
		&userEntry.TotalJobs,
		&userEntry.TasksCompleted,
		&userEntry.UserPoints,
	) {
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
