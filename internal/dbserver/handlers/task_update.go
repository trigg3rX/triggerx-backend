package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) UpdateTaskExecutionData(c *gin.Context) {
	taskID := c.Param("id")

	var taskData types.UpdateTaskExecutionDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateTaskExecutionData] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.ExecutionTimestamp.IsZero() || taskData.ExecutionTxHash == "" {
		h.logger.Error(c.Request.Context(), "[UpdateTaskExecutionData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskExecutionDataInDB(c.Request.Context(), &taskData); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "[UpdateTaskExecutionData] Error updating task execution data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, gin.H{"message": "Task execution data updated successfully"})
	h.logger.Info(c.Request.Context(), "[UpdateTaskExecutionData] Updated task execution data", observability.String("task_id", taskID))
}

func (h *Handler) UpdateTaskAttestationData(c *gin.Context) {
	taskID := c.Param("id")

	var taskData types.UpdateTaskAttestationDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateTaskAttestationData] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.TaskNumber == 0 || len(taskData.TaskAttesterIDs) == 0 || len(taskData.TpSignature) == 0 || len(taskData.TaSignature) == 0 || taskData.TaskSubmissionTxHash == "" {
		h.logger.Error(c.Request.Context(), "[UpdateTaskAttestationData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskAttestationDataInDB(c.Request.Context(), &taskData); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "[UpdateTaskAttestationData] Error updating task attestation data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, gin.H{"message": "Task attestation data updated successfully"})
	h.logger.Info(c.Request.Context(), "[UpdateTaskAttestationData] Updated task attestation data", observability.String("task_id", taskID))
}

func (h *Handler) UpdateTaskFee(c *gin.Context) {
	taskID := c.Param("id")

	var taskFee struct {
		Fee float64 `json:"fee"`
	}
	if err := c.ShouldBindJSON(&taskFee); err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateTaskFee] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateTaskFee] Error parsing task ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "task_data")
	if err := h.taskRepository.UpdateTaskFee(c.Request.Context(), taskIDInt, taskFee.Fee); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "[UpdateTaskFee] Error updating task fee", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found or update failed",
			"code":  "TASK_UPDATE_ERROR",
		})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, taskFee)
	h.logger.Info(c.Request.Context(), "[UpdateTaskFee] Updated task fee", observability.String("task_id", taskID))
}
