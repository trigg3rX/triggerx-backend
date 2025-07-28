package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskBatchProcessor handles batch operations for better performance
type TaskBatchProcessor struct {
	batchSize    int
	batchTimeout time.Duration
	tasks        []*TaskStreamData
	mu           sync.Mutex
	ticker       *time.Ticker
	done         chan struct{}
	streamMgr    *TaskStreamManager
}

// NewTaskBatchProcessor creates a new batch processor for tasks
func NewTaskBatchProcessor(batchSize int, batchTimeout time.Duration, streamMgr *TaskStreamManager) *TaskBatchProcessor {
	return &TaskBatchProcessor{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		tasks:        make([]*TaskStreamData, 0, batchSize),
		done:         make(chan struct{}),
		streamMgr:    streamMgr,
	}
}

// Start begins the batch processing loop
func (bp *TaskBatchProcessor) Start() {
	bp.mu.Lock()
	bp.ticker = time.NewTicker(bp.batchTimeout)
	bp.mu.Unlock()

	go bp.processLoop()
}

// Stop gracefully stops the batch processor
func (bp *TaskBatchProcessor) Stop() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.ticker != nil {
		bp.ticker.Stop()
	}
	close(bp.done)
}

// AddTask adds a task to the batch processor
func (bp *TaskBatchProcessor) AddTask(task *TaskStreamData) error {
	bp.streamMgr.logger.Debug("Adding task to batch processor",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"current_batch_size", len(bp.tasks))

	bp.mu.Lock()
	bp.tasks = append(bp.tasks, task)
	shouldProcess := len(bp.tasks) >= bp.batchSize
	bp.mu.Unlock()

	bp.streamMgr.logger.Debug("Task added to batch",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"new_batch_size", len(bp.tasks),
		"should_process", shouldProcess)

	// Process batch immediately if it reaches the batch size
	// Note: We release the lock before processing to avoid deadlocks
	if shouldProcess {
		bp.streamMgr.logger.Info("Processing batch immediately due to size limit", "batch_size", len(bp.tasks))
		return bp.processBatch()
	}

	return nil
}

// processLoop runs the main processing loop
func (bp *TaskBatchProcessor) processLoop() {
	bp.streamMgr.logger.Info("Batch processor loop started")

	for {
		select {
		case <-bp.done:
			bp.streamMgr.logger.Info("Batch processor stopping")
			// Process any remaining tasks before stopping
			bp.mu.Lock()
			if len(bp.tasks) > 0 {
				bp.streamMgr.logger.Info("Processing final batch before stopping", "batch_size", len(bp.tasks))
				if err := bp.processBatch(); err != nil {
					bp.streamMgr.logger.Error("Failed to process final batch", "error", err)
				}
			}
			bp.mu.Unlock()
			return
		case <-bp.ticker.C:
			bp.mu.Lock()
			if len(bp.tasks) > 0 {
				bp.streamMgr.logger.Debug("Processing batch due to timeout", "batch_size", len(bp.tasks))
				if err := bp.processBatch(); err != nil {
					bp.streamMgr.logger.Error("Failed to process batch", "error", err)
				}
			}
			bp.mu.Unlock()
		}
	}
}

// processBatch processes the current batch of tasks
func (bp *TaskBatchProcessor) processBatch() error {
	bp.mu.Lock()
	if len(bp.tasks) == 0 {
		bp.mu.Unlock()
		return nil
	}

	// Copy tasks to avoid holding lock during processing
	tasksToProcess := make([]*TaskStreamData, len(bp.tasks))
	copy(tasksToProcess, bp.tasks)
	bp.tasks = bp.tasks[:0] // Clear the slice
	bp.mu.Unlock()

	bp.streamMgr.logger.Info("Processing batch of tasks",
		"batch_size", len(tasksToProcess),
		"batch_timeout", bp.batchTimeout)

	// Process tasks with improved error handling
	var wg sync.WaitGroup
	errors := make(chan error, len(tasksToProcess))

	for _, task := range tasksToProcess {
		wg.Add(1)
		go func(t *TaskStreamData) {
			defer wg.Done()
			if err := bp.processSingleTask(t); err != nil {
				errors <- err
				bp.streamMgr.logger.Error("Failed to process task in batch",
					"task_id", t.SendTaskDataToKeeper.TaskID[0],
					"error", err)
			}
		}(task)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		bp.streamMgr.logger.Error("Batch processing completed with errors",
			"error_count", len(errs),
			"total_tasks", len(tasksToProcess))
		// Don't return error to avoid blocking the batch processor
		// Instead, log the errors and continue
	}

	bp.streamMgr.logger.Info("Batch processing completed",
		"processed_tasks", len(tasksToProcess),
		"error_count", len(errs))
	return nil
}

