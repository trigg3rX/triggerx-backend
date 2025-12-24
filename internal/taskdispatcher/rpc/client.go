package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcherClient provides a client for the task dispatcher service
type TaskDispatcherClient struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewTaskDispatcherClient creates a new TaskDispatcherClient
func NewTaskDispatcherClient(address string, logger observability.Logger) (*TaskDispatcherClient, error) {
	config := rpcclient.Config{
		ServiceName: "TaskDispatcher",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}

	client := rpcclient.NewClient(config, logger)

	return &TaskDispatcherClient{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the client connection
func (c *TaskDispatcherClient) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}

// SubmitTask submits a task to the task dispatcher
func (c *TaskDispatcherClient) SubmitTask(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error) {
	c.logger.Debug(ctx, "Submitting task via gRPC",
		observability.String("source", req.Source),
		observability.Int("task_count", len(req.SendTaskDataToKeeper.TaskID)))

	var response types.TaskManagerAPIResponse
	err := c.client.Call(ctx, "submit-task", req, &response)
	if err != nil {
		c.logger.Error(ctx, "gRPC call failed",
			observability.Error(err))
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	c.logger.Debug(ctx, "Task submission completed",
		observability.Bool("success", response.Success),
		observability.Int("task_count", len(response.TaskID)))

	return &response, nil
}
