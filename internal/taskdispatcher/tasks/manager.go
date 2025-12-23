package tasks

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"

	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type TaskStreamManager struct {
	client               redisClient.RedisClientInterface
	logger               observability.Logger
	consumerGroups       map[string]bool
	mu                   sync.RWMutex
	startTime            time.Time
	aggregatorClient     *aggregator.AggregatorClient
	testAggregatorClient *aggregator.AggregatorClient
}

func NewTaskStreamManager(ctx context.Context, client redisClient.RedisClientInterface, aggClient *aggregator.AggregatorClient, testAggregatorClient *aggregator.AggregatorClient, logger observability.Logger) (*TaskStreamManager, error) {
	logger.Info(ctx, "Initializing TaskStreamManager...")

	tsm := &TaskStreamManager{
		client:               client,
		logger:               logger,
		consumerGroups:       make(map[string]bool),
		startTime:            time.Now(),
		aggregatorClient:     aggClient,
		testAggregatorClient: testAggregatorClient,
	}

	logger.Info(ctx, "TaskStreamManager initialized successfully")
	if metrics.ServiceStatus != nil {
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(ctx, 1)
	}
	return tsm, nil
}

func (tsm *TaskStreamManager) Initialize(ctx context.Context) error {
	tsm.logger.Info(ctx, "Initializing task streams...")

	ctx, cancel := context.WithTimeout(ctx, config.GetInitializationTimeout())
	defer cancel()

	// Initialize task streams with specific expiration rules
	streamConfigs := map[string]time.Duration{
		StreamTaskDispatched: TasksProcessingTTL,
		StreamTaskCompleted:  TasksCompletedTTL,
		StreamTaskFailed:     TasksFailedTTL,
		StreamTaskRetry:      TasksRetryTTL,
	}

	for stream, ttl := range streamConfigs {
		tsm.logger.Debug(ctx, "Creating stream", observability.String("stream", stream), observability.Int64("ttl", int64(ttl)))
		if err := tsm.client.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
			tsm.logger.Error(ctx, "Failed to initialize stream",
				observability.String("stream", stream),
				observability.Error(err),
				observability.Int64("ttl", int64(ttl)))
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		tsm.logger.Debug(ctx, "Stream initialized successfully", observability.String("stream", stream), observability.Int64("ttl", int64(ttl)))
	}

	// Register consumer groups for task processing
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskDispatched, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task completion
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskCompleted, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task failure
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskFailed, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task retry
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskRetry, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	go tsm.StartStreamHealthMonitor(ctx)

	tsm.logger.Info(ctx, "All task streams initialized successfully")

	return nil
}

// RegisterConsumerGroup registers a consumer group for a stream
func (tsm *TaskStreamManager) RegisterConsumerGroup(ctx context.Context, stream string, group string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		tsm.logger.Debug(ctx, "Consumer group already exists", observability.String("stream", stream), observability.String("group", group))
		return nil
	}

	tsm.logger.Debug(ctx, "Registering consumer group", observability.String("stream", stream), observability.String("group", group))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		tsm.logger.Error(ctx, "Failed to create consumer group",
			observability.String("stream", stream),
			observability.String("group", group),
			observability.Error(err))
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	tsm.consumerGroups[key] = true
	tsm.logger.Debug(ctx, "Consumer group created successfully", observability.String("stream", stream), observability.String("group", group))
	return nil
}

// GetStreamInfo returns information about task streams
func (tsm *TaskStreamManager) GetStreamInfo(ctx context.Context) map[string]interface{} {
	tsm.logger.Debug(ctx, "Getting stream information")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{StreamTaskDispatched, StreamTaskRetry, StreamTaskCompleted, StreamTaskFailed}

	for _, stream := range streams {
		length, err := tsm.client.XLen(ctx, stream)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to get stream length",
				observability.String("stream", stream),
				observability.Error(err))
			length = -1
		}
		streamLengths[stream] = length

		// Update stream length metrics
		switch stream {
		case StreamTaskDispatched:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("dispatched").Set(ctx, float64(length))
			}
		case StreamTaskRetry:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("retry").Set(ctx, float64(length))
			}
		case StreamTaskCompleted:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("completed").Set(ctx, float64(length))
			}
		case StreamTaskFailed:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("failed").Set(ctx, float64(length))
			}
		}
	}

	info := map[string]interface{}{
		"available":            tsm.client != nil,
		"max_length":           10000, // Default value, can be made configurable
		"tasks_processing_ttl": TasksProcessingTTL.String(),
		"tasks_completed_ttl":  TasksCompletedTTL.String(),
		"tasks_failed_ttl":     TasksFailedTTL.String(),
		"tasks_retry_ttl":      TasksRetryTTL.String(),
		"stream_lengths":       streamLengths,
		"max_retries":          MaxRetryAttempts,
		"consumer_groups":      len(tsm.consumerGroups),
	}

	tsm.logger.Debug(ctx, "Stream information retrieved", observability.Any("info", info))
	return info
}

