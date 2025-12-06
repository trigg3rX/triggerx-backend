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
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTaskDataByID] trace_id=%s - Retrieving task data", traceID)
	taskID := c.Param("id")
	if taskID == "" {
		h.logger.Error("[GetTaskDataByID] No task ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No task ID provided",
			"code":  "MISSING_TASK_ID",
		})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Invalid task ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Retrieving task data for task ID: %d", taskIDInt)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error retrieving task data for taskID %d: %v", taskIDInt, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Successfully retrieved task data for task ID: %d", taskIDInt)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTasksByJobID] trace_id=%s - Retrieving tasks for job", traceID)
	jobIDStr := c.Param("job_id")
	if jobIDStr == "" {
		h.logger.Error("[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobID := new(big.Int)
	if _, ok := jobID.SetString(jobIDStr, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Retrieving tasks for job ID: %s | %s", jobIDStr, jobID.String())

	tasks, err := h.fetchTasksForJob(jobID)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for jobID %s: %v", jobID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %s", len(tasks), jobID.String())
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) GetTasksByUserAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTasksByUserAddress] trace_id=%s - Retrieving tasks for user address", traceID)

	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Error("[GetTasksByUserAddress] No user address provided")
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
		h.logger.Errorf("[GetTasksByUserAddress] Error retrieving jobs for user %s: %v", userAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	if len(jobIDs) == 0 {
		h.logger.Infof("[GetTasksByUserAddress] No jobs found for user %s", userAddress)
		c.JSON(http.StatusOK, gin.H{
			"user_address": userAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Errorf("[GetTasksByUserAddress] Error retrieving tasks for user %s: %v", userAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested user",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTasksByUserAddress] Successfully retrieved tasks for %d jobs owned by %s", len(taskGroups), userAddress)
	c.JSON(http.StatusOK, gin.H{
		"user_address": userAddress,
		"task_groups":  taskGroups,
	})
}

func (h *Handler) GetTasksByApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTasksByApiKey] trace_id=%s - Retrieving tasks for API key owner", traceID)

	requestedAPIKey := strings.TrimSpace(c.Param("api_key"))
	if requestedAPIKey == "" {
		h.logger.Error("[GetTasksByApiKey] Missing api_key query param")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "api_key query param is required",
			"code":  "MISSING_API_KEY",
		})
		return
	}

	h.logger.Infof("[GetTasksByApiKey] Using api_key parameter for lookup")

	trackDBOp := metrics.TrackDBOperation("read", "apikeys")
	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(requestedAPIKey)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTasksByApiKey] Invalid API key: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
			"code":  "INVALID_API_KEY",
		})
		return
	}

	if apiKeyData.Owner == "" {
		h.logger.Error("[GetTasksByApiKey] No owner associated with API key")
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
	h.logger.Infof("[GetTasksBySafeAddress] trace_id=%s - Retrieving tasks for safe address", traceID)

	safeAddress := strings.ToLower(c.Param("safe_address"))
	if safeAddress == "" {
		h.logger.Error("[GetTasksBySafeAddress] No safe address provided")
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
		h.logger.Errorf("[GetTasksBySafeAddress] Error retrieving jobs for safe address %s: %v", safeAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Safe address not found",
			"code":  "SAFE_ADDRESS_NOT_FOUND",
		})
		return
	}

	if len(jobs) == 0 {
		h.logger.Infof("[GetTasksBySafeAddress] No jobs found for safe address %s", safeAddress)
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
		h.logger.Infof("[GetTasksBySafeAddress] No valid job IDs found for safe address %s", safeAddress)
		c.JSON(http.StatusOK, gin.H{
			"safe_address": safeAddress,
			"task_groups":  []types.TasksByJobGroupResponse{},
		})
		return
	}

	taskGroups, err := h.getTasksGroupedByJob(jobIDs)
	if err != nil {
		h.logger.Errorf("[GetTasksBySafeAddress] Error retrieving tasks for safe address %s: %v", safeAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for the requested safe address",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTasksBySafeAddress] Successfully retrieved tasks for %d jobs linked to %s", len(taskGroups), safeAddress)
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
	h.logger.Infof("[GetRecentTasks] trace_id=%s - Retrieving recent tasks", traceID)

	// Parse limit from query parameter, default to 200
	limitStr := c.DefaultQuery("limit", "200")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		h.logger.Errorf("[GetRecentTasks] Invalid limit parameter: %s", limitStr)
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

	h.logger.Infof("[GetRecentTasks] Fetching recent tasks with limit: %d", limit)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetRecentTasks(limit)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetRecentTasks] Error retrieving recent tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve recent tasks",
			"code":  "TASKS_FETCH_ERROR",
		})
		return
	}

	h.logger.Infof("[GetRecentTasks] Successfully retrieved %d recent tasks", len(tasks))
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
		"limit": limit,
	})
}
