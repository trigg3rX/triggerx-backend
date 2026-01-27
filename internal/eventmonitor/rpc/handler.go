package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	case "processTransaction":
		return h.handleProcessTransaction(ctx, request)
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

// handleProcessTransaction handles the processTransaction RPC method
func (h *Handler) handleProcessTransaction(ctx context.Context, request interface{}) (interface{}, error) {
	// Convert request to ProcessTransactionRequest
	var req types.ProcessTransactionRequest

	// Handle JSON request - convert map to JSON bytes then unmarshal
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process transaction request: %w", err)
	}

	// Validate request
	if req.TxHash == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "tx_hash is required",
		}, nil
	}
	if req.ChainID == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "chain_id is required",
		}, nil
	}

	// Process transaction via service
	eventName, err := h.service.ProcessTransaction(ctx, req.TxHash, req.ChainID, req.IsRejected)
	if err != nil {
		h.logger.Error(ctx, "Failed to process transaction",
			observability.Error(err),
			observability.String("tx_hash", req.TxHash),
			observability.String("chain_id", req.ChainID))
		return types.ProcessTransactionResponse{
			Success:   false,
			TxHash:    req.TxHash,
			EventName: "",
			Message:   err.Error(),
		}, nil
	}

	response := types.ProcessTransactionResponse{
		Success:   true,
		TxHash:    req.TxHash,
		EventName: eventName,
		Message:   fmt.Sprintf("Successfully processed transaction, event: %s", eventName),
	}

	h.logger.Info(ctx, "Transaction processed successfully",
		observability.String("tx_hash", req.TxHash),
		observability.String("chain_id", req.ChainID),
		observability.String("event_name", eventName))

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
			Name:         "processTransaction",
			Description:  "Process a transaction and extract TaskSubmitted/TaskRejected events",
			RequestType:  types.ProcessTransactionRequest{},
			ResponseType: types.ProcessTransactionResponse{},
			Timeout:      60 * time.Second,
		},
	}
}
