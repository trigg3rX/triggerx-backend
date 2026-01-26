package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler implements RPCHandler for condition scheduler RPC operations
type Handler struct {
	logger    observability.Logger
	tracer    observability.Tracer
	scheduler *scheduler.ConditionBasedScheduler
}

// NewHandler creates a new RPC handler for condition scheduler
func NewHandler(logger observability.Logger, tracer observability.Tracer, sched *scheduler.ConditionBasedScheduler) *Handler {
	return &Handler{
		logger:    logger,
		tracer:    tracer,
		scheduler: sched,
	}
}

// Handle processes RPC method calls
func (h *Handler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	// Create trace with format "condition-{scheduler_id}-{timestamp}"
	schedulerID := h.scheduler.GetSchedulerID()
	traceName := fmt.Sprintf("condition-%s-%d", schedulerID, time.Now().Unix())

	// Create root span for RPC operation
	ctx, span := h.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", "condition_scheduler"),
			attribute.String("scheduler.id", schedulerID),
			attribute.String("trace.name", traceName),
		),
	)
	defer span.End()

	span.AddEvent("rpc.request.received")

	switch method {
	case "schedule-job":
		return h.handleScheduleJob(ctx, request)
	case "unschedule-job":
		return h.handleUnscheduleJob(ctx, request)
	case "event-notification":
		return h.handleEventNotification(ctx, request)
	default:
		span.SetStatus(codes.Error, fmt.Sprintf("unknown method: %s", method))
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleScheduleJob handles the schedule-job RPC method
func (h *Handler) handleScheduleJob(ctx context.Context, request interface{}) (interface{}, error) {
	// Convert request to ScheduleConditionJobRequest
	var scheduleRequest types.ScheduleConditionJobRequest

	// Handle JSON request - convert map to JSON bytes then unmarshal
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &scheduleRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job request: %w", err)
	}

	// Schedule the job
	if err := h.scheduler.ScheduleJob(ctx, &scheduleRequest); err != nil {
		h.logger.Error(ctx, "Failed to schedule job via RPC",
			observability.String("job_id", scheduleRequest.JobID),
			observability.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	h.logger.Info(ctx, "Job scheduled successfully via RPC",
		observability.String("job_id", scheduleRequest.JobID))

	return map[string]interface{}{
		"success": true,
		"job_id":  scheduleRequest.JobID,
		"message": "Job scheduled successfully",
	}, nil
}

// handleUnscheduleJob handles the unschedule-job RPC method
func (h *Handler) handleUnscheduleJob(ctx context.Context, request interface{}) (interface{}, error) {
	// Extract job ID from request
	var jobID string

	// Handle JSON request - extract job_id
	var requestData map[string]interface{}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &requestData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Extract job_id
	jobIDVal, ok := requestData["job_id"]
	if !ok {
		return nil, fmt.Errorf("job_id is required")
	}

	// Handle different job_id types - convert to string
	switch v := jobIDVal.(type) {
	case string:
		jobID = v
	case float64:
		jobID = fmt.Sprintf("%.0f", v)
	case int:
		jobID = fmt.Sprintf("%d", v)
	case int64:
		jobID = fmt.Sprintf("%d", v)
	default:
		return nil, fmt.Errorf("invalid job_id type: expected string or number, got %T", v)
	}

	// Unschedule the job
	if err := h.scheduler.UnscheduleJob(ctx, jobID); err != nil {
		h.logger.Error(ctx, "Failed to unschedule job via RPC",
			observability.String("job_id", jobID),
			observability.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	h.logger.Info(ctx, "Job unscheduled successfully via RPC",
		observability.String("job_id", jobID))

	return map[string]interface{}{
		"success": true,
		"job_id":  jobID,
		"message": "Job unscheduled successfully",
	}, nil
}

// handleEventNotification handles event notifications from eventmonitor
func (h *Handler) handleEventNotification(ctx context.Context, request interface{}) (interface{}, error) {
	// Import eventmonitor types for EventNotification
	type EventNotification struct {
		RequestID    string    `json:"request_id"`
		ChainID      string    `json:"chain_id"`
		ContractAddr string    `json:"contract_address"`
		EventSig     string    `json:"event_signature"`
		BlockNumber  uint64    `json:"block_number"`
		TxHash       string    `json:"tx_hash"`
		LogIndex     uint      `json:"log_index"`
		Topics       []string  `json:"topics"`
		Data         string    `json:"data"`
		Timestamp    time.Time `json:"timestamp"`
	}

	// Convert request to EventNotification
	var eventNotif EventNotification
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &eventNotif); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event notification: %w", err)
	}

	// Convert EventNotification to TriggerNotification
	// RequestID is already a string, use it directly as JobID
	jobID := eventNotif.RequestID

	triggerNotif := &worker.TriggerNotification{
		JobID:         jobID,
		TriggerTxHash: eventNotif.TxHash,
		TriggerValue:  0, // Event notifications don't have a numeric value
		TriggeredAt:   eventNotif.Timestamp,
	}

	// Handle the notification
	if err := h.scheduler.HandleTriggerNotification(ctx, triggerNotif); err != nil {
		h.logger.Error(ctx, "Failed to handle event notification",
			observability.String("request_id", eventNotif.RequestID),
			observability.String("tx_hash", eventNotif.TxHash),
			observability.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	h.logger.Info(ctx, "Event notification handled successfully",
		observability.String("request_id", eventNotif.RequestID),
		observability.String("tx_hash", eventNotif.TxHash))

	return map[string]interface{}{
		"success": true,
		"message": "Event notification processed successfully",
	}, nil
}

// GetMethods returns available RPC methods
func (h *Handler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:        "schedule-job",
			Description: "Schedule a new condition-based or event-based job",
			RequestType: types.ScheduleConditionJobRequest{},
			ResponseType: map[string]interface{}{
				"success": true,
				"job_id":  "string",
				"message": "string",
			},
			Timeout: 30 * time.Second,
		},
		{
			Name:        "unschedule-job",
			Description: "Unschedule a condition-based or event-based job",
			RequestType: map[string]interface{}{
				"job_id": "string",
			},
			ResponseType: map[string]interface{}{
				"success": true,
				"job_id":  "string",
				"message": "string",
			},
			Timeout: 30 * time.Second,
		},
		{
			Name:        "event-notification",
			Description: "Receive event notification from eventmonitor",
			RequestType: map[string]interface{}{
				"request_id":       "string",
				"chain_id":         "string",
				"contract_address": "string",
				"event_signature":  "string",
				"block_number":     "uint64",
				"tx_hash":          "string",
				"log_index":        "uint",
				"topics":           "[]string",
				"data":             "string",
				"timestamp":        "time.Time",
			},
			ResponseType: map[string]interface{}{
				"success": true,
				"message": "string",
			},
			Timeout: 30 * time.Second,
		},
	}
}
