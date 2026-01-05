package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ExecuteTask handles task execution requests with async processing.
// Returns 202 Accepted immediately after validation, then executes task in background.
// Task completion status is reported to TaskMonitor (not back to caller).
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
	// Extract trace context from HTTP headers
	ctx := c.Request.Context()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

	traceID := h.getTraceID(c)

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

	// Check if this performer should handle the task
	if !strings.EqualFold(config.GetKeeperAddress(), requestData.PerformerData.KeeperAddress) {
		c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
		return
	}

	// Log task info
	taskIDs := requestData.TaskID
	h.logger.Info(ctx, "Task execution started", observability.Any("task_ids", taskIDs), observability.String("trace_id", traceID))
	for _, task := range requestData.TargetData {
		h.logger.Debug(ctx, fmt.Sprintf("Task ID: %d | Target Chain ID: %s", task.TaskID, task.TargetChainID))
	}

	// Return 202 Accepted immediately - task will be processed asynchronously
	// TaskMonitor will receive the completion status via reportTaskStatus
	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"task_ids": taskIDs,
		"trace_id": traceID,
		"message":  "Task accepted for processing. Status will be reported to TaskMonitor.",
	})

	// Create a detached context that preserves trace context but isn't tied to HTTP request lifecycle
	// This prevents the context from being canceled when the HTTP handler returns
	asyncCtx := h.createDetachedContext(ctx)

	// Execute task asynchronously in a goroutine
	// Make a copy of requestData to avoid race conditions
	taskData := requestData
	go h.executeTaskAsync(asyncCtx, taskData, traceID)
}

// createDetachedContext creates a new context that preserves the trace context
// but is not tied to the HTTP request lifecycle. This allows async operations
// to continue even after the HTTP handler returns.
func (h *TaskHandler) createDetachedContext(ctx context.Context) context.Context {
	// Extract span context from the current context
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	// Create a new background context (not tied to HTTP request)
	detachedCtx := context.Background()

	// Inject the span context into the new context to preserve tracing
	if spanContext.HasTraceID() {
		detachedCtx = trace.ContextWithSpanContext(detachedCtx, spanContext)
	}

	return detachedCtx
}

// executeTaskAsync executes the task in background and reports status to TaskMonitor
func (h *TaskHandler) executeTaskAsync(ctx context.Context, requestData types.SendTaskDataToKeeper, traceID string) {
	success, err := h.executor.ExecuteTask(ctx, &requestData, traceID)
	if err != nil {
		h.logger.Error(ctx, "Task execution failed", observability.Any("task_ids", requestData.TaskID), observability.Error(err), observability.String("trace_id", traceID))
		// Note: TaskMonitor reporting is already handled inside ExecuteTask
		return
	}

	h.logger.Info(ctx, "Task execution completed", observability.Any("task_ids", requestData.TaskID), observability.Bool("success", success), observability.String("trace_id", traceID))
}
