package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) IncrementKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[IncrementKeeperTaskCount] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "keeper_data")
	newCount, err := h.keeperRepository.IncrementKeeperTaskCount(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[IncrementKeeperTaskCount] Failed to increment task count", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
	h.logger.Info(c.Request.Context(), "[IncrementKeeperTaskCount] Incremented task count", observability.String("keeper_id", keeperID), observability.Int64("new_count", newCount))
}

func (h *Handler) AddTaskFeeToKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")
	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		TaskID int64 `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.logger.Error(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	taskID := requestBody.TaskID

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskFee, err := h.taskRepository.GetTaskFee(taskID)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	trackDBOp = metrics.TrackDBOperation("update", "keeper_data")
	newPoints, err := h.keeperRepository.UpdateKeeperPoints(keeperIDInt, taskFee)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Failed to update keeper points", observability.String("keeper_id", keeperID), observability.Int64("task_id", taskID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": newPoints,
	})
	h.logger.Info(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Added task fee to keeper points", observability.String("keeper_id", keeperID), observability.Int64("task_id", taskID), observability.Float64("new_points", newPoints))
}

func (h *Handler) UpdateKeeperChatID(c *gin.Context) {

	var requestData types.UpdateKeeperChatIDRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateKeeperChatID] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "keeper_data")
	err := h.keeperRepository.UpdateKeeperChatID(requestData.KeeperAddress, requestData.ChatID)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateKeeperChatID] Error updating chat ID for keeper", observability.String("keeper_address", requestData.KeeperAddress))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
	h.logger.Info(c.Request.Context(), "[UpdateKeeperChatID] Updated chat ID", observability.String("keeper_address", requestData.KeeperAddress))
}
