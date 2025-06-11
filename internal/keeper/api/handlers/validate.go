package handlers

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
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

// ValidateTask handles task validation requests
func (h *TaskHandler) ValidateTask(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info("Validating task ...", "trace_id", traceID)

	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	// Track task by definition ID for validation
	taskDefID := strconv.Itoa(int(taskRequest.TaskDefinitionID))
	metrics.TasksByDefinitionIDTotal.WithLabelValues(taskDefID).Inc()

	// Decode the data if it's hex-encoded (with 0x prefix)
	var decodedData string
	dataBytes, err := hex.DecodeString(taskRequest.Data[2:]) // Remove "0x" prefix before decoding
	if err != nil {
		h.logger.Errorf("Failed to hex-decode data: %v", err)
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to decode hex data: %v", err),
		})
		return
	}
	decodedData = string(dataBytes)
	h.logger.Infof("Decoded Data CID: %s", decodedData)

	ipfsData, err := h.ipfsFetcher.FetchContent(decodedData)
	if err != nil {
		h.logger.Errorf("Failed to fetch IPFS content: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content: %v", err),
		})
		return
	}

	// Validate job based on task definition ID
	isValid := false
	var validationErr error

	h.logger.Info("Validating task ...", "trace_id", traceID)
	isValid, validationErr = h.validator.ValidateTask(ipfsData.TaskData, traceID)

	if validationErr != nil {
		h.logger.Error("Validation error", "error", validationErr, "trace_id", traceID)
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: validationErr.Error(),
		})
		return
	}

	h.logger.Info("Task validation completed", "trace_id", traceID)
	c.JSON(http.StatusOK, ValidationResponse{
		Data:    isValid,
		Error:   false,
		Message: "",
	})
}
