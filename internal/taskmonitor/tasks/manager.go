package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type TaskStreamManager struct {
	redisClient    redisClient.RedisClientInterface
	dbClient       *database.DatabaseClient
	logger         logging.Logger
	consumerGroups map[string]bool
	mu             sync.RWMutex
	startTime      time.Time
	taskIndex      *TaskIndexManager
	timeoutManager *TimeoutManager
}

func NewTaskStreamManager(redisClient redisClient.RedisClientInterface, dbClient *database.DatabaseClient, logger logging.Logger) (*TaskStreamManager, error) {
	tsm := &TaskStreamManager{
		redisClient:    redisClient,
		dbClient:       dbClient,
		logger:         logger,
		consumerGroups: make(map[string]bool),
		startTime:      time.Now(),
	}

	// Initialize the task index manager
	tsm.taskIndex = NewTaskIndexManager(tsm)

	// Initialize the timeout manager
	tsm.timeoutManager = NewTimeoutManager(tsm)

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
		if err := tsm.redisClient.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
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

	// Register consumer groups for timeout checking
	if err := tsm.RegisterConsumerGroup(StreamTaskDispatched, "timeout-checker"); err != nil {
		return fmt.Errorf("failed to register timeout-checker group: %w", err)
	}

	// Register consumer groups for task finding
	if err := tsm.RegisterConsumerGroup(StreamTaskDispatched, "task-finder"); err != nil {
		return fmt.Errorf("failed to register task-finder group: %w", err)
	}

	// go tsm.StartStreamHealthMonitor(ctx)

	tsm.logger.Info("All task streams initialized successfully")

	return nil
}

// RegisterConsumerGroup registers a consumer group for a stream
func (tsm *TaskStreamManager) RegisterConsumerGroup(stream string, group string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		// tsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
		return nil
	}

	tsm.logger.Info("Registering consumer group", "stream", stream, "group", group)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tsm.redisClient.CreateConsumerGroup(ctx, stream, group); err != nil {
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
	// tsm.logger.Debug("Getting stream information")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{StreamTaskDispatched, StreamTaskRetry, StreamTaskCompleted, StreamTaskFailed}

	for _, stream := range streams {
		length, err := tsm.redisClient.XLen(ctx, stream)
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
			metrics.TaskStreamLengths.WithLabelValues("ready").Set(float64(length))
		case StreamTaskRetry:
			metrics.TaskStreamLengths.WithLabelValues("retry").Set(float64(length))
		case StreamTaskCompleted:
			metrics.TaskStreamLengths.WithLabelValues("completed").Set(float64(length))
		case StreamTaskFailed:
			metrics.TaskStreamLengths.WithLabelValues("failed").Set(float64(length))
		}
	}

	info := map[string]interface{}{
		"available":            tsm.redisClient != nil,
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

// GetDatabaseClient returns the database client
func (tsm *TaskStreamManager) GetDatabaseClient() *database.DatabaseClient {
	return tsm.dbClient
}

// GetTaskIndexManager returns the task index manager
func (tsm *TaskStreamManager) GetTaskIndexManager() *TaskIndexManager {
	return tsm.taskIndex
}

// GetTimeoutManager returns the timeout manager
func (tsm *TaskStreamManager) GetTimeoutManager() *TimeoutManager {
	return tsm.timeoutManager
}

// GetPendingEntriesInfo returns information about pending entries in consumer groups
func (tsm *TaskStreamManager) GetPendingEntriesInfo() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pendingInfo := make(map[string]interface{})
	consumerGroups := []string{"task-processors", "timeout-checker", "task-finder"}

	for _, group := range consumerGroups {
		pending, err := tsm.redisClient.XPending(ctx, StreamTaskDispatched, group)
		if err != nil {
			tsm.logger.Warn("Failed to get pending entries info",
				"consumer_group", group,
				"error", err)
			pendingInfo[group] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}

		pendingInfo[group] = map[string]interface{}{
			"count":     pending.Count,
			"min_id":    pending.Lower,
			"max_id":    pending.Higher,
			"consumers": pending.Consumers,
		}

		// Log warning if there are too many pending entries
		if pending.Count > 100 {
			tsm.logger.Warn("High number of pending entries detected",
				"consumer_group", group,
				"pending_count", pending.Count,
				"min_id", pending.Lower,
				"max_id", pending.Higher)
		}
	}

	return pendingInfo
}

