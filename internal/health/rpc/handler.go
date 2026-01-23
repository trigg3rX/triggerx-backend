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
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// Handler implements RPCHandler for health service RPC operations
type Handler struct {
	logger       observability.Logger
	tracer       observability.Tracer
	stateManager *keeper.StateManager
}

// NewHandler creates a new RPC handler for health service
func NewHandler(logger observability.Logger, tracer observability.Tracer, stateManager *keeper.StateManager) *Handler {
	return &Handler{
		logger:       logger,
		tracer:       tracer,
		stateManager: stateManager,
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

	// For now, use hardcoded values (same as HTTP endpoint)
	// TODO: Implement Redis-based round-robin selection when ready
	if string(req.Network) == string(types.NetworkMainnet) {
		return types.GetPerformerResponse{
			Performer: types.PerformerData{
				OperatorID:    1002,
				KeeperAddress: "0x235813b36eea7e48b7069821a78c0bc8384a3c79",
				Network:       types.NetworkMainnet,
			},
			Success: true,
		}, nil
	}

	// Fallback performers for testnet
	fallbackPerformers := []types.PerformerData{
		{
			OperatorID:    2,
			KeeperAddress: "0x0a067a261c5F5e8C4c0b9137430b4FE1255EB62e",
			Network:       types.NetworkSepolia,
		},
		{
			OperatorID:    1,
			KeeperAddress: "0xcacce39134e3b9d5d9220d87fc546c6f0fb9cc37",
			Network:       types.NetworkImua,
		},
	}

	// Filter by Imua status
	var filteredPerformer types.PerformerData
	for _, performer := range fallbackPerformers {
		if string(performer.Network) == string(req.Network) {
			filteredPerformer = performer
			break
		}
	}

	if filteredPerformer == (types.PerformerData{}) {
		return types.GetPerformerResponse{
			Success: false,
			Error:   fmt.Sprintf("no suitable performers available for network=%s", string(req.Network)),
		}, nil
	}

	h.logger.Debug(ctx, "Selected performer via gRPC",
		observability.Int64("operator_id", filteredPerformer.OperatorID),
		observability.String("keeper_address", filteredPerformer.KeeperAddress),
		observability.String("network", string(filteredPerformer.Network)),
	)

	return types.GetPerformerResponse{
		Performer: filteredPerformer,
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

