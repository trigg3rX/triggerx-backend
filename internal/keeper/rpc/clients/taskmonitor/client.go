package taskmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client represents a client for communicating with the taskmonitor service
type Client struct {
	rpcClient *client.Client
	logger    observability.Logger
	tracer    observability.Tracer
}

// NewClient creates a new taskmonitor client
func NewClient(logger observability.Logger, tracer observability.Tracer) (*Client, error) {
	rpcUrl := config.GetTaskMonitorRPCUrl()
	if rpcUrl == "" {
		return nil, fmt.Errorf("task monitor RPC URL is not configured")
	}

	rpcClient := client.NewClient(client.Config{
		ServiceName: rpcUrl,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger, tracer)

	return &Client{
		rpcClient: rpcClient,
		logger:    logger,
		tracer:    tracer,
	}, nil
}

// ReportTaskExecutionStatus reports task execution status to taskmonitor
// This should be called after the aggregator submission attempt (regardless of success or failure)
func (c *Client) ReportTaskExecutionStatus(ctx context.Context, request types.ReportTaskExecutionStatusRequest) error {
	// Make RPC call
	var response types.ReportTaskExecutionStatusResponse
	err := c.rpcClient.Call(ctx, "report-task-execution-status", &request, &response)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("taskmonitor reported failure: %s", response.Message)
	}

	c.logger.Debug(ctx, "Task status reported successfully to taskmonitor",
		observability.Int64("task_id", request.TaskID),
		observability.Bool("execution_successful", request.ExecutionSuccessful),
		observability.Bool("aggregator_submitted", request.AggregatorSubmitted),
		observability.String("ipfs_data_cid", request.IPFSDataCID))

	return nil
}

// Close closes the taskmonitor client
func (c *Client) Close(ctx context.Context) error {
	if c.rpcClient != nil {
		return c.rpcClient.Close(ctx)
	}
	return nil
}
