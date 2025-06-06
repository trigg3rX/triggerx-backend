package handlers

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ExecuteTask handles task execution requests
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
	h.logger.Infof("Executing task ...")

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

	// Convert to proper types
	var performerData types.GetPerformerData
	performerDataBytes, err := json.Marshal(requestData.PerformerData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performer data format"})
		return
	}
	if err := json.Unmarshal(performerDataBytes, &performerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse performer data"})
		return
	}

	if config.GetKeeperAddress() != performerData.KeeperAddress {
		h.logger.Infof("I am not the performer: %s", performerData.KeeperAddress)
		c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
		return
	} else {
		var targetData types.SendTaskTargetDataToKeeper
		targetDataBytes, err := json.Marshal(requestData.TargetData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time job data format"})
			return
		}
		if err := json.Unmarshal(targetDataBytes, &targetData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse time job data"})
			return
		}

		var triggerData types.SendTaskTriggerDataToKeeper
		triggerDataBytes, err := json.Marshal(requestData.TriggerData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data format"})
			return
		}
		if err := json.Unmarshal(triggerDataBytes, &triggerData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse trigger data"})
			return
		}

		success, err := h.executor.ExecuteTask(&targetData, &triggerData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": strconv.FormatBool(success)})
	}
}
