package redis

import (
	"context"
	"fmt"
	"time"
)

// WaitForAggregatorResponse waits for a response from the aggregator for a given job.
// If a response is received within 5 seconds, moves the job to TasksProcessingStream.
// If not, increments RetryCount and moves to TasksRetryStream, or TasksFailedStream if retries exhausted.
func WaitForAggregatorResponse(tsm *TaskStreamManager, job *TaskStreamData, performerID int64) error {
	const (
		responseTimeout = 5 * time.Second
		pollInterval    = 100 * time.Millisecond
	)
	ctx, cancel := context.WithTimeout(context.Background(), responseTimeout)
	defer cancel()

	redisKey := fmt.Sprintf("aggregator:response:%d", job.JobID)
	found := false

outerLoop:
	for {
		select {
		case <-ctx.Done():
			found = false
			break outerLoop
		default:
			// Try to get the response from Redis
			val, err := tsm.client.Get(context.Background(), redisKey)
			if err != nil {
				tsm.logger.Warnf("Error checking aggregator response for job %d: %v", job.JobID, err)
			}
			if val != "" {
				found = true
				break outerLoop
			}
			time.Sleep(pollInterval)
		}
		if found || ctx.Err() != nil {
			break
		}
	}

	if found {
		// Clean up the response key
		_ = tsm.client.Del(context.Background(), redisKey)
		// Move to processing stream
		return tsm.AddTaskToProcessingStream(job, performerID)
	}

	// Not found: handle retry logic
	job.RetryCount++
	if job.RetryCount >= MaxRetryAttempts {
		tsm.logger.Warnf("Job %d exceeded max aggregator response retries, moving to failed stream", job.JobID)
		return tsm.AddTaskToFailedStream(job)
	}
	tsm.logger.Infof("No aggregator response for job %d, retrying (attempt %d)", job.JobID, job.RetryCount)
	return tsm.AddTaskToRetryStream(job, "no aggregator response")
}

// AddTaskToFailedStream moves a job to the failed stream.
func (tsm *TaskStreamManager) AddTaskToFailedStream(task *TaskStreamData) error {
	return tsm.addTaskToStream(TasksFailedStream, task)
}
