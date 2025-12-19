package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ExecuteTask handles task execution requests with async processing.
// Returns 202 Accepted immediately after validation, then executes task in background.
// Task completion status is reported to TaskMonitor (not back to caller).
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info("Received task execution request", "trace_id", traceID)

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	var requestBody struct {
		Data string `json:"data"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Decode hex data
	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

	decodedData, err := hex.DecodeString(hexData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hex data"})
		return
	}

	decodedDataString := string(decodedData)

	var requestData types.SendTaskDataToKeeper
	if err := json.Unmarshal([]byte(decodedDataString), &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse JSON data",
		})
		return
	}

	// Check if this performer should handle the task
	if !strings.EqualFold(config.GetKeeperAddress(), requestData.PerformerData.KeeperAddress) {
		h.logger.Infof("I am not the performer: %s", requestData.PerformerData.KeeperAddress)
		c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
		return
	}

	h.logger.Infof("I am the performer: %s", requestData.PerformerData.KeeperAddress)

	// Log task info
	taskIDs := requestData.TaskID
	h.logger.Info("Task accepted for async execution", "task_ids", taskIDs, "trace_id", traceID)
	for _, task := range requestData.TargetData {
		h.logger.Infof("Task ID: %d | Target Chain ID: %s", task.TaskID, task.TargetChainID)
	}

	// Return 202 Accepted immediately - task will be processed asynchronously
	// TaskMonitor will receive the completion status via reportTaskStatus
	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"task_ids": taskIDs,
		"trace_id": traceID,
		"message":  "Task accepted for processing. Status will be reported to TaskMonitor.",
	})

	// Execute task asynchronously in a goroutine
	// Make a copy of requestData to avoid race conditions
	taskData := requestData
	go h.executeTaskAsync(taskData, traceID)
}

// executeTaskAsync executes the task in background and reports status to TaskMonitor
func (h *TaskHandler) executeTaskAsync(requestData types.SendTaskDataToKeeper, traceID string) {
	h.logger.Info("Starting async task execution", "task_ids", requestData.TaskID, "trace_id", traceID)

	success, err := h.executor.ExecuteTask(context.Background(), &requestData, traceID)
	if err != nil {
		h.logger.Error("Async task execution failed", "task_ids", requestData.TaskID, "error", err, "trace_id", traceID)
		// Note: TaskMonitor reporting is already handled inside ExecuteTask
		return
	}

	h.logger.Info("Async task execution completed", "task_ids", requestData.TaskID, "success", success, "trace_id", traceID)
	// Note: TaskMonitor reporting is already handled inside ExecuteTask
}
