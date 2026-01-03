package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// TaskMonitorHandler implements the generic RPC handler interface
type TaskMonitorHandler struct {
	logger  observability.Logger
	monitor TaskMonitorInterface
}

// TaskMonitorInterface defines the interface for task monitor operations
type TaskMonitorInterface interface {
	// ReportTaskStatus handles task status reports from keepers (both success and failure)
	ReportTaskStatus(ctx context.Context, req *types.ReportTaskStatusRequest) (*types.ReportTaskStatusResponse, error)
	// ReportConsensusEvent handles consensus event reports from eventmonitor (TaskSubmitted or TaskRejected)
	ReportConsensusEvent(ctx context.Context, req *types.ReportConsensusEventRequest) (*types.ReportConsensusEventResponse, error)
}

// NewTaskMonitorHandler creates a new RPC handler
func NewTaskMonitorHandler(logger observability.Logger, monitor TaskMonitorInterface) *TaskMonitorHandler {
	return &TaskMonitorHandler{
		logger:  logger,
		monitor: monitor,
	}
}

// Handle routes incoming RPC requests based on the method name
func (h *TaskMonitorHandler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	switch method {

	case "report-task-status":
		// Convert request to the expected type
		req, ok := request.(*types.ReportTaskStatusRequest)
		if !ok {
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				var err error
				req, err = h.convertMapToStatusRequest(reqMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid request type for report-task-status: %T", request)
			}
		}

		// Validate keeper signature
		if err := h.validateStatusSignature(req); err != nil {
			h.logger.Error(ctx, "Invalid keeper signature for task status report",
				observability.Int64("task_id", req.TaskID),
				observability.String("keeper_address", req.KeeperAddress),
				observability.Error(err))
			return nil, fmt.Errorf("invalid signature: %w", err)
		}

		resp, err := h.monitor.ReportTaskStatus(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	case "report-consensus-event":
		// Convert request to the expected type
		consensusReq, ok := request.(*types.ReportConsensusEventRequest)
		if !ok {
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				var err error
				consensusReq, err = h.convertMapToConsensusRequest(reqMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid request type for report-consensus-event: %T", request)
			}
		}

		resp, err := h.monitor.ReportConsensusEvent(ctx, consensusReq)
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
			Name:         "report-task-status",
			Description:  "Report task execution status from a keeper (success or failure)",
			RequestType:  &types.ReportTaskStatusRequest{},
			ResponseType: &types.ReportTaskStatusResponse{},
			Timeout:      30 * time.Second,
		},
		{
			Name:         "report-consensus-event",
			Description:  "Report consensus event from eventmonitor (TaskSubmitted or TaskRejected)",
			RequestType:  &types.ReportConsensusEventRequest{},
			ResponseType: &types.ReportConsensusEventResponse{},
			Timeout:      30 * time.Second,
		},
	}
}

// convertMapToStatusRequest converts a map to ReportTaskStatusRequest
// This is used when the request comes as JSON-decoded map
func (h *TaskMonitorHandler) convertMapToStatusRequest(reqMap map[string]interface{}) (*types.ReportTaskStatusRequest, error) {
	jsonData, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request map: %w", err)
	}

	var req types.ReportTaskStatusRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &req, nil
}

// validateStatusSignature validates the keeper's signature for the status report
func (h *TaskMonitorHandler) validateStatusSignature(req *types.ReportTaskStatusRequest) error {
	// Create a struct for signing (without signature field)
	// Must match exactly what the keeper signs
	signData := struct {
		TaskID              int64  `json:"task_id"`
		KeeperAddress       string `json:"keeper_address"`
		ExecutionSuccessful bool   `json:"execution_successful"`
		AggregatorSubmitted bool   `json:"aggregator_submitted"`
		ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`
		ProofCID            string `json:"proof_cid,omitempty"`
		Error               string `json:"error,omitempty"`
	}{
		TaskID:              req.TaskID,
		KeeperAddress:       req.KeeperAddress,
		ExecutionSuccessful: req.ExecutionSuccessful,
		AggregatorSubmitted: req.AggregatorSubmitted,
		ExecutionTxHash:     req.ExecutionTxHash,
		ProofCID:            req.ProofCID,
		Error:               req.Error,
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

// convertMapToConsensusRequest converts a map to ReportConsensusEventRequest
func (h *TaskMonitorHandler) convertMapToConsensusRequest(reqMap map[string]interface{}) (*types.ReportConsensusEventRequest, error) {
	jsonData, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request map: %w", err)
	}

	var req types.ReportConsensusEventRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &req, nil
}
