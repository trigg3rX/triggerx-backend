package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcherClient provides a client for the task dispatcher service
type TaskDispatcherClient struct {
	client *rpcclient.Client
	logger logging.Logger
}

// NewTaskDispatcherClient creates a new TaskDispatcherClient
func NewTaskDispatcherClient(address string, logger logging.Logger) (*TaskDispatcherClient, error) {
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
func (c *TaskDispatcherClient) Close() error {
	return c.client.Close()
}

// SubmitTask submits a task to the task dispatcher
func (c *TaskDispatcherClient) SubmitTask(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error) {
	c.logger.Debug("Submitting task via gRPC",
		"source", req.Source,
		"task_count", len(req.SendTaskDataToKeeper.TaskID))

	var response types.TaskManagerAPIResponse
	err := c.client.Call(ctx, "submit-task", req, &response)
	if err != nil {
		c.logger.Error("gRPC call failed",
			"error", err)
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	c.logger.Debug("Task submission completed",
		"success", response.Success,
		"task_count", len(response.TaskID))

	return &response, nil
}
