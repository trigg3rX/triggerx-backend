package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/redis"

	// "github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
	TaskID           int64  `json:"task_id"`
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

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(decodedDataString), &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse JSON data",
		})
		return
	}

	taskDefinitionID := requestData["taskDefinitionId"]
	performerDataRaw := requestData["performerData"]

	// Convert to proper types
	var performerData types.PerformerData
	performerDataBytes, err := json.Marshal(performerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performer data format"})
		return
	}
	if err := json.Unmarshal(performerDataBytes, &performerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse performer data"})
		return
	}

	// Create job data and add to job stream
	jobData := &redis.JobStreamData{
		JobID:     int64(taskDefinitionID.(float64)), // Convert from interface{}
		ManagerID: 1,                                 // Default manager ID
		TaskIDs:   []int64{},                         // Will be populated based on actual implementation
	}

	// Add job to running stream
	if err := h.jobStreamMgr.AddJobToRunningStream(jobData); err != nil {
		h.logger.Error("Failed to add job to running stream", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process job"})
		return
	}

	h.logger.Info("Job processed", "job_id", taskDefinitionID)
	c.JSON(http.StatusOK, gin.H{"message": "Job processed successfully"})
}

// ValidateTask validates a task and updates the appropriate stream
func (h *Handler) ValidateTask(c *gin.Context) {
	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	// Fetch the IPFS file from the URL in Data
	ipfsURL := taskRequest.Data
	resp, err := http.Get(ipfsURL)
	if err != nil {
		h.logger.Error("Failed to fetch IPFS file", "error", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Failed to fetch IPFS file",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("IPFS file fetch returned status", "status", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "IPFS file not found",
		})
		return
	}

	// Parse the file as types.IPFSData (nested JSON)
	var ipfsData types.IPFSData
	if err := json.NewDecoder(resp.Body).Decode(&ipfsData); err != nil {
		h.logger.Error("Failed to decode IPFSData", "error", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Failed to decode IPFSData",
		})
		return
	}

	// Reconstruct taskData from request
	taskData := &redis.TaskStreamData{
		TaskID: taskRequest.TaskID,
	}

	if ipfsData.ActionData.ActionTxHash != "" {
		err := h.taskStreamMgr.AddTaskToCompletedStream(taskData, map[string]interface{}{
			"action_tx_hash": ipfsData.ActionData.ActionTxHash,
		})
		if err != nil {
			h.logger.Error("Failed to add task to completed stream", "error", err)
		}
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    true,
			Error:   false,
			Message: "Task completed successfully",
		})
	} else {
		err := h.taskStreamMgr.AddTaskToFailedStream(taskData)
		if err != nil {
			h.logger.Error("Failed to add task to failed stream", "error", err)
		}
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Task failed: ActionTxHash is empty",
		})
	}
}
