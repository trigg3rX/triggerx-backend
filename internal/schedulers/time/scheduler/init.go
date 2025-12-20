package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeBasedScheduler struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	logger               observability.Logger
	activeTasks          map[int64]*types.ScheduleTimeTaskData
	dbClient             *dbserver.DBServerClient
	taskDispatcherClient *client.Client // RPC client for task dispatcher
	metrics              *metrics.Collector
	schedulerID          int
	pollingInterval      time.Duration
	pollingLookAhead     time.Duration
	taskBatchSize        int
	performerLockTTL     time.Duration
	taskCacheTTL         time.Duration
	duplicateTaskWindow  time.Duration
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger observability.Logger, dbClient *dbserver.DBServerClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize RPC client for task dispatcher
	taskDispatcherClient := client.NewClient(client.Config{
		ServiceName: config.GetTaskDispatcherRPCUrl(),
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger)

	scheduler := &TimeBasedScheduler{
		ctx:                  ctx,
		cancel:               cancel,
		logger:               logger,
		activeTasks:          make(map[int64]*types.ScheduleTimeTaskData),
		dbClient:             dbClient,
		taskDispatcherClient: taskDispatcherClient,
		metrics:              metrics.NewCollector(),
		schedulerID:          config.GetSchedulerID(),
		pollingInterval:      config.GetPollingInterval(),
		pollingLookAhead:     config.GetPollingLookAhead(),
		taskBatchSize:        config.GetTaskBatchSize(),
		performerLockTTL:     config.GetPerformerLockTTL(),
		taskCacheTTL:         config.GetTaskCacheTTL(),
		duplicateTaskWindow:  config.GetDuplicateTaskWindow(),
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info(ctx, "Time-based scheduler initialized",
		observability.Int("scheduler_id", scheduler.schedulerID),
		observability.String("task_dispatcher_url", config.GetTaskDispatcherRPCUrl()),
		observability.String("polling_interval", scheduler.pollingInterval.String()),
		observability.String("polling_look_ahead", scheduler.pollingLookAhead.String()),
		observability.Int("task_batch_size", scheduler.taskBatchSize),
		observability.String("performer_lock_ttl", scheduler.performerLockTTL.String()),
		observability.String("task_cache_ttl", scheduler.taskCacheTTL.String()),
		observability.String("duplicate_task_window", scheduler.duplicateTaskWindow.String()),
	)

	return scheduler, nil
}

// Start begins the scheduler's main polling and execution loop
func (s *TimeBasedScheduler) Start(ctx context.Context) {
	s.logger.Info(ctx, "Starting time-based scheduler", observability.Int("scheduler_id", s.schedulerID))

	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()
	// Poll and schedule tasks immediately on startup
	s.pollAndScheduleTasks(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info(ctx, "Scheduler context cancelled, stopping")
			return
		case <-s.ctx.Done():
			s.logger.Info(ctx, "Scheduler stopped")
			return
		case <-ticker.C:
			s.pollAndScheduleTasks(ctx)
		}
	}
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop(ctx context.Context) {
	startTime := time.Now()
	s.logger.Info(ctx, "Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeTasksCount := len(s.activeTasks)

	s.cancel()

	duration := time.Since(startTime)

	s.logger.Info(ctx, "Time-based scheduler stopped",
		observability.Duration("duration", duration),
		observability.Int("active_tasks_stopped", activeTasksCount),
		observability.String("performer_lock_ttl", s.performerLockTTL.String()),
		observability.String("task_cache_ttl", s.taskCacheTTL.String()),
		observability.String("duplicate_task_window", s.duplicateTaskWindow.String()),
	)
}
