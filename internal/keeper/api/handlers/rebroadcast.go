package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// RebroadcastTask handles task rebroadcast requests
// POST /task/rebroadcast
// Body: { "task_id": 123 }
func (h *TaskHandler) RebroadcastTask(c *gin.Context) {
	ctx := c.Request.Context()

	var request struct {
		TaskID int64 `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info(ctx, "Received rebroadcast request",
		observability.Int64("task_id", request.TaskID))

	// Call executor to rebroadcast
	if err := h.executor.RebroadcastTask(ctx, request.TaskID); err != nil {
		h.logger.Error(ctx, "Failed to rebroadcast task",
			observability.Int64("task_id", request.TaskID),
			observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rebroadcast task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task rebroadcast initiated",
		"task_id": request.TaskID,
	})
}
