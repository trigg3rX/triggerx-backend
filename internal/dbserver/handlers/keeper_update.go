package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

func (h *Handler) IncrementKeeperTaskCount(c *gin.Context) {
	traceID := c.GetString("trace_id")
	h.logger.Infof("[IncrementKeeperTaskCount][trace_id: %s] Incrementing task count for keeper with ID: %s", traceID, c.Param("id"))
	keeperID := c.Param("id")
	h.logger.Infof("[IncrementKeeperTaskCount] Incrementing task count for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error parsing keeper ID: %v", err)
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
		h.logger.Errorf("[IncrementKeeperTaskCount] Error retrieving current task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[IncrementKeeperTaskCount] Successfully incremented task count to %d for keeper ID: %s", newCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
}

func (h *Handler) AddTaskFeeToKeeperPoints(c *gin.Context) {
	traceID := c.GetString("trace_id")
	h.logger.Infof("[AddTaskFeeToKeeperPoints][trace_id: %s] Adding task fee to keeper with ID: %s", traceID, c.Param("id"))
	keeperID := c.Param("id")
	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		TaskID int64 `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	taskID := requestBody.TaskID
	h.logger.Infof("[AddTaskFeeToKeeperPoints] Processing task fee for task ID %d to keeper with ID: %s", taskID, keeperID)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskFee, err := h.taskRepository.GetTaskFee(taskID)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID %d: %v", taskID, err)
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
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving current points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[AddTaskFeeToKeeperPoints] Successfully added task fee %f from task ID %d to keeper ID: %s, new points: %f",
		taskFee, taskID, keeperID, newPoints)
	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": newPoints,
	})
}

func (h *Handler) UpdateKeeperChatID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateKeeperChatID] trace_id=%s - Updating keeper chat ID", traceID)

	var requestData types.UpdateKeeperChatIDRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error decoding request body: %v", err)
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
		h.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper: %s", requestData.KeeperAddress)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Successfully updated chat ID for keeper: %s", requestData.KeeperAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
}
