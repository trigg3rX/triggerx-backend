package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) UpdateTaskExecutionData(c *gin.Context) {
	logger := h.getLogger(c)
	taskID := c.Param("id")
	logger.Debugf("PUT [UpdateTaskExecutionData] For task with ID: %s", taskID)

	var taskData types.UpdateTaskExecutionDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.ExecutionTimestamp.IsZero() || taskData.ExecutionTxHash == "" {
		logger.Errorf("Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	// Get task
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(c.Request.Context(), taskData.TaskID)
	trackDBOp(err)
	if err != nil || task == nil {
		logger.Errorf("Task not found for ID %d: %v", taskData.TaskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	// Update task fields
	task.ExecutionTimestamp = taskData.ExecutionTimestamp
	task.ExecutionTxHash = taskData.ExecutionTxHash
	task.TaskPerformerID = taskData.TaskPerformerID

	trackDBOp = metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.Update(c.Request.Context(), task); err != nil {
		trackDBOp(err)
		logger.Errorf("Error updating task execution data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Task update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	logger.Debugf("Successfully updated task execution data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task execution data updated successfully"})
}

func (h *Handler) UpdateTaskAttestationData(c *gin.Context) {
	logger := h.getLogger(c)
	taskID := c.Param("id")
	logger.Debugf("PUT [UpdateTaskAttestationData] For task with ID: %s", c.Param("id"))

	var taskData types.UpdateTaskAttestationDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.TaskNumber == 0 || len(taskData.TaskAttesterIDs) == 0 || len(taskData.TpSignature) == 0 || len(taskData.TaSignature) == 0 || taskData.TaskSubmissionTxHash == "" {
		logger.Errorf("Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	// Get task
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(c.Request.Context(), taskData.TaskID)
	trackDBOp(err)
	if err != nil || task == nil {
		logger.Errorf("Task not found for ID %d: %v", taskData.TaskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	// Update task fields
	task.TaskNumber = taskData.TaskNumber
	task.TaskAttesterIDs = taskData.TaskAttesterIDs
	task.SubmissionTxHash = taskData.TaskSubmissionTxHash

	trackDBOp = metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.Update(c.Request.Context(), task); err != nil {
		trackDBOp(err)
		logger.Errorf("Error updating task attestation data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Task update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	logger.Debugf("Successfully updated task attestation data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task attestation data updated successfully"})
}
