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
	h.logger.Debug(c.Request.Context(), "[IncrementKeeperTaskCount] Incrementing task count for keeper with ID", observability.String("keeper_id", keeperID))

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
		h.logger.Error(c.Request.Context(), "[IncrementKeeperTaskCount] Error retrieving current task count", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug(c.Request.Context(), "[IncrementKeeperTaskCount] Successfully incremented task count to for keeper ID", observability.Int64("new_count", newCount), observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
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
	h.logger.Debug(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Processing task fee for task ID to keeper with ID", observability.Int64("task_id", taskID), observability.String("keeper_id", keeperID))

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
		h.logger.Error(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Error retrieving current points", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug(c.Request.Context(), "[AddTaskFeeToKeeperPoints] Successfully added task fee from task ID to keeper ID", observability.Float64("task_fee", taskFee), observability.Int64("task_id", taskID), observability.String("keeper_id", keeperID), observability.Float64("new_points", newPoints))
	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": newPoints,
	})
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

	h.logger.Info(c.Request.Context(), "[UpdateKeeperChatID] Successfully updated chat ID for keeper", observability.String("keeper_address", requestData.KeeperAddress))
	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
}
