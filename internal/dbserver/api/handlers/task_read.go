package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.logger.Error(c.Request.Context(), "[GetTaskDataByID] No task ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No task ID provided",
			"code":  "MISSING_TASK_ID",
		})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTaskDataByID] Invalid task ID format", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTaskDataByID] Failed to retrieve task data", observability.Int64("task_id", taskIDInt), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, taskData)
	h.logger.Debug(c.Request.Context(), "[GetTaskDataByID] Retrieved task data", observability.Int64("task_id", taskIDInt))
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	if jobIDStr == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	tasks, err := h.fetchTasksForJob(jobIDStr)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTasksByJobID] Failed to retrieve tasks", observability.String("job_id", jobIDStr), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
	h.logger.Debug(c.Request.Context(), "[GetTasksByJobID] Retrieved tasks", observability.String("job_id", jobIDStr), observability.Int("tasks_count", len(tasks)))
}

func (h *Handler) GetRecentTasks(c *gin.Context) {
	// Parse limit from query parameter, default to 200
	limitStr := c.DefaultQuery("limit", "200")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		h.logger.Error(c.Request.Context(), "[GetRecentTasks] Invalid limit parameter", observability.String("limit_str", limitStr))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
			"code":  "INVALID_LIMIT",
		})
		return
	}

	// Enforce maximum limit of 200
	if limit > 200 {
		limit = 200
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetRecentTasks(limit)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetRecentTasks] Failed to retrieve recent tasks", observability.Int("limit", limit), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve recent tasks",
			"code":  "TASKS_FETCH_ERROR",
		})
		return
	}

	// Fetch global statistics
	trackStatsOp := metrics.TrackDBOperation("read", "statistics")
	statistics, err := h.taskRepository.GetGlobalStatistics()
	trackStatsOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetRecentTasks] Failed to retrieve statistics", observability.Error(err))
		// Continue without statistics rather than failing the entire request
		statistics = nil
	}

	c.JSON(http.StatusOK, types.RecentTasksResponse{
		Tasks:      tasks,
		Count:      len(tasks),
		Limit:      limit,
		Statistics: statistics,
	})
	h.logger.Debug(c.Request.Context(), "[GetRecentTasks] Retrieved recent tasks with statistics", observability.Int("tasks_count", len(tasks)), observability.Int("limit", limit))
}

func (h *Handler) GetTasksByUserAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByUserAddress] No user address provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No user address provided",
			"code":  "MISSING_USER_ADDRESS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	jobIDs, err := h.userRepository.GetUserJobIDs(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTasksByUserAddress] Failed to retrieve jobs for user", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"user_address": userAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTasksByUserAddress] Failed to retrieve tasks", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested user",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_address": userAddress,
		"task_groups":  taskGroups,
	})
	h.logger.Debug(c.Request.Context(), "[GetTasksByUserAddress] Retrieved tasks", observability.String("user_address", userAddress), observability.Int("task_groups_count", len(taskGroups)))
}

func (h *Handler) GetTasksByApiKey(c *gin.Context) {
	requestedAPIKey := strings.TrimSpace(c.Param("api_key"))
	if requestedAPIKey == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByApiKey] Missing api_key query param")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "api_key query param is required",
			"code":  "MISSING_API_KEY",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "apikeys")
	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(requestedAPIKey)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTasksByApiKey] Invalid API key", observability.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
			"code":  "INVALID_API_KEY",
		})
		return
	}

	if apiKeyData.Owner == "" {
		h.logger.Warn(c.Request.Context(), "[GetTasksByApiKey] No owner associated with API key")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No owner associated with API key",
			"code":  "OWNER_NOT_FOUND",
		})
		return
	}

	c.Params = append(c.Params, gin.Param{Key: "user_address", Value: strings.ToLower(apiKeyData.Owner)})
	h.GetTasksByUserAddress(c)
}

func (h *Handler) GetTasksBySafeAddress(c *gin.Context) {
	safeAddress := strings.ToLower(c.Param("safe_address"))
	if safeAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksBySafeAddress] No safe address provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No safe address provided",
			"code":  "MISSING_SAFE_ADDRESS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	jobs, err := h.jobRepository.GetJobsBySafeAddress(safeAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksBySafeAddress] Error retrieving jobs for safe address", observability.String("safe_address", safeAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Safe address not found",
			"code":  "SAFE_ADDRESS_NOT_FOUND",
		})
		return
	}

	if len(jobs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"safe_address": safeAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	jobIDs := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if job.JobID != "" {
			jobIDs = append(jobIDs, job.JobID)
		}
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"safe_address": safeAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTasksBySafeAddress] Failed to retrieve tasks", observability.String("safe_address", safeAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested safe address",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"safe_address": safeAddress,
		"task_groups":  taskGroups,
	})
	h.logger.Debug(c.Request.Context(), "[GetTasksBySafeAddress] Retrieved tasks", observability.String("safe_address", safeAddress), observability.Int("task_groups_count", len(taskGroups)))
}

func (h *Handler) getTasksGroupedByJob(jobIDs []string) ([]types.TasksByJobGroupResponse, error) {
	taskGroups := make([]types.TasksByJobGroupResponse, 0, len(jobIDs))
	seen := make(map[string]struct{})

	for _, jobID := range jobIDs {
		if jobID == "" {
			continue
		}

		if _, exists := seen[jobID]; exists {
			continue
		}
		seen[jobID] = struct{}{}

		tasks, err := h.fetchTasksForJob(jobID)
		if err != nil {
			return nil, err
		}

		taskGroups = append(taskGroups, types.TasksByJobGroupResponse{
			JobID: jobID,
			Tasks: tasks,
		})
	}

	return taskGroups, nil
}

func (h *Handler) fetchTasksForJob(jobID string) ([]types.TasksByJobIDResponse, error) {
	trackTasksOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetTasksByJobID(jobID)
	trackTasksOp(err)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
