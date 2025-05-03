package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetKeeperLeaderboard retrieves the leaderboard data for all keepers with pagination
func (h *Handler) GetKeeperLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetKeeperLeaderboard] Fetching keeper leaderboard data")

	// Get pagination parameters from query string
	pageSize := 10  // Default page size
	pageNumber := 1 // Default page number

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if size, err := strconv.Atoi(pageSizeStr); err == nil && size > 0 {
			pageSize = size
		}
	}

	if pageNumberStr := r.URL.Query().Get("page"); pageNumberStr != "" {
		if page, err := strconv.Atoi(pageNumberStr); err == nil && page > 0 {
			pageNumber = page
		}
	}

	h.logger.Infof("[GetKeeperLeaderboard] Pagination parameters - Page: %d, PageSize: %d", pageNumber, pageSize)

	// Get total count first
	var totalCount int
	if err := h.db.Session().Query(`
		SELECT COUNT(*) FROM triggerx.keeper_data 
		WHERE status = true ALLOW FILTERING`).Scan(&totalCount); err != nil {
		h.logger.Errorf("[GetKeeperLeaderboard] Error counting keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperLeaderboard] Total keepers found: %d", totalCount)

	// Calculate total pages
	totalPages := (totalCount + pageSize - 1) / pageSize
	h.logger.Infof("[GetKeeperLeaderboard] Total pages calculated: %d", totalPages)

	// Validate page number
	if pageNumber > totalPages {
		h.logger.Warnf("[GetKeeperLeaderboard] Requested page %d is greater than total pages %d, adjusting to last page", pageNumber, totalPages)
		pageNumber = totalPages
	}

	// Query with pagination
	query := `SELECT keeper_id, keeper_address, keeper_name, no_exctask, keeper_points 
              FROM triggerx.keeper_data 
              WHERE status = true ALLOW FILTERING`

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var pageState []byte
	var currentPage int = 1

	// Keep fetching pages until we reach the desired page
	for currentPage <= pageNumber {
		iter := h.db.Session().Query(query).PageSize(pageSize).PageState(pageState).Iter()

		var keeperEntry types.KeeperLeaderboardEntry
		for iter.Scan(
			&keeperEntry.KeeperID,
			&keeperEntry.KeeperAddress,
			&keeperEntry.KeeperName,
			&keeperEntry.TasksExecuted,
			&keeperEntry.KeeperPoints,
		) {
			if currentPage == pageNumber {
				keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
			}
		}

		if err := iter.Close(); err != nil {
			h.logger.Errorf("[GetKeeperLeaderboard] Error fetching keeper leaderboard data: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the page state for the next iteration
		pageState = iter.PageState()
		if pageState == nil {
			break // No more pages
		}
		currentPage++
	}

	h.logger.Infof("[GetKeeperLeaderboard] Retrieved %d keepers for page %d", len(keeperLeaderboard), pageNumber)

	response := struct {
		Data       []types.KeeperLeaderboardEntry `json:"data"`
		Page       int                            `json:"page"`
		PageSize   int                            `json:"page_size"`
		TotalPages int                            `json:"total_pages"`
		TotalItems int                            `json:"total_items"`
	}{
		Data:       keeperLeaderboard,
		Page:       pageNumber,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalCount,
	}

	h.logger.Infof("[GetKeeperLeaderboard] Successfully retrieved keeper leaderboard data for page %d", pageNumber)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserLeaderboard retrieves the leaderboard data for all users with pagination
func (h *Handler) GetUserLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetUserLeaderboard] Fetching user leaderboard data")

	// Get pagination parameters from query string
	pageSize := 10  // Default page size
	pageNumber := 1 // Default page number

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if size, err := strconv.Atoi(pageSizeStr); err == nil && size > 0 {
			pageSize = size
		}
	}

	if pageNumberStr := r.URL.Query().Get("page"); pageNumberStr != "" {
		if page, err := strconv.Atoi(pageNumberStr); err == nil && page > 0 {
			pageNumber = page
		}
	}

	h.logger.Infof("[GetUserLeaderboard] Pagination parameters - Page: %d, PageSize: %d", pageNumber, pageSize)

	// Get total count first
	var totalCount int
	if err := h.db.Session().Query(`
		SELECT COUNT(*) FROM triggerx.user_data`).Scan(&totalCount); err != nil {
		h.logger.Errorf("[GetUserLeaderboard] Error counting users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetUserLeaderboard] Total users found: %d", totalCount)

	// Calculate total pages
	totalPages := (totalCount + pageSize - 1) / pageSize
	h.logger.Infof("[GetUserLeaderboard] Total pages calculated: %d", totalPages)

	// Validate page number
	if pageNumber > totalPages {
		h.logger.Warnf("[GetUserLeaderboard] Requested page %d is greater than total pages %d, adjusting to last page", pageNumber, totalPages)
		pageNumber = totalPages
	}

	// CQL query to get user leaderboard data with pagination
	query := `SELECT user_id, user_address, user_points 
              FROM triggerx.user_data`

	var userLeaderboard []types.UserLeaderboardEntry
	var pageState []byte
	var currentPage int = 1

	// Keep fetching pages until we reach the desired page
	for {
		iter := h.db.Session().Query(query).PageSize(pageSize).PageState(pageState).Iter()

		var userEntry types.UserLeaderboardEntry
		for iter.Scan(
			&userEntry.UserID,
			&userEntry.UserAddress,
			&userEntry.UserPoints,
		) {
			if currentPage == pageNumber {
				// Count total jobs for the user
				jobCountQuery := `SELECT COUNT(*) FROM triggerx.job_data WHERE user_address = ? ALLOW FILTERING`
				var totalJobs int
				if err := h.db.Session().Query(jobCountQuery, userEntry.UserAddress).Scan(&totalJobs); err != nil {
					h.logger.Errorf("[GetUserLeaderboard] Error counting jobs for user %s: %v", userEntry.UserAddress, err)
					totalJobs = 0 // Default to 0 if there's an error
				}
				userEntry.TotalJobs = int64(totalJobs)

				// Count tasks completed for the user
				var tasksCompleted int
				tasksCountQuery := `SELECT COUNT(*) FROM triggerx.task_data WHERE user_address = ? AND execution_timestamp IS NOT NULL ALLOW FILTERING`
				if err := h.db.Session().Query(tasksCountQuery, userEntry.UserAddress).Scan(&tasksCompleted); err != nil {
					h.logger.Errorf("[GetUserLeaderboard] Error counting tasks for user %s: %v", userEntry.UserAddress, err)
					tasksCompleted = 0 // Default to 0 if there's an error
				}
				userEntry.TasksCompleted = int64(tasksCompleted)

				userLeaderboard = append(userLeaderboard, userEntry)
			}
		}

		if err := iter.Close(); err != nil {
			h.logger.Errorf("[GetUserLeaderboard] Error fetching user leaderboard data: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the page state for the next iteration
		pageState = iter.PageState()

		// If we've reached the desired page or there are no more pages, break
		if currentPage == pageNumber || pageState == nil {
			break
		}

		currentPage++
	}

	h.logger.Infof("[GetUserLeaderboard] Retrieved %d users for page %d", len(userLeaderboard), pageNumber)

	response := struct {
		Data       []types.UserLeaderboardEntry `json:"data"`
		Page       int                          `json:"page"`
		PageSize   int                          `json:"page_size"`
		TotalPages int                          `json:"total_pages"`
		TotalItems int                          `json:"total_items"`
	}{
		Data:       userLeaderboard,
		Page:       pageNumber,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalCount,
	}

	h.logger.Infof("[GetUserLeaderboard] Successfully retrieved user leaderboard data for page %d", pageNumber)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
