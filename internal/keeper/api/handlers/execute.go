package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

// ExecuteTask handles task execution requests
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info("Executing task ...", "trace_id", traceID)

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

	if config.GetKeeperAddress() != requestData.PerformerData.KeeperAddress {
		h.logger.Infof("I am not the performer: %s", requestData.PerformerData.KeeperAddress)
		c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
		return
	} else {
		h.logger.Infof("I am the performer: %s", requestData.PerformerData.KeeperAddress)

		TaskDefinitionID := requestData.TriggerData[0].TaskDefinitionID
		switch TaskDefinitionID {
		case 1, 2:
			h.logger.Info("Execution starts for following tasks:", "trace_id", traceID)
			for _, task := range requestData.TargetData {
				h.logger.Infof("Task ID: %d | Target Chain ID: %s", task.TaskID, task.TargetChainID)
			}
		case 3, 4, 5, 6:
			h.logger.Info("Execution starts for task:", "task_id", requestData.TargetData[0].TaskID, "target_chain_id", requestData.TargetData[0].TargetChainID, "trace_id", traceID)
		}
		success, err := h.executor.ExecuteTask(context.Background(), &requestData, traceID)
		if err != nil {
			h.logger.Error("Task execution failed", "error", err, "trace_id", traceID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
			return
		}

		h.logger.Info("Task execution completed", "success", success, "trace_id", traceID)
		c.JSON(http.StatusOK, gin.H{"success": strconv.FormatBool(success)})
	}
}
