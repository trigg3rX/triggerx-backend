package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

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
}

func NewTaskStreamManager(redisClient redisClient.RedisClientInterface, dbClient *database.DatabaseClient, logger logging.Logger) (*TaskStreamManager, error) {
	tsm := &TaskStreamManager{
		redisClient:    redisClient,
		dbClient:       dbClient,
		logger:         logger,
		consumerGroups: make(map[string]bool),
		startTime:      time.Now(),
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

// FindTaskInDispatched finds a specific task in the dispatched stream
func (tsm *TaskStreamManager) FindTaskInDispatched(taskID int64) (*TaskStreamData, error) {
	tasks, _, err := tsm.ReadTasksFromStream(StreamTaskDispatched, "task-finder", "finder", 100)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found in dispatched stream", taskID)
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
		}
	}
}
