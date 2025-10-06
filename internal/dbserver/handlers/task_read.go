package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasksData, err := h.taskRepository.GetTasksByJobID(jobID)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for jobID %s: %v", jobID.String(), err)
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
			IsAccepted:         task.IsAccepted,
			TaskStatus:         task.TaskStatus,
			ConvertedArguments:  task.ConvertedArguments ,
		}
	}
	//find the created_chain id for the job using jobIDBig from database
	var createdChainID string
	createdChainID, err = h.taskRepository.GetCreatedChainIDByJobID(jobID)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving created_chain_id for jobID %s: %v", jobID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	// Set tx_url for each task
	explorerBaseURL := getExplorerBaseURL(createdChainID)
	for i := range tasks {
		if tasks[i].ExecutionTxHash != "" {
			tasks[i].TxURL = fmt.Sprintf("%s%s", explorerBaseURL, tasks[i].ExecutionTxHash)
		}
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %s", len(tasks), jobID.String())
	c.JSON(http.StatusOK, tasks)
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
		return "https:/base-sepolia.blockscout.com/tx/"
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
		return "https:/arbitrum.blockscout.com/tx/"
	default:
		return "https://sepolia.etherscan.io/tx/"
	}
}
