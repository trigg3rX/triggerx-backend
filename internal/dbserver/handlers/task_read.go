package handlers

import (
	"math/big"
	"net/http"
	"strconv"

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
	jobID := c.Param("job_id")
	if jobID == "" {
		h.logger.Error("[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobIDBig := new(big.Int)
	_, ok := jobIDBig.SetString(jobID, 10)
	if !ok {
		h.logger.Errorf("[GetTasksByJobID] Invalid job ID format: %v", jobID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Retrieving tasks for job ID: %s", jobIDBig.String())

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasksData, err := h.taskRepository.GetTasksByJobID(jobIDBig)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for jobID %s: %v", jobIDBig.String(), err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	// Convert GetTasksByJobID to TasksByJobIDResponse
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
			IsSuccessful:       task.IsSuccessful,
			TaskStatus:         task.TaskStatus,
		}
	}
	//find the created_chain id for the job using jobIDBig from database
	var createdChainID string
	createdChainID, err = h.taskRepository.GetCreatedChainIDByJobID(jobIDBig)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving created_chain_id for jobID %s: %v", jobIDBig.String(), err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	// Set tx_url for each task
	blockscoutBaseURL := getBlockscoutBaseURL(createdChainID)
	for i := range tasks {
		if tasks[i].ExecutionTxHash != "" {
			tasks[i].TxURL = blockscoutBaseURL + tasks[i].ExecutionTxHash
		}
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %s", len(tasks), jobIDBig.String())
	c.JSON(http.StatusOK, tasks)
}

// Helper function to get Blockscout base URL from chain ID
func getBlockscoutBaseURL(chainID string) string {
	switch chainID {
	case "1337":
		return "https://sepolia.etherscan.io/tx/"
	case "11155420": // OP Sepolia
		return "https://sepolia-optimism.etherscan.io/tx/"
	case "84532": // Base Sepolia
		return "https://sepolia.basescan.org/tx/"
	default:
		return "https://sepolia.etherscan.io/tx/"
	}
}
