package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

// Listen for Custom P2P Messages, and
func (h *handler) HandleP2PMessage(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[HandleP2PMessage] Handling P2P message", "trace_id", traceID)
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

	h.logger.Info("[HandleP2PMessage] Task data Broadcast successfull", "trace_id", traceID, "task_id", requestData.TaskID)
	c.JSON(http.StatusOK, gin.H{"success": "P2P message received"})
}

// ValidateTask updates the appropriate stream after performer broadcast
func (h *handler) HandleValidateRequest(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[HandleValidateRequest] Validating task", "trace_id", traceID)
	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	// Decode the data if it's hex-encoded (with 0x prefix)
	var decodedData string
	dataBytes, err := hex.DecodeString(taskRequest.Data[2:]) // Remove "0x" prefix before decoding
	if err != nil {
		h.logger.Error("[HandleValidateRequest] Failed to hex-decode data", "trace_id", traceID, "error", err)
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to decode hex data: %v", err),
		})
		return
	}
	decodedData = string(dataBytes)
	h.logger.Info("[HandleValidateRequest] Decoded Data CID", "trace_id", traceID, "cid", decodedData)

	ipfsData, err := ipfs.FetchIPFSContent(config.GetPinataHost(), decodedData)
	if err != nil {
		h.logger.Error("[HandleValidateRequest] Failed to fetch IPFS content", "trace_id", traceID, "error", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content: %v", err),
		})
		return
	}

	h.logger.Info("[HandleValidateRequest] Updating task stream and database ...", "trace_id", traceID)

	h.taskStreamMgr.UpdateDatabase(ipfsData)

	h.logger.Info("[HandleValidateRequest] Task validation completed", "trace_id", traceID)
	c.JSON(http.StatusOK, ValidationResponse{
		Data:    true,
		Error:   false,
		Message: "",
	})
}
