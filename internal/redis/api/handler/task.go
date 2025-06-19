package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/redis/redis"
)

type TaskValidationRequest struct {
	TaskID       int64  `json:"task_id" binding:"required"`
	JobID        int64  `json:"job_id" binding:"required"`
	ActionsHash  string `json:"actions_hash" binding:"required"`
	Data         string `json:"data" binding:"required"` // hex encoded
	PerformerID  int64  `json:"performer_id" binding:"required"`
	ChainID      int64  `json:"chain_id" binding:"required"`
	JobManagerID int64  `json:"job_manager_id" binding:"required"`
}

type P2PMessageRequest struct {
	TaskID       int64  `json:"task_id" binding:"required"`
	JobID        int64  `json:"job_id" binding:"required"`
	ActionsHash  string `json:"actions_hash" binding:"required"`
	Data         string `json:"data" binding:"required"` // hex encoded
	PerformerID  int64  `json:"performer_id" binding:"required"`
	ChainID      int64  `json:"chain_id" binding:"required"`
	JobManagerID int64  `json:"job_manager_id" binding:"required"`
}

type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handler encapsulates the dependencies for health handlers
// Add tsm (TaskStreamManager) to Handler

// NewHandler creates a new instance of Handler

// HandleP2PMessage handles peer-to-peer messages (following keeper pattern)
func (h *Handler) HandleP2PMessage(c *gin.Context) {
	start := time.Now()
	h.logger.Info("Received P2P message",
		"remote_addr", c.Request.RemoteAddr,
		"user_agent", c.Request.UserAgent(),
		"content_length", c.Request.ContentLength)

	var req P2PMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind P2P message request",
			"error", err,
			"request_body_size", c.Request.ContentLength,
			"remote_addr", c.Request.RemoteAddr)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	h.logger.Info("P2P message request validated",
		"task_id", req.TaskID,
		"job_id", req.JobID,
		"performer_id", req.PerformerID,
		"chain_id", req.ChainID,
		"job_manager_id", req.JobManagerID,
		"actions_hash", req.ActionsHash,
		"data_length", len(req.Data))

	// Decode hex data
	data, err := hex.DecodeString(req.Data)
	if err != nil {
		h.logger.Error("Failed to decode hex data in P2P message",
			"task_id", req.TaskID,
			"error", err,
			"data_prefix", req.Data[:min(20, len(req.Data))])

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hex data", "details": err.Error()})
		return
	}

	h.logger.Debug("Hex data decoded successfully",
		"task_id", req.TaskID,
		"original_length", len(req.Data),
		"decoded_length", len(data))

	// Create job stream data using the correct structure
	jobData := &redis.JobStreamData{
		JobID:            req.JobID,
		ManagerID:        req.JobManagerID,
		TaskIDs:          []int64{req.TaskID},
		TaskDefinitionID: 1, // Default value
		SuccessCount:     0,
		FailureCount:     0,
	}

	h.logger.Info("Adding job data to stream",
		"task_id", req.TaskID,
		"job_id", req.JobID,
		"performer_id", req.PerformerID,
		"data_size", len(data),
		"received_at", time.Now().Format(time.RFC3339))

	err = h.jobStreamMgr.AddJobToRunningStream(jobData)
	if err != nil {
		h.logger.Error("Failed to add job data to stream",
			"task_id", req.TaskID,
			"job_id", req.JobID,
			"error", err,
			"stream_operation_duration", time.Since(start))

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message", "details": err.Error()})
		return
	}

	processingDuration := time.Since(start)

	h.logger.Info("P2P message processed successfully",
		"task_id", req.TaskID,
		"job_id", req.JobID,
		"processing_duration", processingDuration,
		"data_size", len(data))

	c.JSON(http.StatusOK, gin.H{
		"message":      "P2P message processed successfully",
		"task_id":      req.TaskID,
		"job_id":       req.JobID,
		"processed_at": time.Now().Format(time.RFC3339),
		"data_size":    len(data),
	})
}

