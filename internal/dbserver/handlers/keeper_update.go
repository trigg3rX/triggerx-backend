package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) IncrementKeeperTaskCount(c *gin.Context) {
	logger := h.getLogger(c)
	keeperID := c.Param("id")
	logger.Debugf("POST [IncrementKeeperTaskCount] For keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	// Increment task count
	if keeper.NoExecutedTasks == 0 {
		keeper.NoExecutedTasks = 1
	} else {
		keeper.NoExecutedTasks = keeper.NoExecutedTasks + 1
	}

	trackDBOp = metrics.TrackDBOperation("update", "keeper_data")
	err = h.keeperRepository.Update(c.Request.Context(), keeper)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error updating keeper: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newCount := keeper.NoExecutedTasks
	logger.Debugf("Successfully incremented task count to %d for keeper ID: %s", newCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
}

func (h *Handler) AddTaskFeeToKeeperPoints(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("POST [AddTaskFeeToKeeperPoints] Adding task fee to keeper with ID: %s", c.Param("id"))
	keeperID := c.Param("id")
	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		TaskID int64 `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	taskID := requestBody.TaskID
	logger.Debugf("POST [AddTaskFeeToKeeperPoints] Processing task fee for task ID %d to keeper with ID: %s", taskID, keeperID)

	// Get task to extract fee
	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	task, err := h.taskRepository.GetByID(c.Request.Context(), taskID)
	trackDBOp(err)
	if err != nil || task == nil {
		logger.Errorf("Error retrieving task for task ID %d: %v", taskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	taskFee := &task.TaskOpxActualCost

	// Get keeper
	trackDBOp = metrics.TrackDBOperation("read", "keeper_data")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	// Update keeper points
	keeper.KeeperPoints = types.Add(keeper.KeeperPoints, *taskFee)

	trackDBOp = metrics.TrackDBOperation("update", "keeper_data")
	err = h.keeperRepository.Update(c.Request.Context(), keeper)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error updating keeper points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Debugf("Successfully added task fee from task ID %d to keeper ID: %s",
		taskID, keeperID)
	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": keeper.KeeperPoints,
	})
}

func (h *Handler) UpdateKeeperChatID(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("PUT [UpdateKeeperChatID] Updating keeper chat ID")

	var requestData types.UpdateKeeperChatIDRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_address", strings.ToLower(requestData.KeeperAddress))
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Keeper not found for address: %s, error: %v", requestData.KeeperAddress, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	// Update chat ID
	keeper.ChatID = requestData.ChatID

	trackDBOp = metrics.TrackDBOperation("update", "keeper_data")
	err = h.keeperRepository.Update(c.Request.Context(), keeper)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error updating chat ID for keeper: %s, error: %v", requestData.KeeperAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update chat ID",
			"code":  "UPDATE_ERROR",
		})
		return
	}

	logger.Debugf("Successfully updated chat ID for keeper: %s", requestData.KeeperAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
}