// processSingleTask processes a single task from the batch
func (bp *TaskBatchProcessor) processSingleTask(task *TaskStreamData) error {
	bp.streamMgr.logger.Debug("Processing single task in batch",
		"task_id", task.SendTaskDataToKeeper.TaskID[0])

	// Add task to ready stream
	performerData, err := bp.streamMgr.AddTaskToReadyStream(*task)
	if err != nil {
		bp.streamMgr.logger.Error("Failed to process task in batch",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return err
	}

	bp.streamMgr.logger.Info("Successfully processed single task in batch",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"performer_id", performerData.OperatorID,
		"performer_address", performerData.KeeperAddress)
	return nil
}

// GetBatchStats returns current batch statistics
func (bp *TaskBatchProcessor) GetBatchStats() map[string]interface{} {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	return map[string]interface{}{
		"current_batch_size": len(bp.tasks),
		"batch_size_limit":   bp.batchSize,
		"batch_timeout":      bp.batchTimeout,
		"is_running":         bp.ticker != nil,
	}
}

// Ready to be sent to the performer with improved error handling and performance
func (tsm *TaskStreamManager) AddTaskToReadyStream(task TaskStreamData) (types.PerformerData, error) {
	// Use dynamic performer selection instead of hardcoded selection
	performerData, err := tsm.performerManager.GetPerformerData(task.SendTaskDataToKeeper.TargetData[0].IsImua)
	if err != nil {
		tsm.logger.Error("Failed to get performer data dynamically",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to get performer: %w", err)
	}

	// Update task with performer information
	task.SendTaskDataToKeeper.PerformerData = performerData

	// Sign the task data with improved error handling
	signature, err := cryptography.SignJSONMessage(task.SendTaskDataToKeeper, config.GetRedisSigningKey())
	if err != nil {
		tsm.logger.Error("Failed to sign batch task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to sign task data: %w", err)
	}
	task.SendTaskDataToKeeper.ManagerSignature = signature

	// Add task to stream with improved performance
	err = tsm.addTaskToStream(TasksReadyStream, &task)
	if err != nil {
		tsm.logger.Error("Failed to add task to ready stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"performer_id", performerData.OperatorID,
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to add task to ready stream: %w", err)
	}

	// Send task to performer asynchronously for better performance
	go func() {
		tsm.sendTaskToPerformer(task)
	}()

	tsm.logger.Info("Task added to ready stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"performer_id", performerData.OperatorID,
		"performer_address", performerData.KeeperAddress)

	return performerData, nil
}

// Failed to send to the performer, sent to the retry stream with improved backoff strategy
func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	tsm.logger.Warn("Adding task to retry stream",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"job_id", task.JobID,
		"retry_count", task.RetryCount,
		"retry_reason", retryReason)

	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now

	// Improved exponential backoff with jitter and maximum cap
	baseBackoff := time.Duration(task.RetryCount) * RetryBackoffBase
	maxBackoff := config.GetMaxRetryBackoff()
	if baseBackoff > maxBackoff {
		baseBackoff = maxBackoff
	}

	jitter := time.Duration(rand.Int63n(int64(RetryBackoffBase))) // Up to 1 base backoff
	backoffDuration := baseBackoff + jitter

	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	tsm.logger.Info("Task retry scheduled",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"retry_count", task.RetryCount,
		"scheduled_for", scheduledFor.Format(time.RFC3339),
		"backoff_duration", backoffDuration)

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Error("Task exceeded max retry attempts, moving to failed stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"retry_count", task.RetryCount,
			"max_attempts", MaxRetryAttempts)

		metrics.TaskMaxRetriesExceededTotal.Inc()
		metrics.TasksMovedToFailedStreamTotal.Inc()

		err := tsm.addTaskToStream(TasksFailedStream, task)
		if err != nil {
			tsm.logger.Error("Failed to add task to failed stream",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"error", err)
			return fmt.Errorf("failed to add task to failed stream: %w", err)
		}

		// Notify scheduler about task failure
		// tsm.notifySchedulerTaskComplete(task, false)

		tsm.logger.Error("Task permanently failed",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"final_retry_count", task.RetryCount)

		return nil
	}

	err := tsm.addTaskToStream(TasksRetryStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to retry stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "failure").Inc()
		return fmt.Errorf("failed to add task to retry stream: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	return nil
}

func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.GetStreamOperationTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"error", err)
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc()
		tsm.logger.Error("Failed to add task to stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to add task to stream: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return nil
}
