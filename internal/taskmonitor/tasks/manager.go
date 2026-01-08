package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type TaskStreamManager struct {
	redisClient       redisClient.RedisClientInterface
	dbClient          *database.DatabaseClient
	notifier          notify.Notifier
	logger            observability.Logger
	tracer            observability.Tracer
	consumerGroups    map[string]bool
	mu                sync.RWMutex
	startTime         time.Time
	taskIndex         *TaskIndexManager
	expirationManager *ExpirationManager
	keeperClient      KeeperClientInterface
}

// KeeperClientInterface defines the interface for keeper client operations
type KeeperClientInterface interface {
	RebroadcastTask(ctx context.Context, taskID int64) error
}

func NewTaskStreamManager(ctx context.Context, redisClient redisClient.RedisClientInterface, dbClient *database.DatabaseClient, logger observability.Logger, tracer observability.Tracer, keeperClient KeeperClientInterface) (*TaskStreamManager, error) {
	tsm := &TaskStreamManager{
		redisClient:    redisClient,
		dbClient:       dbClient,
		notifier:       notify.NewCompositeNotifier(logger, notify.NewWebhookNotifier(logger), notify.NewSMTPNotifier(logger)),
		logger:         logger,
		tracer:         tracer,
		consumerGroups: make(map[string]bool),
		startTime:      time.Now(),
		keeperClient:   keeperClient,
	}

	// Initialize the task index manager
	tsm.taskIndex = NewTaskIndexManager(tsm)

	// Initialize the unified expiration manager (handles both task timeouts and stream entry expirations)
	tsm.expirationManager = NewExpirationManager(tsm)

	logger.Info(ctx, "TaskStreamManager initialized successfully", observability.String("component", "task_stream_manager"))
	if metrics.ServiceStatus != nil {
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(ctx, 1)
	}
	return tsm, nil
}

func (tsm *TaskStreamManager) Initialize(ctx context.Context) error {
	tsm.logger.Info(ctx, "Initializing task streams...", observability.String("component", "task_stream_manager"))

	ctx, cancel := context.WithTimeout(context.Background(), config.GetInitializationTimeout())
	defer cancel()

	// Initialize task streams with specific expiration rules
	streamConfigs := map[string]time.Duration{
		StreamTaskDispatched: TasksProcessingTTL,
		StreamTaskExecuted:   TasksExecutedTTL,
		StreamTaskValidated:  TasksValidatedTTL,
		StreamTaskFailed:     TasksFailedTTL,
		StreamTaskRetry:      TasksRetryTTL,
	}

	for stream, ttl := range streamConfigs {
		tsm.logger.Debug(ctx, "Creating stream", observability.String("stream", stream), observability.Duration("ttl", ttl))
		if err := tsm.redisClient.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
			tsm.logger.Error(ctx, "Failed to initialize stream",
				observability.String("stream", stream),
				observability.Error(err),
				observability.Duration("ttl", ttl))
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		tsm.logger.Debug(ctx, "Stream initialized successfully", observability.String("stream", stream), observability.Duration("ttl", ttl))
	}

	// Register consumer groups for task processing
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskDispatched, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for executed tasks (pending validation)
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskExecuted, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group for executed stream: %w", err)
	}

	// Register consumer groups for validated tasks
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskValidated, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group for validated stream: %w", err)
	}

	// Register consumer groups for task failure
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskFailed, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for task retry
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskRetry, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	// Register consumer groups for timeout checking on executed stream
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskExecuted, "timeout-checker"); err != nil {
		return fmt.Errorf("failed to register timeout-checker group: %w", err)
	}

	// Register consumer groups for task finding
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskDispatched, "task-finder"); err != nil {
		return fmt.Errorf("failed to register task-finder group: %w", err)
	}

	// Register consumer groups for task finding in executed stream
	if err := tsm.RegisterConsumerGroup(ctx, StreamTaskExecuted, "task-finder"); err != nil {
		return fmt.Errorf("failed to register task-finder group for executed stream: %w", err)
	}

	// go tsm.StartStreamHealthMonitor(ctx)

	tsm.logger.Info(ctx, "All task streams initialized successfully", observability.String("component", "task_stream_manager"))

	return nil
}

// RegisterConsumerGroup registers a consumer group for a stream
func (tsm *TaskStreamManager) RegisterConsumerGroup(ctx context.Context, stream string, group string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		// tsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
		return nil
	}

	tsm.logger.Debug(ctx, "Registering consumer group", observability.String("stream", stream), observability.String("group", group))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tsm.redisClient.CreateConsumerGroup(ctx, stream, group); err != nil {
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
	// tsm.logger.Debug("Getting stream information")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{StreamTaskDispatched, StreamTaskExecuted, StreamTaskValidated, StreamTaskRetry, StreamTaskFailed}

	for _, stream := range streams {
		length, err := tsm.redisClient.XLen(ctx, stream)
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
				metrics.TaskStreamLengths.WithLabelValues("ready").Set(ctx, float64(length))
			}
		case StreamTaskRetry:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("retry").Set(ctx, float64(length))
			}
		case StreamTaskFailed:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("failed").Set(ctx, float64(length))
			}
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

	// tsm.logger.Debug("Stream information retrieved", "info", info)
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

