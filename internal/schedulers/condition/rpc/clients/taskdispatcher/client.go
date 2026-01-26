package taskdispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client handles gRPC communication with the task dispatcher
type Client struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewClient creates a new task dispatcher gRPC client
func NewClient(serviceURL string, logger observability.Logger, tracer observability.Tracer) (*Client, error) {
	client := rpcclient.NewClient(rpcclient.Config{
		ServiceName: serviceURL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger, tracer)

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// SubmitTask submits a task to the task dispatcher via gRPC
func (c *Client) SubmitTask(ctx context.Context, request *types.SchedulerTaskRequest) (*types.TaskDispatcherRPCResponse, error) {
	var response types.TaskDispatcherRPCResponse
	if err := c.client.Call(ctx, "submit-task", request, &response); err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	if !response.Success {
		return &response, fmt.Errorf("task dispatcher processing failed: %s - %s", response.Message, response.Error)
	}

	return &response, nil
}

// Call makes a generic RPC call to the task dispatcher service.
func (c *Client) Call(ctx context.Context, method string, request interface{}, response interface{}) error {
	return c.client.Call(ctx, method, request, response)
}

// Close closes the gRPC client
func (c *Client) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}
