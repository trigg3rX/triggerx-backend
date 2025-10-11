package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTaskDataByTaskID(c *gin.Context) {
	logger := h.getLogger(c)
	taskID := c.Param("id")
	if taskID == "" {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "No task ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetTaskDataByID] For task ID: %s", taskID)

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetByID(c.Request.Context(), taskIDInt)
	trackDBOp(err)
	if err != nil || taskData == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	logger.Infof("GET [GetTaskDataByTaskID] Successful, task ID: %d", taskIDInt)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("job_id")
	if jobID == "" {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetTasksByJobID] For job ID: %s", jobID)

	// Get job to find task IDs and chain ID
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Get tasks by task IDs
	trackDBOp = metrics.TrackDBOperation("read", "task_data")
	var tasksData []*types.TaskDataEntity
	for _, taskID := range job.TaskIDs {
		task, err := h.taskRepository.GetByID(c.Request.Context(), taskID)
		if err == nil && task != nil {
			tasksData = append(tasksData, task)
		}
	}
	trackDBOp(nil)

	// Convert to GetTasksByJobIDResponse
	explorerBaseURL := getExplorerBaseURL(job.CreatedChainID)
	tasks := make([]types.GetTasksByJobIDResponse, len(tasksData))
	for i, task := range tasksData {
		tasks[i] = types.GetTasksByJobIDResponse{
			TaskNumber:         task.TaskNumber,
			TaskOpXPredictedCost: task.TaskOpxPredictedCost,
			TaskOpXActualCost: task.TaskOpxActualCost,
			ExecutionTimestamp: task.ExecutionTimestamp,
			ExecutionTxHash:    task.ExecutionTxHash,
			TaskPerformerID:    task.TaskPerformerID,
			TaskAttesterIDs:    task.TaskAttesterIDs,
			ConvertedArguments: []string{task.ConvertedArguments},
			IsSuccessful:       task.IsSuccessful,
			IsAccepted:         task.IsAccepted,
			TxURL:              fmt.Sprintf("%s%s", explorerBaseURL, task.ExecutionTxHash),
		}
	}

	logger.Infof("GET [GetTasksByJobID] Successful, job ID: %s, tasks: %d", jobID, len(tasks))
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
