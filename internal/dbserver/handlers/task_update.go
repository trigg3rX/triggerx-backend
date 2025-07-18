package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
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

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskExecutionDataInDB(&taskData); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskExecutionData] Error updating task execution data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
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

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskAttestationDataInDB(&taskData); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskAttestationData] Error updating task attestation data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
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

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskFee(taskIDInt, taskFee.Fee); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateTaskFee] Error updating task fee: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[UpdateTaskFee] Successfully updated task fee for task with ID: %s", taskID)
	c.JSON(http.StatusOK, taskFee)
}
