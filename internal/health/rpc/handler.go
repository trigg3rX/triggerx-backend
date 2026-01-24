package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// Handler implements RPCHandler for health service RPC operations
type Handler struct {
	logger            observability.Logger
	tracer            observability.Tracer
	stateManager      *keeper.StateManager
	performerSelector *keeper.PerformerSelector
}

// NewHandler creates a new RPC handler for health service
func NewHandler(logger observability.Logger, tracer observability.Tracer, stateManager *keeper.StateManager, performerSelector *keeper.PerformerSelector) *Handler {
	return &Handler{
		logger:            logger,
		tracer:            tracer,
		stateManager:      stateManager,
		performerSelector: performerSelector,
	}
}

// Handle processes RPC method calls
func (h *Handler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	// Create trace
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("health.%s", method),
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", "health"),
		),
	)
	defer span.End()

	span.AddEvent("rpc.request.received")

	switch method {
	case "get-performer":
		return h.handleGetPerformer(ctx, request)
	case "health":
		return h.handleHealth(ctx, request)
	default:
		span.SetStatus(codes.Error, fmt.Sprintf("unknown method: %s", method))
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleGetPerformer handles the get-performer RPC method
func (h *Handler) handleGetPerformer(ctx context.Context, request interface{}) (interface{}, error) {
	// Convert request to GetPerformerRequest
	var req types.GetPerformerRequest

	// Handle JSON request - convert map to JSON bytes then unmarshal
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get performer request: %w", err)
	}

	// Use Redis-based round-robin selection with fallback
	performer, err := h.performerSelector.GetNextPerformerWithFallback(ctx, req.Network)
	if err != nil {
		h.logger.Error(ctx, "Failed to get performer",
			observability.Error(err),
			observability.String("network", string(req.Network)),
		)
		return types.GetPerformerResponse{
			Success: false,
			Error:   fmt.Sprintf("no suitable performers available for network=%s: %v", string(req.Network), err),
		}, nil
	}

	h.logger.Debug(ctx, "Selected performer via gRPC",
		observability.Int64("operator_id", performer.OperatorID),
		observability.String("keeper_address", performer.KeeperAddress),
		observability.String("network", string(performer.Network)),
	)

	return types.GetPerformerResponse{
		Performer: *performer,
		Success:   true,
	}, nil
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// handleHealth handles the health RPC method
func (h *Handler) handleHealth(ctx context.Context, request interface{}) (interface{}, error) {
	return HealthResponse{
		Status:    "healthy",
		Version:   config.GetVersion(),
		Timestamp: time.Now().UTC(),
	}, nil
}

// GetMethods returns available RPC methods
func (h *Handler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:         "get-performer",
			Description:  "Get a performer based on isImua and isMainnet flags",
			RequestType:  types.GetPerformerRequest{},
			ResponseType: types.GetPerformerResponse{},
			Timeout:      config.GetRPCGetPerformerTimeout(),
		},
		{
			Name:         "health",
			Description:  "Health check endpoint",
			RequestType:  map[string]interface{}{},
			ResponseType: HealthResponse{},
			Timeout:      config.GetRPCHealthTimeout(),
		},
	}
}

