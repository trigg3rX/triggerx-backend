package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskStreamManager manages task streams in Redis for the dispatcher
type TaskStreamManager struct {
	redisClient          *redis.Client
	logger               observability.Logger
	consumerGroups       map[string]bool
	mu                   sync.RWMutex
	startTime            time.Time
	taskRepo             repository.TaskRepository
	aggregatorClient     *aggregator.AggregatorClient
	testAggregatorClient *aggregator.AggregatorClient
}

// NewTaskStreamManager creates a new TaskStreamManager instance
func NewTaskStreamManager(
	ctx context.Context,
	redisClient *redis.Client,
	taskRepo repository.TaskRepository,
	aggClient *aggregator.AggregatorClient,
	testAggregatorClient *aggregator.AggregatorClient,
	logger observability.Logger,
) (*TaskStreamManager, error) {
	logger.Info(ctx, "Initializing TaskStreamManager...")

	tsm := &TaskStreamManager{
		redisClient:          redisClient,
		logger:               logger,
		consumerGroups:       make(map[string]bool),
		startTime:            time.Now(),
		taskRepo:             taskRepo,
		aggregatorClient:     aggClient,
		testAggregatorClient: testAggregatorClient,
	}

	logger.Info(ctx, "TaskStreamManager initialized successfully")
	if metrics.ServiceStatus != nil {
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(ctx, 1)
	}
	return tsm, nil
}

// Initialize initializes all task streams and consumer groups
func (tsm *TaskStreamManager) Initialize(ctx context.Context) error {
	tsm.logger.Info(ctx, "Initializing task streams...")

	// Initialize streams using the redis client
	if err := tsm.redisClient.InitializeStreams(ctx); err != nil {
		return fmt.Errorf("failed to initialize streams: %w", err)
	}

	// Initialize consumer groups using the redis client
	if err := tsm.redisClient.InitializeConsumerGroups(ctx); err != nil {
		return fmt.Errorf("failed to initialize consumer groups: %w", err)
	}

	// Mark all consumer groups as registered in the local tracking map
	tsm.mu.Lock()
	consumerGroupConfigs := map[string][]string{
		types.StreamTaskDispatched: {"task-processors"},
		types.StreamTaskCompleted:  {"task-processors"},
		types.StreamTaskFailed:     {"task-processors"},
		types.StreamTaskRetry:      {"task-processors"},
	}
	for stream, groups := range consumerGroupConfigs {
		for _, group := range groups {
			key := fmt.Sprintf("%s:%s", stream, group)
			tsm.consumerGroups[key] = true
		}
	}
	tsm.mu.Unlock()

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

// Close closes the TaskStreamManager and its resources
func (tsm *TaskStreamManager) Close(ctx context.Context) error {
	tsm.logger.Info(ctx, "Closing TaskStreamManager")

	err := tsm.redisClient.Close()
	if err != nil {
		tsm.logger.Error(ctx, "Failed to close Redis client", observability.Error(err))
		return err
	}

	if tsm.aggregatorClient != nil {
		tsm.aggregatorClient.Close()
	}
	if tsm.testAggregatorClient != nil {
		tsm.testAggregatorClient.Close()
	}

	tsm.logger.Info(ctx, "TaskStreamManager closed successfully")
	return nil
}
