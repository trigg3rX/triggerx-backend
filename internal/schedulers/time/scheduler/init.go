package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	activeTasks map[int64]*types.ScheduleTimeTaskData
	taskQueue   chan *types.ScheduleTimeTaskData
	dbClient   *client.DBServerClient
	aggClient  *aggregator.AggregatorClient
	metrics    *metrics.Collector
	schedulerSigningAddress  string
	pollingInterval time.Duration
	pollingLookAhead time.Duration
	jobBatchSize int
	performerLockTTL time.Duration
	taskCacheTTL time.Duration
	duplicateTaskWindow time.Duration
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient, aggClient *aggregator.AggregatorClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &TimeBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		activeTasks: make(map[int64]*types.ScheduleTimeTaskData),
		taskQueue:   make(chan *types.ScheduleTimeTaskData, 100),
		dbClient:   dbClient,
		aggClient:  aggClient,
		metrics:    metrics.NewCollector(),
		schedulerSigningAddress: config.GetSchedulerSigningAddress(),
		pollingInterval: config.GetPollingInterval(),
		pollingLookAhead: config.GetPollingLookAhead(),
		jobBatchSize: config.GetJobBatchSize(),
		performerLockTTL: config.GetPerformerLockTTL(),
		taskCacheTTL: config.GetTaskCacheTTL(),
		duplicateTaskWindow: config.GetDuplicateTaskWindow(),
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Time-based scheduler initialized",
		"scheduler_signing_address", scheduler.schedulerSigningAddress,
		"polling_interval", scheduler.pollingInterval,
		"polling_look_ahead", scheduler.pollingLookAhead,
		"performer_lock_ttl", scheduler.performerLockTTL,
		"task_cache_ttl", scheduler.taskCacheTTL,
		"duplicate_task_window", scheduler.duplicateTaskWindow,
	)

	return scheduler, nil
}
