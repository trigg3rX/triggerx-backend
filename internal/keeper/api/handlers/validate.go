package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
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
	h.logger.Info(c.Request.Context(), "Validating task ...", observability.String("trace_id", traceID))

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
	if metrics.TasksByDefinitionIDTotal != nil {
		metrics.TasksByDefinitionIDTotal.WithLabelValues(taskDefID).Add(c.Request.Context(), 1)
	}

	// Validate job based on task definition ID
	isValid := false
	var validationErr error

	h.logger.Info(c.Request.Context(), "Validating task ...", observability.String("trace_id", traceID))
	isValid, validationErr = h.validator.ValidateTask(context.Background(), taskRequest.Data, traceID)

	if validationErr != nil {
		h.logger.Error(c.Request.Context(), "Validation error", observability.Error(validationErr), observability.String("trace_id", traceID))
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: validationErr.Error(),
		})
		return
	}

	h.logger.Info(c.Request.Context(), "Task validation completed", observability.String("trace_id", traceID))
	c.JSON(http.StatusOK, ValidationResponse{
		Data:    isValid,
		Error:   false,
		Message: "",
	})
}
