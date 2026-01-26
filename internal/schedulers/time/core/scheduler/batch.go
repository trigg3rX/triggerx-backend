package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// processBatch processes a batch of tasks by submitting them to the task dispatcher.
// It filters out expired tasks, builds the task data structures, and submits
// the batch via RPC to the task dispatcher service.
func (s *TimeBasedScheduler) processBatch(ctx context.Context, tasks types.SendTaskDataToKeeper) {
	s.logger.Debug(ctx, "Processing batch of time-based tasks", observability.Int("task_count", len(tasks.TaskID)))

	for _, triggerData := range tasks.TriggerData {
		// Final safety check: This should not happen as expired jobs are filtered before task creation
		// This check catches edge cases where a job expired between fetching and processing
		if triggerData.ExpirationTime.Before(time.Now()) {
			s.logger.Warn(ctx, "Task has expired (unexpected - should have been filtered earlier), skipping execution",
				observability.Int64("task_id", triggerData.TaskID),
				observability.Time("expiration_time", triggerData.ExpirationTime))
			metrics.TrackTaskExpired()
			continue
		}

		// Track task by schedule type
		metrics.TrackTaskByScheduleType(triggerData.TimeScheduleType)
	}

	// If no valid tasks, return early
	if len(tasks.TaskID) == 0 {
		s.logger.Debug(ctx, "No valid tasks in batch after filtering expired tasks")
		return
	}

	// Create request for task dispatcher
	request := types.SchedulerTaskRequest{
		SendTaskDataToKeeper: tasks,
		Source:               "time_scheduler",
	}

	// Convert validTaskIDs ([]int64) to []string for joining
	taskIDStrs := make([]string, len(tasks.TaskID))
	for i, id := range tasks.TaskID {
		taskIDStrs[i] = fmt.Sprintf("%d", id)
	}
	taskIDs := strings.Join(taskIDStrs, ", ")

	// Submit batch to task dispatcher
	success := s.submitBatchToTaskDispatcher(ctx, request, taskIDs, len(tasks.TaskID))

	if success {
		s.logger.Info(ctx, "Batch processing completed successfully", observability.Int("task_count", len(tasks.TaskID)))
		metrics.UpdateTasksDispatched("success")
	} else {
		s.logger.Error(ctx, "Batch processing failed", observability.Int("task_count", len(tasks.TaskID)))
		metrics.UpdateTasksDispatched("failed")
	}
}
