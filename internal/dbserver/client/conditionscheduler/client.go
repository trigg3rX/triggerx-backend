package conditionscheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client handles gRPC communication with the condition scheduler
type Client struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewClient creates a new condition scheduler gRPC client
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

// ScheduleJob schedules a job via gRPC
func (c *Client) ScheduleJob(ctx context.Context, jobData *types.ScheduleConditionJobData) error {
	var response map[string]interface{}
	if err := c.client.Call(ctx, "schedule-job", jobData, &response); err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	// Check response
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return fmt.Errorf("schedule job failed: %s", errorMsg)
	}

	c.logger.Info(ctx, "Job scheduled successfully via gRPC",
		observability.String("job_id", jobData.JobID.String()))

	return nil
}

// UnscheduleJob unschedules a job via gRPC
func (c *Client) UnscheduleJob(ctx context.Context, jobID *types.BigInt) error {
	req := map[string]interface{}{
		"job_id": jobID.String(),
	}

	var response map[string]interface{}
	if err := c.client.Call(ctx, "unschedule-job", req, &response); err != nil {
		return fmt.Errorf("failed to unschedule job: %w", err)
	}

	// Check response
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return fmt.Errorf("unschedule job failed: %s", errorMsg)
	}

	c.logger.Info(ctx, "Job unscheduled successfully via gRPC",
		observability.String("job_id", jobID.String()))

	return nil
}

// Close closes the gRPC client
func (c *Client) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}
