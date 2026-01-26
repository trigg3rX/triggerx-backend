package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// submitBatchToTaskDispatcher submits the batch task data to Task Dispatcher via RPC.
// It implements retry logic with exponential backoff for handling transient failures.
// Returns true if the submission was successful, false otherwise.
func (s *TimeBasedScheduler) submitBatchToTaskDispatcher(ctx context.Context, request types.SchedulerTaskRequest, taskIDs string, taskCount int) bool {
	startTime := time.Now()

	// Create retry configuration for task dispatcher calls
	retryConfig := &retry.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.2,
		ShouldRetry: func(err error, attempt int) bool {
			// Retry on network errors, timeouts, and temporary failures
			// Don't retry on permanent errors like invalid requests
			return err != nil && !strings.Contains(err.Error(), "invalid") &&
				!strings.Contains(err.Error(), "permission denied")
		},
	}

	// Create context with timeout for the entire retry operation
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Define the operation to retry
	operation := func() (bool, error) {
		// Create context with timeout for individual RPC call
		rpcCtx, rpcCancel := context.WithTimeout(ctx, 30*time.Second)
		defer rpcCancel()

		// Extract trace ID for logging
		span := trace.SpanFromContext(rpcCtx)
		traceID := span.SpanContext().TraceID().String()
		if traceID != "" {
			s.logger.Debug(rpcCtx, "Propagating trace to task dispatcher",
				observability.String("trace_id", traceID),
				observability.String("task_ids", taskIDs),
			)
		}

		// Make RPC call to task dispatcher
		var response types.TaskDispatcherRPCResponse
		err := s.taskDispatcherClient.Call(rpcCtx, "submit-task", &request, &response)
		if err != nil {
			return false, fmt.Errorf("RPC call failed: %w", err)
		}

		if !response.Success {
			return false, fmt.Errorf("task dispatcher processing failed: %s - %s", response.Message, response.Error)
		}

		return true, nil
	}

	// Execute with retry logic
	success, err := retry.Retry(ctx, operation, retryConfig)
	if err != nil {
		duration := time.Since(startTime)
		s.logger.Error(ctx, "Failed to submit batch to task dispatcher after retries",
			observability.String("task_ids", taskIDs),
			observability.Int("task_count", taskCount),
			observability.Error(err),
			observability.Duration("duration", duration))
		return false
	}

	duration := time.Since(startTime)
	s.logger.Debug(ctx, "Successfully submitted batch to task dispatcher",
		observability.String("task_ids", taskIDs),
		observability.Int("task_count", taskCount),
		observability.Duration("duration", duration))

	return success
}
