package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidationRequest struct {
	Data string `json:"data"`
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

	// Unmarshal the JSON data into IPFSData struct
	var ipfsData types.IPFSData
	if err := json.Unmarshal([]byte(taskRequest.Data), &ipfsData); err != nil {
		h.logger.Errorf("Failed to unmarshal IPFS data: %v", err)
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to unmarshal IPFS data: %v", err),
		})
		return
	}

	// Validate that we have the required task data
	if ipfsData.TaskData == nil {
		h.logger.Error("IPFS data missing task_data")
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "IPFS data missing task_data",
		})
		return
	}

	// Validate job based on task definition ID
	isValid := false
	var validationErr error

	// h.logger.Info("Validating task ...", "trace_id", traceID)
	isValid, validationErr = h.validator.ValidateTask(context.Background(), ipfsData, traceID)

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
