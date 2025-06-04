package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

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

	jobDataRaw := requestData["jobData"]

	// Convert to proper types
	var jobData types.HandleCreateJobData
	jobDataBytes, err := json.Marshal(jobDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job data format"})
		return
	}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse job data"})
		return
	}

	h.logger.Infof("Job ID: %d", jobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
}

func (h *Handler) ValidateTask(c *gin.Context) {
	var taskRequest types.TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	var decodedData string
	if strings.HasPrefix(taskRequest.Data, "0x") {
		dataBytes, err := hex.DecodeString(taskRequest.Data[2:])
		if err != nil {
			h.logger.Errorf("Failed to hex-decode data: %v", err)
			c.JSON(http.StatusBadRequest, types.ValidationResponse{
				Data:    false,
				Error:   true,
				Message: fmt.Sprintf("Failed to decode hex data: %v", err),
			})
			return
		}
		decodedData = string(dataBytes)
		h.logger.Infof("Decoded Data CID: %s", decodedData)
	} else {
		decodedData = taskRequest.Data
	}
	
	h.logger.Infof("Decoded Data: %s", decodedData[:20])
	c.JSON(http.StatusOK, types.ValidationResponse{
		Data:    true,
		Error:   false,
		Message: "",
	})
}