func (tsm *TaskStreamManager) Close(ctx context.Context) error {
	tsm.logger.Info(ctx, "Closing TaskStreamManager")

	err := tsm.client.Close()
	if err != nil {
		tsm.logger.Error(ctx, "Failed to close Redis client", observability.Error(err))
		return err
	}

	tsm.logger.Info(ctx, "TaskStreamManager closed successfully")
	return nil
}

// storeTaskIndex stores the mapping from taskID to messageID in Redis hash
func (tsm *TaskStreamManager) storeTaskIndex(ctx context.Context, taskID int64, messageID string) error {
	start := time.Now()

	taskIDStr := strconv.FormatInt(taskID, 10)

	err := tsm.client.HSet(ctx, "task_id_to_message_id", taskIDStr, messageID)
	duration := time.Since(start)

	if err != nil {
		tsm.logger.Error(ctx, "Failed to store task index",
			observability.Int64("task_id", taskID),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to store task index: %w", err)
	}

	// Set TTL on the hash to ensure it expires (2 hours)
	err = tsm.client.SetTTL(ctx, "task_id_to_message_id", 2*time.Hour)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to set TTL on task index",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		// Don't return error as the main operation succeeded
	}

	tsm.logger.Debug(ctx, "Task index stored successfully",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return nil
}

// addTaskToTimeoutTracking adds a task to the timeout tracking sorted set
func (tsm *TaskStreamManager) addTaskToTimeoutTracking(ctx context.Context, taskID int64) error {
	start := time.Now()

	// Calculate timeout timestamp (1 hour from now)
	timeoutTimestamp := float64(time.Now().Add(TasksProcessingTTL).Unix())
	taskIDStr := strconv.FormatInt(taskID, 10)

	// Add to sorted set with timeout timestamp as score
	_, err := tsm.client.ZAdd(ctx, "dispatched_timeouts", redis.Z{
		Score:  timeoutTimestamp,
		Member: taskIDStr,
	})
	duration := time.Since(start)

	if err != nil {
		tsm.logger.Error(ctx, "Failed to add task to timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Float64("timeout_timestamp", timeoutTimestamp),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to add task to timeout tracking: %w", err)
	}

	// Set TTL on the sorted set to ensure it expires (2 hours)
	err = tsm.client.SetTTL(ctx, "dispatched_timeouts", 2*time.Hour)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to set TTL on timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		// Don't return error as the main operation succeeded
	}

	tsm.logger.Debug(ctx, "Task added to timeout tracking successfully",
		observability.Int64("task_id", taskID),
		observability.Float64("timeout_timestamp", timeoutTimestamp),
		observability.Duration("duration", duration))

	return nil
}

// startStreamHealthMonitor monitors the health of Redis streams
func (tsm *TaskStreamManager) StartStreamHealthMonitor(ctx context.Context) {
	tsm.logger.Info(ctx, "Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second) // Check health every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info(ctx, "Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			taskInfo := tsm.GetStreamInfo(ctx)

			// logger.Info("Stream health status",
			// 	"job_streams", jobInfo,
			// 	"task_streams", taskInfo)

			// Log warnings for high stream length
			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 && stream != StreamTaskFailed { // Warn if more than 50 tasks in any stream, ignore the StreamTaskFailed stream
						tsm.logger.Warn(ctx, "High task stream length detected",
							observability.String("stream", stream),
							observability.Int64("length", length))
					}
				}
			}
		}
	}
}