// CleanupPendingEntries attempts to acknowledge old pending entries
func (tsm *TaskStreamManager) CleanupPendingEntries(ctx context.Context) error {
	consumerGroups := []string{"task-processors", "timeout-checker", "task-finder"}
	totalCleaned := 0

	for _, group := range consumerGroups {
		// Get pending entries for this group
		pendingExt, err := tsm.redisClient.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: StreamTaskDispatched,
			Group:  group,
			Start:  "-",
			End:    "+",
			Count:  100, // Limit to 100 entries per cleanup
		})
		if err != nil {
			tsm.logger.Warn("Failed to get pending entries for cleanup",
				"consumer_group", group,
				"error", err)
			continue
		}

		cleaned := 0
		for _, entry := range pendingExt {
			// If the entry is older than 1 hour, acknowledge it to prevent PEL growth
			if entry.Idle > time.Hour {
				err := tsm.redisClient.XAck(ctx, StreamTaskDispatched, group, entry.ID)
				if err != nil {
					tsm.logger.Error("Failed to acknowledge old pending entry",
						"consumer_group", group,
						"message_id", entry.ID,
						"idle_time", entry.Idle,
						"error", err)
				} else {
					cleaned++
					tsm.logger.Debug("Cleaned up old pending entry",
						"consumer_group", group,
						"message_id", entry.ID,
						"idle_time", entry.Idle)
				}
			}
		}

		if cleaned > 0 {
			tsm.logger.Info("Cleaned up pending entries",
				"consumer_group", group,
				"cleaned_count", cleaned)
			totalCleaned += cleaned
		}
	}

	if totalCleaned > 0 {
		tsm.logger.Info("Pending entries cleanup completed",
			"total_cleaned", totalCleaned)
	}

	return nil
}

// FindTaskInDispatched finds a specific task in the dispatched stream
func (tsm *TaskStreamManager) FindTaskInDispatched(taskID int64) (*TaskStreamData, error) {
	ctx := context.Background()
	task, _, err := tsm.taskIndex.FindTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// AddTaskToStream adds a task to a specific stream
func (tsm *TaskStreamManager) AddTaskToStream(ctx context.Context, stream string, task *TaskStreamData) error {
	return tsm.addTaskToStream(ctx, stream, task)
}

func (tsm *TaskStreamManager) Close() error {
	tsm.logger.Info("Closing TaskStreamManager")

	err := tsm.redisClient.Close()
	if err != nil {
		tsm.logger.Error("Failed to close Redis client", "error", err)
		return err
	}

	tsm.logger.Info("TaskStreamManager closed successfully")
	return nil
}

// startStreamHealthMonitor monitors the health of Redis streams
func (tsm *TaskStreamManager) StartStreamHealthMonitor(ctx context.Context) {
	tsm.logger.Info("Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second)       // Check health every 30 seconds
	cleanupTicker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			taskInfo := tsm.GetStreamInfo()

			// Log warnings for high stream lengths
			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 { // Warn if more than 50 tasks in any stream
						tsm.logger.Warn("High task stream length detected",
							"stream", stream,
							"length", length)
					}
				}
			}

			// Check pending entries
			pendingInfo := tsm.GetPendingEntriesInfo()
			for group, info := range pendingInfo {
				if infoMap, ok := info.(map[string]interface{}); ok {
					if count, exists := infoMap["count"]; exists {
						if countInt, ok := count.(int64); ok && countInt > 50 {
							tsm.logger.Warn("High number of pending entries detected",
								"consumer_group", group,
								"pending_count", countInt)
						}
					}
				}
			}
		case <-cleanupTicker.C:
			// Periodic cleanup of old pending entries
			if err := tsm.CleanupPendingEntries(ctx); err != nil {
				tsm.logger.Error("Failed to cleanup pending entries", "error", err)
			}
		}
	}
}
