package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTaskDataByID] trace_id=%s - Retrieving task data", observability.String("trace_id", traceID))
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
		h.logger.Error(c.Request.Context(), "[GetTaskDataByID] Invalid task ID format: %v", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTaskDataByID] Retrieving task data for task ID: %d", observability.Int64("task_id", taskIDInt))

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTaskDataByID] Error retrieving task data for taskID %d: %v", observability.Int64("task_id", taskIDInt), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTaskDataByID] Successfully retrieved task data for task ID: %d", observability.Int64("task_id", taskIDInt))
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTasksByJobID] trace_id=%s - Retrieving tasks for job", observability.String("trace_id", traceID))
	jobIDStr := c.Param("job_id")
	if jobIDStr == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobID := new(big.Int)
	if _, ok := jobID.SetString(jobIDStr, 10); !ok {
		h.logger.Error(c.Request.Context(), "[GetTasksByJobID] Invalid job_id format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTasksByJobID] Retrieving tasks for job ID: %s | %s", observability.String("job_id", jobIDStr), observability.String("job_id_big_int", jobID.String()))

	tasks, err := h.fetchTasksForJob(jobID)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksByJobID] Error retrieving tasks for jobID %s: %v", observability.String("job_id", jobID.String()), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %s", observability.Int("tasks_count", len(tasks)), observability.String("job_id", jobID.String()))
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) GetTasksByUserAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTasksByUserAddress] trace_id=%s - Retrieving tasks for user address", observability.String("trace_id", traceID))

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
	_, jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksByUserAddress] Error retrieving jobs for user %s: %v", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	if len(jobIDs) == 0 {
		h.logger.Info(c.Request.Context(), "[GetTasksByUserAddress] No jobs found for user %s", observability.String("user_address", userAddress))
		c.JSON(http.StatusOK, gin.H{
			"user_address": userAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksByUserAddress] Error retrieving tasks for user %s: %v", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested user",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTasksByUserAddress] Successfully retrieved tasks for %d jobs owned by %s", observability.Int("task_groups_count", len(taskGroups)), observability.String("user_address", userAddress))
	c.JSON(http.StatusOK, gin.H{
		"user_address": userAddress,
		"task_groups":  taskGroups,
	})
}

