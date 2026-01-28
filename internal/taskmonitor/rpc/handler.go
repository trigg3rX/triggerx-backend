package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskMonitorHandler implements the generic RPC handler interface
type TaskMonitorHandler struct {
	logger   observability.Logger
	monitor  TaskMonitorInterface
}

// TaskMonitorInterface defines the interface for task monitor operations
type TaskMonitorInterface interface {
	// ReportTaskStatus handles task status reports from keepers (both success and failure)
	ReportTaskExecutionStatus(ctx context.Context, req *types.ReportTaskExecutionStatusRequest) (*types.ReportTaskExecutionStatusResponse, error)
	// ReportConsensusEvent handles consensus event reports from eventmonitor (TaskSubmitted or TaskRejected)
	ReportTaskConsensusStatus(ctx context.Context, req *types.ReportTaskConsensusStatusRequest) (*types.ReportTaskConsensusStatusResponse, error)
}

// NewTaskMonitorHandler creates a new RPC handler
func NewTaskMonitorHandler(logger observability.Logger, monitor TaskMonitorInterface) *TaskMonitorHandler {
	return &TaskMonitorHandler{
		logger:   logger,
		monitor:  monitor,
	}
}

// Handle routes incoming RPC requests based on the method name
func (h *TaskMonitorHandler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	switch method {

	case "report-task-execution-status":
		// Convert request to the expected type
		req, ok := request.(*types.ReportTaskExecutionStatusRequest)
		if !ok {
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				var err error
				req, err = h.convertMapToStatusRequest(reqMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid request type for report-task-execution-status: %T", request)
			}
		}

		resp, err := h.monitor.ReportTaskExecutionStatus(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	case "report-task-consensus-status":
		// Convert request to the expected type
		consensusReq, ok := request.(*types.ReportTaskConsensusStatusRequest)
		if !ok {
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				var err error
				consensusReq, err = h.convertMapToConsensusRequest(reqMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid request type for report-task-consensus-status: %T", request)
			}
		}

		resp, err := h.monitor.ReportTaskConsensusStatus(ctx, consensusReq)
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
			Name:         "report-task-execution-status",
			Description:  "Report task execution status from a keeper (success or failure)",
			RequestType:  &types.ReportTaskExecutionStatusRequest{},
			ResponseType: &types.ReportTaskExecutionStatusResponse{},
			Timeout:      30 * time.Second,
		},
		{
			Name:         "report-task-consensus-status",
			Description:  "Report consensus event from eventmonitor (TaskSubmitted or TaskRejected)",
			RequestType:  &types.ReportTaskConsensusStatusRequest{},
			ResponseType: &types.ReportTaskConsensusStatusResponse{},
			Timeout:      30 * time.Second,
		},
	}
}

// convertMapToStatusRequest converts a map to ReportTaskStatusRequest
// This is used when the request comes as JSON-decoded map
func (h *TaskMonitorHandler) convertMapToStatusRequest(reqMap map[string]interface{}) (*types.ReportTaskExecutionStatusRequest, error) {
	jsonData, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request map: %w", err)
	}

	var req types.ReportTaskExecutionStatusRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &req, nil
}

// convertMapToConsensusRequest converts a map to ReportTaskConsensusStatusRequest
func (h *TaskMonitorHandler) convertMapToConsensusRequest(reqMap map[string]interface{}) (*types.ReportTaskConsensusStatusRequest, error) {
	jsonData, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request map: %w", err)
	}

	var req types.ReportTaskConsensusStatusRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &req, nil
}