// GetExpirationManager returns the expiration manager
func (tsm *TaskStreamManager) GetExpirationManager() *ExpirationManager {
	return tsm.expirationManager
}

// GetPendingEntriesInfo returns information about pending entries in consumer groups
func (tsm *TaskStreamManager) GetPendingEntriesInfo(ctx context.Context) map[string]interface{} {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pendingInfo := make(map[string]interface{})
	consumerGroups := []string{"task-processors", "timeout-checker", "task-finder"}

	for _, group := range consumerGroups {
		pending, err := tsm.redisClient.XPending(ctx, StreamTaskDispatched, group)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to get pending entries info",
				observability.String("consumer_group", group),
				observability.Error(err))
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
			tsm.logger.Warn(ctx, "High number of pending entries detected",
				observability.String("consumer_group", group),
				observability.Int64("pending_count", pending.Count),
				observability.String("min_id", pending.Lower),
				observability.String("max_id", pending.Higher))
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
			tsm.logger.Warn(ctx, "Failed to get pending entries for cleanup",
				observability.String("consumer_group", group),
				observability.Error(err))
			continue
		}

		cleaned := 0
		for _, entry := range pendingExt {
			// If the entry is older than 1 hour, acknowledge it to prevent PEL growth
			if entry.Idle > time.Hour {
				err := tsm.redisClient.XAck(ctx, StreamTaskDispatched, group, entry.ID)
				if err != nil {
					tsm.logger.Error(ctx, "Failed to acknowledge old pending entry",
						observability.String("consumer_group", group),
						observability.String("message_id", entry.ID),
						observability.Duration("idle_time", entry.Idle),
						observability.Error(err))
				} else {
					cleaned++
					tsm.logger.Debug(ctx, "Cleaned up old pending entry",
						observability.String("consumer_group", group),
						observability.String("message_id", entry.ID),
						observability.Duration("idle_time", entry.Idle))
				}
			}
		}

		if cleaned > 0 {
			tsm.logger.Info(ctx, "Cleaned up pending entries",
				observability.String("consumer_group", group),
				observability.Int("cleaned_count", cleaned))
			totalCleaned += cleaned
		}
	}

	if totalCleaned > 0 {
		tsm.logger.Info(ctx, "Pending entries cleanup completed",
			observability.Int("total_cleaned", totalCleaned))
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

// FindTaskByIDInStream finds a task by ID in a specific stream
func (tsm *TaskStreamManager) FindTaskByIDInStream(ctx context.Context, taskID int64, stream string) (*TaskStreamData, string, error) {
	return tsm.taskIndex.FindTaskByIDInStream(ctx, taskID, stream)
}

// RemoveTaskIndex removes a task from the index
func (tsm *TaskStreamManager) RemoveTaskIndex(ctx context.Context, taskID int64) error {
	return tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
}

// RemoveExecutedTaskTimeout removes a task from executed timeout tracking
func (tsm *TaskStreamManager) RemoveExecutedTaskTimeout(ctx context.Context, taskID int64) error {
	return tsm.expirationManager.RemoveExecutedTaskTimeout(ctx, taskID)
}

// AddTaskToStream adds a task to a specific stream
func (tsm *TaskStreamManager) AddTaskToStream(ctx context.Context, stream string, task *TaskStreamData) error {
	return tsm.addTaskToStream(ctx, stream, task)
}

func (tsm *TaskStreamManager) Close(ctx context.Context) error {
	tsm.logger.Info(ctx, "Closing TaskStreamManager")

	err := tsm.redisClient.Close()
	if err != nil {
		tsm.logger.Error(ctx, "Failed to close Redis client", observability.Error(err))
		return err
	}

	tsm.logger.Info(ctx, "TaskStreamManager closed successfully")
	return nil
}

// startStreamHealthMonitor monitors the health of Redis streams
func (tsm *TaskStreamManager) StartStreamHealthMonitor(ctx context.Context) {
	tsm.logger.Info(ctx, "Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second)          // Check health every 30 seconds
	cleanupTicker := time.NewTicker(5 * time.Minute)    // Cleanup every 5 minutes
	trimTicker := time.NewTicker(10 * time.Minute)      // Trim streams every 10 minutes
	expirationTicker := time.NewTicker(1 * time.Minute) // Check for expired entries every minute
	defer ticker.Stop()
	defer cleanupTicker.Stop()
	defer trimTicker.Stop()
	defer expirationTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info(ctx, "Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			taskInfo := tsm.GetStreamInfo(ctx)

			// Log warnings for high stream lengths
			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 { // Warn if more than 50 tasks in any stream
						tsm.logger.Warn(ctx, "High task stream length detected",
							observability.String("stream", stream),
							observability.Int64("length", length))
					}
				}
			}

			// Check pending entries
			pendingInfo := tsm.GetPendingEntriesInfo(ctx)
			for group, info := range pendingInfo {
				if infoMap, ok := info.(map[string]interface{}); ok {
					if count, exists := infoMap["count"]; exists {
						if countInt, ok := count.(int64); ok && countInt > 50 {
							tsm.logger.Warn(ctx, "High number of pending entries detected",
								observability.String("consumer_group", group),
								observability.Int64("pending_count", countInt))
						}
					}
				}
			}
		case <-cleanupTicker.C:
			// Periodic cleanup of old pending entries
			if err := tsm.CleanupPendingEntries(ctx); err != nil {
				tsm.logger.Error(ctx, "Failed to cleanup pending entries", observability.Error(err))
			}
		case <-trimTicker.C:
			// Periodic trimming of streams to remove old messages
			if err := tsm.TrimStreams(ctx); err != nil {
				tsm.logger.Error(ctx, "Failed to trim streams", observability.Error(err))
			}
		case <-expirationTicker.C:
			// Periodic cleanup of expired stream entries
			if err := tsm.CleanupExpiredStreamEntries(ctx); err != nil {
				tsm.logger.Error(ctx, "Failed to cleanup expired stream entries", observability.Error(err))
			}
		}
	}
}