func (h *Handler) GetTasksByApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTasksByApiKey] trace_id=%s - Retrieving tasks for API key owner", observability.String("trace_id", traceID))

	requestedAPIKey := strings.TrimSpace(c.Param("api_key"))
	if requestedAPIKey == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByApiKey] Missing api_key query param")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "api_key query param is required",
			"code":  "MISSING_API_KEY",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTasksByApiKey] Using api_key parameter for lookup")

	trackDBOp := metrics.TrackDBOperation("read", "apikeys")
	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(requestedAPIKey)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksByApiKey] Invalid API key: %v", observability.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
			"code":  "INVALID_API_KEY",
		})
		return
	}

	if apiKeyData.Owner == "" {
		h.logger.Error(c.Request.Context(), "[GetTasksByApiKey] No owner associated with API key")
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
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTasksBySafeAddress] trace_id=%s - Retrieving tasks for safe address", observability.String("trace_id", traceID))

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
		h.logger.Error(c.Request.Context(), "[GetTasksBySafeAddress] Error retrieving jobs for safe address %s: %v", observability.String("safe_address", safeAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Safe address not found",
			"code":  "SAFE_ADDRESS_NOT_FOUND",
		})
		return
	}

	if len(jobs) == 0 {
		h.logger.Info(c.Request.Context(), "[GetTasksBySafeAddress] No jobs found for safe address %s", observability.String("safe_address", safeAddress))
		c.JSON(http.StatusOK, gin.H{
			"safe_address": safeAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	jobIDs := make([]*big.Int, 0, len(jobs))
	for _, job := range jobs {
		if job.JobID == nil {
			continue
		}
		jobIDs = append(jobIDs, job.JobID.ToBigInt())
	}

	if len(jobIDs) == 0 {
		h.logger.Info(c.Request.Context(), "[GetTasksBySafeAddress] No valid job IDs found for safe address %s", observability.String("safe_address", safeAddress))
		c.JSON(http.StatusOK, gin.H{
			"safe_address": safeAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTasksBySafeAddress] Error retrieving tasks for safe address %s: %v", observability.String("safe_address", safeAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested safe address",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetTasksBySafeAddress] Successfully retrieved tasks for %d jobs linked to %s", observability.Int("task_groups_count", len(taskGroups)), observability.String("safe_address", safeAddress))
	c.JSON(http.StatusOK, gin.H{
		"safe_address": safeAddress,
		"task_groups":  taskGroups,
	})
}

// Helper function to get Explorer base URL from chain ID
func getExplorerBaseURL(chainID string) string {
	switch chainID {
	// Testnets
	case "11155111":
		return "https://eth-sepolia.blockscout.com/tx/"
	case "11155420": // OP Sepolia
		return "https://testnet-explorer.optimism.io/tx/"
	case "84532": // Base Sepolia
		return "https://base-sepolia.blockscout.com/tx/"
	case "421614": // Arbitrum Sepolia
		return "https://arbitrum-sepolia.blockscout.com/tx/"

	// Mainnets
	case "1": // Ethereum Mainnet
		return "https://eth.blockscout.com/tx/"
	case "10": // Optimism Mainnet
		return "https://explorer.optimism.io/tx/"
	case "8453": // Base Mainnet
		return "https://base.blockscout.com/tx/"
	case "42161": // Arbitrum Mainnet
		return "https://arbitrum.blockscout.com/tx/"
	default:
		return "https://sepolia.etherscan.io/tx/"
	}
}

func (h *Handler) getTasksGroupedByJob(jobIDs []*big.Int) ([]types.TasksByJobGroupResponse, error) {
	taskGroups := make([]types.TasksByJobGroupResponse, 0, len(jobIDs))
	seen := make(map[string]struct{})

	for _, jobID := range jobIDs {
		if jobID == nil {
			continue
		}

		jobIDStr := jobID.String()
		if _, exists := seen[jobIDStr]; exists {
			continue
		}
		seen[jobIDStr] = struct{}{}

		tasks, err := h.fetchTasksForJob(jobID)
		if err != nil {
			return nil, err
		}

		taskGroups = append(taskGroups, types.TasksByJobGroupResponse{
			JobID: jobIDStr,
			Tasks: tasks,
		})
	}

	return taskGroups, nil
}

func (h *Handler) fetchTasksForJob(jobID *big.Int) ([]types.TasksByJobIDResponse, error) {
	trackTasksOp := metrics.TrackDBOperation("read", "task_data")
	tasksData, err := h.taskRepository.GetTasksByJobID(jobID)
	trackTasksOp(err)
	if err != nil {
		return nil, err
	}

	tasks := convertTasksData(tasksData)

	trackChainOp := metrics.TrackDBOperation("read", "job_data")
	createdChainID, err := h.taskRepository.GetCreatedChainIDByJobID(jobID)
	trackChainOp(err)
	if err != nil {
		return nil, err
	}

	explorerBaseURL := getExplorerBaseURL(createdChainID)
	for i := range tasks {
		if tasks[i].ExecutionTxHash != "" {
			tasks[i].TxURL = fmt.Sprintf("%s%s", explorerBaseURL, tasks[i].ExecutionTxHash)
		}
	}

	return tasks, nil
}

func convertTasksData(tasksData []types.GetTasksByJobID) []types.TasksByJobIDResponse {
	tasks := make([]types.TasksByJobIDResponse, len(tasksData))
	for i, task := range tasksData {
		tasks[i] = types.TasksByJobIDResponse{
			TaskID:             task.TaskID,
			TaskNumber:         task.TaskNumber,
			TaskOpXCost:        task.TaskOpXCost,
			ExecutionTimestamp: task.ExecutionTimestamp,
			ExecutionTxHash:    task.ExecutionTxHash,
			TaskPerformerID:    task.TaskPerformerID,
			TaskAttesterIDs:    task.TaskAttesterIDs,
			IsAccepted:         task.IsAccepted,
			TaskStatus:         task.TaskStatus,
			TaskError:          task.TaskError,
			ConvertedArguments: task.ConvertedArguments,
		}
	}
	return tasks
}

func (h *Handler) GetRecentTasks(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetRecentTasks] trace_id=%s - Retrieving recent tasks", observability.String("trace_id", traceID))

	// Parse limit from query parameter, default to 200
	limitStr := c.DefaultQuery("limit", "200")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		h.logger.Error(c.Request.Context(), "[GetRecentTasks] Invalid limit parameter: %s", observability.String("limit_str", limitStr))
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

	h.logger.Info(c.Request.Context(), "[GetRecentTasks] Fetching recent tasks with limit: %d", observability.Int("limit", limit))

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetRecentTasks(limit)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetRecentTasks] Error retrieving recent tasks: %v", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve recent tasks",
			"code":  "TASKS_FETCH_ERROR",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetRecentTasks] Successfully retrieved %d recent tasks", observability.Int("tasks_count", len(tasks)))
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
		"limit": limit,
	})
}