// ValidateTask validates a task and updates the appropriate stream
func (h *Handler) ValidateTask(c *gin.Context) {
	start := time.Now()
	h.logger.Info("Starting task validation",
		"remote_addr", c.Request.RemoteAddr,
		"user_agent", c.Request.UserAgent())

	var req TaskValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind task validation request",
			"error", err,
			"request_body_size", c.Request.ContentLength,
			"remote_addr", c.Request.RemoteAddr)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	h.logger.Info("Task validation request received",
		"task_id", req.TaskID,
		"job_id", req.JobID,
		"performer_id", req.PerformerID,
		"chain_id", req.ChainID,
		"job_manager_id", req.JobManagerID,
		"actions_hash", req.ActionsHash,
		"data_length", len(req.Data))

	// Decode hex data
	data, err := hex.DecodeString(req.Data)
	if err != nil {
		h.logger.Error("Failed to decode hex data in task validation",
			"task_id", req.TaskID,
			"error", err,
			"data_prefix", req.Data[:min(20, len(req.Data))])

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hex data", "details": err.Error()})
		return
	}

	h.logger.Debug("Hex data decoded for validation",
		"task_id", req.TaskID,
		"original_length", len(req.Data),
		"decoded_length", len(data))

	// Simple validation based on data structure
	h.logger.Info("Starting data validation",
		"task_id", req.TaskID,
		"actions_hash", req.ActionsHash)

	validationStart := time.Now()
	isValid, validationDetails := h.validateTaskData(req.TaskID, req.ActionsHash, data)
	validationDuration := time.Since(validationStart)

	totalDuration := time.Since(start)

	if isValid {
		h.logger.Info("Task validation successful",
			"task_id", req.TaskID,
			"job_id", req.JobID,
			"performer_id", req.PerformerID,
			"total_duration", totalDuration,
			"validation_duration", validationDuration,
			"validation_details", validationDetails)

		// Move task to completed stream
		taskData := &redis.TaskStreamData{
			TaskID:      req.TaskID,
			JobID:       req.JobID,
			ManagerID:   req.JobManagerID,
			PerformerID: req.PerformerID,
		}

		if err := h.taskStreamMgr.AddTaskToCompletedStream(taskData); err != nil {
			h.logger.Error("Failed to move task to completed stream",
				"task_id", req.TaskID,
				"error", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"valid":               true,
			"task_id":             req.TaskID,
			"job_id":              req.JobID,
			"validated_at":        time.Now().Format(time.RFC3339),
			"validation_duration": totalDuration.String(),
			"details":             validationDetails,
		})
	} else {
		h.logger.Warn("Task validation failed - invalid data",
			"task_id", req.TaskID,
			"job_id", req.JobID,
			"performer_id", req.PerformerID,
			"total_duration", totalDuration,
			"validation_duration", validationDuration,
			"validation_details", validationDetails)

		// Move task to failed stream
		taskData := &redis.TaskStreamData{
			TaskID:      req.TaskID,
			JobID:       req.JobID,
			ManagerID:   req.JobManagerID,
			PerformerID: req.PerformerID,
		}

		if err := h.taskStreamMgr.AddTaskToRetryStream(taskData, "validation_failed"); err != nil {
			h.logger.Error("Failed to move task to retry stream",
				"task_id", req.TaskID,
				"error", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"valid":               false,
			"task_id":             req.TaskID,
			"job_id":              req.JobID,
			"validated_at":        time.Now().Format(time.RFC3339),
			"validation_duration": totalDuration.String(),
			"details":             validationDetails,
		})
	}
}

func (h *Handler) validateTaskData(taskID int64, actionsHash string, data []byte) (bool, map[string]interface{}) {
	h.logger.Info("Validating task data",
		"task_id", taskID,
		"actions_hash", actionsHash,
		"data_size", len(data))

	validationDetails := map[string]interface{}{
		"actions_hash":         actionsHash,
		"data_size":            len(data),
		"validation_timestamp": time.Now().Format(time.RFC3339),
		"checks_performed":     []string{},
		"validation_result":    false,
	}

	checks := []string{}

	// Basic validation checks
	if len(data) == 0 {
		h.logger.Warn("Task data is empty", "task_id", taskID)
		validationDetails["error"] = "empty_data"
		return false, validationDetails
	}
	checks = append(checks, "data_not_empty")

	if actionsHash == "" {
		h.logger.Warn("Actions hash is empty", "task_id", taskID)
		validationDetails["error"] = "empty_actions_hash"
		return false, validationDetails
	}
	checks = append(checks, "actions_hash_not_empty")

	// Try to parse as JSON
	var executionData map[string]interface{}
	if err := json.Unmarshal(data, &executionData); err != nil {
		h.logger.Warn("Failed to parse execution data as JSON",
			"task_id", taskID,
			"error", err)
		validationDetails["error"] = "invalid_json"
		return false, validationDetails
	}
	checks = append(checks, "valid_json")

	// Check for basic required fields
	requiredFields := []string{"timestamp", "task_id"}
	for _, field := range requiredFields {
		if _, exists := executionData[field]; !exists {
			h.logger.Warn("Missing required field in execution data",
				"task_id", taskID,
				"missing_field", field)
			validationDetails["error"] = fmt.Sprintf("missing_field_%s", field)
			return false, validationDetails
		}
	}
	checks = append(checks, "required_fields_present")

	validationDetails["checks_performed"] = checks
	validationDetails["validation_result"] = true
	validationDetails["execution_data_keys"] = getMapKeys(executionData)

	h.logger.Info("Task data validation completed successfully",
		"task_id", taskID,
		"checks_performed", len(checks))

	return true, validationDetails
}

// Helper function to get map keys
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
