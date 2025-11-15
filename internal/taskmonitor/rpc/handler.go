package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// TaskMonitorHandler implements the generic RPC handler interface
type TaskMonitorHandler struct {
	logger  logging.Logger
	monitor TaskMonitorInterface
}

// TaskMonitorInterface defines the interface for task monitor operations
type TaskMonitorInterface interface {
	ReportTaskError(ctx context.Context, req *types.ReportTaskErrorRequest) (*types.ReportTaskErrorResponse, error)
}

// NewTaskMonitorHandler creates a new RPC handler
func NewTaskMonitorHandler(logger logging.Logger, monitor TaskMonitorInterface) *TaskMonitorHandler {
	return &TaskMonitorHandler{
		logger:  logger,
		monitor: monitor,
	}
}

// Handle routes incoming RPC requests based on the method name
func (h *TaskMonitorHandler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	switch method {
	case "report-task-error":
		// Convert request to the expected type
		req, ok := request.(*types.ReportTaskErrorRequest)
		if !ok {
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				var err error
				req, err = h.convertMapToRequest(reqMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid request type for report-task-error: %T", request)
			}
		}

		// Validate keeper signature
		if err := h.validateSignature(req); err != nil {
			h.logger.Error("Invalid keeper signature for task error report",
				"task_id", req.TaskID,
				"keeper_address", req.KeeperAddress,
				"error", err)
			return nil, fmt.Errorf("invalid signature: %w", err)
		}

		resp, err := h.monitor.ReportTaskError(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// GetMethods advertises the supported RPC methods for discovery/metrics
func (h *TaskMonitorHandler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:         "report-task-error",
			Description:  "Report a task execution error from a keeper",
			RequestType:  &types.ReportTaskErrorRequest{},
			ResponseType: &types.ReportTaskErrorResponse{},
			Timeout:      30 * time.Second,
		},
	}
}

// convertMapToRequest converts a map to ReportTaskErrorRequest
// This is used when the request comes as JSON-decoded map
func (h *TaskMonitorHandler) convertMapToRequest(reqMap map[string]interface{}) (*types.ReportTaskErrorRequest, error) {
	// Convert map back to JSON and then unmarshal to proper struct
	jsonData, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request map: %w", err)
	}

	var req types.ReportTaskErrorRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &req, nil
}

// validateSignature validates the keeper's signature for the error report
func (h *TaskMonitorHandler) validateSignature(req *types.ReportTaskErrorRequest) error {
	// Create a struct for signing (without signature field)
	signData := struct {
		TaskID        int64  `json:"task_id"`
		KeeperAddress string `json:"keeper_address"`
		Error         string `json:"error"`
	}{
		TaskID:        req.TaskID,
		KeeperAddress: req.KeeperAddress,
		Error:         req.Error,
	}

	// Verify signature using JSON verification (same as other services)
	isValid, err := cryptography.VerifySignatureFromJSON(signData, req.Signature, req.KeeperAddress)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	if !isValid {
		return fmt.Errorf("invalid signature for keeper %s", req.KeeperAddress)
	}

	return nil
}
