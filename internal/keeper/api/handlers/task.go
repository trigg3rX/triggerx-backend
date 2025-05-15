package handlers

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor defines the interface for task execution
type TaskExecutor interface {
	Execute(job *types.HandleCreateJobData) (types.ActionData, error)
}

// TaskValidator defines the interface for task validation
type TaskValidator interface {
	ValidateTask(task *types.TaskData) (bool, error)
}

// TaskHandler handles task-related requests
type TaskHandler struct {
	logger    logging.Logger
	executor  TaskExecutor
	validator TaskValidator
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(logger logging.Logger, executor TaskExecutor, validator TaskValidator) *TaskHandler {
	return &TaskHandler{
		logger:    logger,
		executor:  executor,
		validator: validator,
	}
}

// ExecuteTask handles task execution requests
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
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

	// Parse job data
	var requestData struct {
		JobData       *types.HandleCreateJobData `json:"jobData"`
		TriggerData   *types.TriggerData         `json:"triggerData"`
		PerformerData *types.GetPerformerData    `json:"performerData"`
	}

	if err := json.Unmarshal(decodedData, &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request data"})
		return
	}

	// Execute task
	actionData, err := h.executor.Execute(requestData.JobData)
	if err != nil {
		h.logger.Error("Failed to execute task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
		return
	}

	// Set task ID from trigger data
	actionData.TaskID = requestData.TriggerData.TaskID

	// Convert result to bytes
	resultBytes, err := json.Marshal(actionData)
	if err != nil {
		h.logger.Error("Failed to marshal result", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process result"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", resultBytes)
}

// ValidateTask handles task validation requests
func (h *TaskHandler) ValidateTask(c *gin.Context) {
	// var request types.TaskValidationRequest
	// if err := c.BindJSON(&request); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	// 	return
	// }

	// // Validate task
	// isValid, err := h.validator.ValidateTask(&request)
	// if err != nil {
	// 	h.logger.Error("Task validation failed", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"isValid": true,
	})
}
