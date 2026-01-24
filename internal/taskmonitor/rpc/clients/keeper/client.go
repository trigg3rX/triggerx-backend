package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Client represents a client for communicating with keeper API
type Client struct {
	httpClient *http.Client
	logger     observability.Logger
	tracer     observability.Tracer
}

// NewClient creates a new keeper client
func NewClient(logger observability.Logger, tracer observability.Tracer) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
		tracer: tracer,
	}
}

// RebroadcastTaskRequest represents the request to rebroadcast a task
type RebroadcastTaskRequest struct {
	TaskID int64 `json:"task_id"`
}

// RebroadcastTaskResponse represents the response from rebroadcast request
type RebroadcastTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	TaskID  int64  `json:"task_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RebroadcastTask calls the keeper's rebroadcast endpoint
// Uses the fixed performer API URL from config
func (c *Client) RebroadcastTask(ctx context.Context, taskID int64) error {
	// Create span for rebroadcast call
	ctx, span := c.tracer.Start(ctx, "keeper.rebroadcast",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.Int64("task.id", taskID),
		),
	)
	defer span.End()

	// Use fixed performer URL from config
	performerURL := config.GetPerformerAPIUrl() + "/task/rebroadcast"

	// Create request
	reqBody := RebroadcastTaskRequest{
		TaskID: taskID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return fmt.Errorf("failed to marshal rebroadcast request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, performerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return fmt.Errorf("failed to create rebroadcast request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Inject trace context
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request failed")
		c.logger.Error(ctx, "Failed to send rebroadcast request to keeper",
			observability.String("performer_url", performerURL),
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return fmt.Errorf("failed to send rebroadcast request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn(ctx, "Failed to close response body",
				observability.Error(err))
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error   string `json:"error"`
			Details string `json:"details,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			span.SetStatus(codes.Error, errResp.Error)
			c.logger.Error(ctx, "Keeper returned error for rebroadcast",
				observability.Int64("task_id", taskID),
				observability.Int("status_code", resp.StatusCode),
				observability.String("error", errResp.Error))
			return fmt.Errorf("keeper returned error: %s", errResp.Error)
		}
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return fmt.Errorf("keeper returned non-OK status: %d", resp.StatusCode)
	}

	// Parse response
	var response RebroadcastTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return fmt.Errorf("failed to decode rebroadcast response: %w", err)
	}

	if !response.Success {
		span.SetStatus(codes.Error, response.Error)
		return fmt.Errorf("rebroadcast failed: %s", response.Error)
	}

	span.SetStatus(codes.Ok, "rebroadcast successful")
	c.logger.Info(ctx, "Task rebroadcast request sent successfully",
		observability.String("performer_url", performerURL),
		observability.Int64("task_id", taskID))

	return nil
}
