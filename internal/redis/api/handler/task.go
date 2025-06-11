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

	// var resultBytes []byte
	// if config.GetKeeperAddress() != performerData.KeeperAddress {
	// 	h.logger.Infof("I am not the performer: %s", performerData.KeeperAddress)
	// 	c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
	// 	return
	// } else {
	// 	switch taskDefinitionID {
	// 	case 1, 2:
	// 		var timeJobData types.ScheduleTimeJobData
	// 		timeJobDataBytes, err := json.Marshal(requestData["timeJobData"])
	// 		if err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time job data format"})
	// 			return
	// 		}
	// 		if err := json.Unmarshal(timeJobDataBytes, &timeJobData); err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse time job data"})
	// 			return
	// 		}

	// 		// TODO: Execute the task
	// 		actionData, err := h.executor.ExecuteTimeBasedTask(&timeJobData)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
	// 			return
	// 		}

	// 		// Convert result to bytes
	// 		resultBytes, err = json.Marshal(actionData)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal result"})
	// 			return
	// 		}
	// 	case 3, 4, 5, 6:
	// 		var taskTargetData types.SendTaskTargetData
	// 		var triggerData types.SendTriggerData

	// 		taskTargetDataBytes, err := json.Marshal(requestData["taskTargetData"])
	// 		if err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event job data format"})
	// 			return
	// 		}
	// 		if err := json.Unmarshal(taskTargetDataBytes, &taskTargetData); err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse task target data"})
	// 			return
	// 		}

	// 		triggerDataBytes, err := json.Marshal(requestData["triggerData"])
	// 		if err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data format"})
	// 			return
	// 		}
	// 		if err := json.Unmarshal(triggerDataBytes, &triggerData); err != nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse trigger data"})
	// 			return
	// 		}

	// 		// TODO: Execute the task
	// 		actionData, err := h.executor.ExecuteTask(&taskTargetData, &triggerData)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
	// 			return
	// 		}

	// 		// Convert result to bytes
	// 		resultBytes, err = json.Marshal(actionData)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal result"})
	// 			return
	// 		}
	// 	default:
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
	// 		return
	// 	}
	// }

	h.logger.Infof("Job ID: %d", taskDefinitionID)
	c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
}

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
		h.logger.Errorf("Failed to fetch IPFS file: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Failed to fetch IPFS file",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("IPFS file fetch returned status: %d", resp.StatusCode)
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
		h.logger.Errorf("Failed to decode IPFSData: %v", err)
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
		err := h.tsm.AddTaskToCompletedStream(taskData, map[string]interface{}{
			"action_tx_hash": ipfsData.ActionData.ActionTxHash,
		})
		if err != nil {
			h.logger.Errorf("Failed to add task to completed stream: %v", err)
		}
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    true,
			Error:   false,
			Message: "Task completed successfully",
		})
	} else {
		err := h.tsm.AddTaskToFailedStream(taskData)
		if err != nil {
			h.logger.Errorf("Failed to add task to failed stream: %v", err)
		}
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Task failed: ActionTxHash is empty",
		})
	}
}
