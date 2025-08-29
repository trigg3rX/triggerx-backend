package tasks

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"

	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type TaskStreamManager struct {
	client           redisClient.RedisClientInterface
	logger           logging.Logger
	consumerGroups   map[string]bool
	mu               sync.RWMutex
	startTime        time.Time
	aggregatorClient *aggregator.AggregatorClient
	testAggregatorClient *aggregator.AggregatorClient
}

func NewTaskStreamManager(client redisClient.RedisClientInterface, aggClient *aggregator.AggregatorClient, testAggregatorClient *aggregator.AggregatorClient, logger logging.Logger) (*TaskStreamManager, error) {
	logger.Info("Initializing TaskStreamManager...")

	tsm := &TaskStreamManager{
		client:           client,
		logger:           logger,
		consumerGroups:   make(map[string]bool),
		startTime:        time.Now(),
		aggregatorClient:  aggClient,
		testAggregatorClient: testAggregatorClient,
	}

	logger.Info("TaskStreamManager initialized successfully")
	metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(1)
	return tsm, nil
}

func (tsm *TaskStreamManager) Initialize() error {
	tsm.logger.Info("Initializing task streams...")

	ctx, cancel := context.WithTimeout(context.Background(), config.GetInitializationTimeout())
	defer cancel()

	// Initialize task streams with specific expiration rules
	streamConfigs := map[string]time.Duration{
		StreamTaskDispatched: TasksProcessingTTL,
		StreamTaskCompleted:  TasksCompletedTTL,
		StreamTaskFailed:     TasksFailedTTL,
		StreamTaskRetry:      TasksRetryTTL,
	}

	for stream, ttl := range streamConfigs {
		tsm.logger.Debug("Creating stream", "stream", stream, "ttl", ttl)
		if err := tsm.client.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
			tsm.logger.Error("Failed to initialize stream",
				"stream", stream,
				"error", err,
				"ttl", ttl)
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		tsm.logger.Info("Stream initialized successfully", "stream", stream, "ttl", ttl)
	}

	// Register consumer groups for task processing
	if err := tsm.RegisterConsumerGroup(StreamTaskDispatched, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task completion
	if err := tsm.RegisterConsumerGroup(StreamTaskCompleted, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task failure
	if err := tsm.RegisterConsumerGroup(StreamTaskFailed, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task retry
	if err := tsm.RegisterConsumerGroup(StreamTaskRetry, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	go tsm.StartStreamHealthMonitor(ctx)

	tsm.logger.Info("All task streams initialized successfully")

	return nil
}

// RegisterConsumerGroup registers a consumer group for a stream
func (tsm *TaskStreamManager) RegisterConsumerGroup(stream string, group string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		tsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
		return nil
	}

	tsm.logger.Info("Registering consumer group", "stream", stream, "group", group)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		tsm.logger.Error("Failed to create consumer group",
			"stream", stream,
			"group", group,
			"error", err)
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	tsm.consumerGroups[key] = true
	tsm.logger.Info("Consumer group created successfully", "stream", stream, "group", group)
	return nil
}

// GetStreamInfo returns information about task streams
func (tsm *TaskStreamManager) GetStreamInfo() map[string]interface{} {
	tsm.logger.Debug("Getting stream information")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{StreamTaskDispatched, StreamTaskRetry, StreamTaskCompleted, StreamTaskFailed}

	for _, stream := range streams {
		length, err := tsm.client.XLen(ctx, stream)
		if err != nil {
			tsm.logger.Warn("Failed to get stream length",
				"stream", stream,
				"error", err)
			length = -1
		}
		streamLengths[stream] = length

		// Update stream length metrics
		switch stream {
		case StreamTaskDispatched:
			metrics.TaskStreamLengths.WithLabelValues("dispatched").Set(float64(length))
		case StreamTaskRetry:
			metrics.TaskStreamLengths.WithLabelValues("retry").Set(float64(length))
		case StreamTaskCompleted:
			metrics.TaskStreamLengths.WithLabelValues("completed").Set(float64(length))
		case StreamTaskFailed:
			metrics.TaskStreamLengths.WithLabelValues("failed").Set(float64(length))
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

	tsm.logger.Debug("Stream information retrieved", "info", info)
	return info
}

func (tsm *TaskStreamManager) Close() error {
	tsm.logger.Info("Closing TaskStreamManager")

	err := tsm.client.Close()
	if err != nil {
		tsm.logger.Error("Failed to close Redis client", "error", err)
		return err
	}

	tsm.logger.Info("TaskStreamManager closed successfully")
	return nil
}

// storeTaskIndex stores the mapping from taskID to messageID in Redis hash
func (tsm *TaskStreamManager) storeTaskIndex(ctx context.Context, taskID int64, messageID string) error {
	start := time.Now()

	taskIDStr := strconv.FormatInt(taskID, 10)

	err := tsm.client.HSet(ctx, "task_id_to_message_id", taskIDStr, messageID)
	duration := time.Since(start)

	if err != nil {
		tsm.logger.Error("Failed to store task index",
			"task_id", taskID,
			"message_id", messageID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to store task index: %w", err)
	}

	// Set TTL on the hash to ensure it expires (2 hours)
	err = tsm.client.SetTTL(ctx, "task_id_to_message_id", 2*time.Hour)
	if err != nil {
		tsm.logger.Warn("Failed to set TTL on task index",
			"task_id", taskID,
			"error", err)
		// Don't return error as the main operation succeeded
	}

	tsm.logger.Debug("Task index stored successfully",
		"task_id", taskID,
		"message_id", messageID,
		"duration", duration)

	return nil
}

// startStreamHealthMonitor monitors the health of Redis streams
func (tsm *TaskStreamManager) StartStreamHealthMonitor(ctx context.Context) {
	tsm.logger.Info("Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second) // Check health every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			taskInfo := tsm.GetStreamInfo()

			// logger.Info("Stream health status",
			// 	"job_streams", jobInfo,
			// 	"task_streams", taskInfo)

			// Log warnings for high stream length
			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 && stream != StreamTaskFailed { // Warn if more than 50 tasks in any stream, ignore the StreamTaskFailed stream
						tsm.logger.Warn("High task stream length detected",
							"stream", stream,
							"length", length)
					}
				}
			}
		}
	}
}
