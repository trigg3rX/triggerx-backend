package handlers

import (
	"context"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) UpdateTaskExecutionData(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskExecutionData] Updating task execution data for task with ID: %s", taskID)

	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateTaskExecutionData] trace_id=%s - Updating task execution data", traceID)

	var taskData types.UpdateTaskExecutionDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskExecutionData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.ExecutionTimestamp.IsZero() || taskData.ExecutionTxHash == "" {
		h.logger.Errorf("[UpdateTaskExecutionData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	ctx := context.Background()

	// Get task
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(ctx, taskData.TaskID)
	trackDBOp(err)
	if err != nil || task == nil {
		h.logger.Errorf("[UpdateTaskExecutionData] Task not found for ID %d: %v", taskData.TaskID, err)
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
	if err := h.taskRepository.Update(ctx, task); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskExecutionData] Error updating task execution data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Task update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[UpdateTaskExecutionData] Successfully updated task execution data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task execution data updated successfully"})
}

func (h *Handler) UpdateTaskAttestationData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateTaskAttestationData] trace_id=%s - Updating task attestation data", traceID)
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskAttestationData] Updating task attestation data for task with ID: %s", taskID)

	var taskData types.UpdateTaskAttestationDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskAttestationData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.TaskNumber == 0 || len(taskData.TaskAttesterIDs) == 0 || len(taskData.TpSignature) == 0 || len(taskData.TaSignature) == 0 || taskData.TaskSubmissionTxHash == "" {
		h.logger.Errorf("[UpdateTaskAttestationData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	ctx := context.Background()

	// Get task
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(ctx, taskData.TaskID)
	trackDBOp(err)
	if err != nil || task == nil {
		h.logger.Errorf("[UpdateTaskAttestationData] Task not found for ID %d: %v", taskData.TaskID, err)
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
	if err := h.taskRepository.Update(ctx, task); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskAttestationData] Error updating task attestation data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Task update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[UpdateTaskAttestationData] Successfully updated task attestation data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task attestation data updated successfully"})
}

func (h *Handler) UpdateTaskFee(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateTaskFee] trace_id=%s - Updating task fee", traceID)
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskFee] Updating task fee for task with ID: %s", taskID)

	var taskFee struct {
		Fee float64 `json:"fee"`
	}
	if err := c.ShouldBindJSON(&taskFee); err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error parsing task ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	ctx := context.Background()

	// Get task
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(ctx, taskIDInt)
	trackDBOp(err)
	if err != nil || task == nil {
		h.logger.Errorf("[UpdateTaskFee] Task not found for ID %d: %v", taskIDInt, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	// Update task fee
	feeBigInt := big.NewInt(int64(taskFee.Fee))
	task.TaskOpxActualCost = *feeBigInt

	trackDBOp = metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.Update(ctx, task); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskFee] Error updating task fee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Task update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[UpdateTaskFee] Successfully updated task fee for task with ID: %s", taskID)
	c.JSON(http.StatusOK, taskFee)
}
