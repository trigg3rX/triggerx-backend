package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/core/registry"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/core/service"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler implements RPCHandler for event monitor RPC operations
type Handler struct {
	logger          observability.Logger
	tracer          observability.Tracer
	registryManager *registry.RegistryManager
	service         *service.Service
}

// NewHandler creates a new RPC handler for event monitor
func NewHandler(logger observability.Logger, tracer observability.Tracer, registryManager *registry.RegistryManager, svc *service.Service) *Handler {
	return &Handler{
		logger:          logger,
		tracer:          tracer,
		registryManager: registryManager,
		service:         svc,
	}
}

// Handle processes RPC method calls
func (h *Handler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	// Create trace
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("event-monitor.%s", method),
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", "event_monitor"),
		),
	)
	defer span.End()

	span.AddEvent("rpc.request.received")

	switch method {
	case "register":
		return h.handleRegister(ctx, request)
	case "unregister":
		return h.handleUnregister(ctx, request)
	case "health":
		return h.handleHealth(ctx, request)
	default:
		span.SetStatus(codes.Error, fmt.Sprintf("unknown method: %s", method))
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleRegister handles the register RPC method
func (h *Handler) handleRegister(ctx context.Context, request interface{}) (interface{}, error) {
	// Convert request to MonitoringRequest
	var req types.MonitoringRequest

	// Handle JSON request - convert map to JSON bytes then unmarshal
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monitoring request: %w", err)
	}

	// Validate expiration time
	if req.ExpiresAt.Before(time.Now()) {
		return map[string]interface{}{
			"success": false,
			"error":   "expires_at must be in the future",
		}, nil
	}

	// Register the request (service handles worker creation)
	if err := h.service.Register(&req); err != nil {
		h.logger.Error(ctx, "Failed to register monitoring request",
			observability.Error(err),
			observability.String("request_id", req.RequestID))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	response := types.RegisterResponse{
		Success:   true,
		RequestID: req.RequestID,
		Status:    "registered",
		Message:   "Monitoring request registered successfully",
	}

	h.logger.Debug(ctx, "Monitoring request registered",
		observability.String("request_id", req.RequestID))

	return response, nil
}

// handleUnregister handles the unregister RPC method
func (h *Handler) handleUnregister(ctx context.Context, request interface{}) (interface{}, error) {
	// Extract request ID from request
	var requestData map[string]interface{}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &requestData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	requestID, ok := requestData["request_id"].(string)
	if !ok {
		return nil, fmt.Errorf("request_id is required")
	}

	// Unregister the request (service handles worker cleanup)
	if err := h.service.Unregister(requestID); err != nil {
		h.logger.Warn(ctx, "Failed to unregister monitoring request",
			observability.Error(err),
			observability.String("request_id", requestID))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	response := types.UnregisterResponse{
		Success:   true,
		RequestID: requestID,
		Status:    "unregistered",
		Message:   "Monitoring request unregistered successfully",
	}

	h.logger.Debug(ctx, "Monitoring request unregistered",
		observability.String("request_id", requestID))

	return response, nil
}

// handleHealth handles the health RPC method
func (h *Handler) handleHealth(ctx context.Context, request interface{}) (interface{}, error) {
	response := types.HealthResponse{
		Status:          "healthy",
		Version:         config.GetVersion(),
		ActiveMonitors:  h.registryManager.GetActiveMonitorCount(),
		ChainsSupported: h.registryManager.GetChainsSupported(),
	}

	return response, nil
}

// GetMethods returns available RPC methods
func (h *Handler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:         "register",
			Description:  "Register a new monitoring request for blockchain events",
			RequestType:  types.MonitoringRequest{},
			ResponseType: types.RegisterResponse{},
			Timeout:      30 * time.Second,
		},
		{
			Name:        "unregister",
			Description: "Unregister a monitoring request",
			RequestType: map[string]interface{}{
				"request_id": "string",
			},
			ResponseType: types.UnregisterResponse{},
			Timeout:      30 * time.Second,
		},
		{
			Name:         "health",
			Description:  "Health check endpoint",
			RequestType:  map[string]interface{}{},
			ResponseType: types.HealthResponse{},
			Timeout:      5 * time.Second,
		},
	}
}
