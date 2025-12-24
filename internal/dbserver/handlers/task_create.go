package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) CreateTaskData(c *gin.Context) {
	var taskData types.CreateTaskDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Error(c.Request.Context(), "[CreateTaskData] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("create", "task_data")
	taskID, err := h.taskRepository.CreateTaskDataInDB(c.Request.Context(), &taskData)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[CreateTaskData] Error creating task", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"code":  "TASK_CREATION_ERROR",
		})
		return
	}

	trackDBOp = metrics.TrackDBOperation("update", "add_task_id")
	err = h.taskRepository.AddTaskIDToJob(taskData.JobID, taskID)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[CreateTaskData] Error adding task ID to job", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add task ID to job",
			"code":  "TASK_ID_ADDITION_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
	h.logger.Info(c.Request.Context(), "[CreateTaskData] Created task", observability.Int64("task_id", taskID))
}