// TrimStreams periodically trims old messages from streams to prevent unbounded growth
func (tsm *TaskStreamManager) TrimStreams(ctx context.Context) error {
	tsm.logger.Debug(ctx, "Starting periodic stream trimming")

	// Define streams and their max lengths
	streamConfigs := map[string]int64{
		StreamTaskDispatched: 10000, // Keep max 10000 messages in dispatched stream
		StreamTaskFailed:     5000,  // Keep max 5000 messages in failed stream
		StreamTaskRetry:      5000,  // Keep max 5000 messages in retry stream
	}

	totalTrimmed := int64(0)
	for stream, maxLen := range streamConfigs {
		trimmed, err := tsm.redisClient.XTrim(ctx, stream, maxLen, true) // Use approximate trimming for better performance
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to trim stream",
				observability.String("stream", stream),
				observability.Int64("max_len", maxLen),
				observability.Error(err))
			continue
		}

		if trimmed > 0 {
			totalTrimmed += trimmed
			tsm.logger.Info(ctx, "Trimmed old messages from stream",
				observability.String("stream", stream),
				observability.Int64("messages_trimmed", trimmed),
				observability.Int64("max_len", maxLen))
		}
	}

	if totalTrimmed > 0 {
		tsm.logger.Info(ctx, "Periodic stream trimming completed",
			observability.Int64("total_messages_trimmed", totalTrimmed))
	}

	return nil
}

// CleanupExpiredStreamEntries periodically removes expired entries from all streams
func (tsm *TaskStreamManager) CleanupExpiredStreamEntries(ctx context.Context) error {
	tsm.logger.Debug(ctx, "Starting periodic cleanup of expired stream entries")

	// Get expired entries for all streams
	expiredEntries, err := tsm.expirationManager.GetExpiredMessagesForAllStreams(ctx)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to get expired stream entries", observability.Error(err))
		return err
	}

	totalDeleted := 0
	for stream, messageIDs := range expiredEntries {
		if len(messageIDs) == 0 {
			continue
		}

		tsm.logger.Info(ctx, "Found expired entries for cleanup",
			observability.String("stream", stream),
			observability.Int("expired_count", len(messageIDs)))

		// Delete expired messages from streams
		for _, messageID := range messageIDs {
			deleted, err := tsm.redisClient.XDel(ctx, stream, messageID)
			if err != nil {
				tsm.logger.Warn(ctx, "Failed to delete expired message from stream",
					observability.String("stream", stream),
					observability.String("message_id", messageID),
					observability.Error(err))
				continue
			}
			if deleted > 0 {
				totalDeleted++
				tsm.logger.Debug(ctx, "Deleted expired stream entry",
					observability.String("stream", stream),
					observability.String("message_id", messageID))
			}
		}

		// Remove from expiration tracking
		err = tsm.expirationManager.RemoveMultipleMessageExpirations(ctx, stream, messageIDs)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to remove expired entries from expiration tracking",
				observability.String("stream", stream),
				observability.Int("count", len(messageIDs)),
				observability.Error(err))
		}
	}

	if totalDeleted > 0 {
		tsm.logger.Info(ctx, "Periodic cleanup of expired stream entries completed",
			observability.Int("total_deleted", totalDeleted))
	}

	return nil
}
