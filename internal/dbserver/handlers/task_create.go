package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) CreateTaskData(c *gin.Context) {
	var taskData types.CreateTaskDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[CreateTaskData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("create", "task_data")
	taskID, err := h.taskRepository.CreateTaskDataInDB(&taskData)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"code":  "TASK_CREATION_ERROR",
		})
		return
	}

	h.logger.Infof("[CreateTaskData] Successfully created task with ID: %d", taskID)
	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}
