package scheduler

import (
	"context"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeBasedScheduler struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	logger              logging.Logger
	activeTasks         map[int64]*types.ScheduleTimeTaskData
	dbClient            *dbserver.DBServerClient
	httpClient          *http.Client
	redisAPIURL         string
	metrics             *metrics.Collector
	schedulerID         int
	pollingInterval     time.Duration
	pollingLookAhead    time.Duration
	taskBatchSize       int
	performerLockTTL    time.Duration
	taskCacheTTL        time.Duration
	duplicateTaskWindow time.Duration
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient *dbserver.DBServerClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize HTTP client for Redis API calls
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	scheduler := &TimeBasedScheduler{
		ctx:                 ctx,
		cancel:              cancel,
		logger:              logger,
		activeTasks:         make(map[int64]*types.ScheduleTimeTaskData),
		dbClient:            dbClient,
		httpClient:          httpClient,
		redisAPIURL:         config.GetRedisRPCUrl(),
		metrics:             metrics.NewCollector(),
		schedulerID:         config.GetSchedulerID(),
		pollingInterval:     config.GetPollingInterval(),
		pollingLookAhead:    config.GetPollingLookAhead(),
		taskBatchSize:       config.GetTaskBatchSize(),
		performerLockTTL:    config.GetPerformerLockTTL(),
		taskCacheTTL:        config.GetTaskCacheTTL(),
		duplicateTaskWindow: config.GetDuplicateTaskWindow(),
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Time-based scheduler initialized",
		"scheduler_id", scheduler.schedulerID,
		"redis_api_url", scheduler.redisAPIURL,
		"polling_interval", scheduler.pollingInterval,
		"polling_look_ahead", scheduler.pollingLookAhead,
		"task_batch_size", scheduler.taskBatchSize,
		"performer_lock_ttl", scheduler.performerLockTTL,
		"task_cache_ttl", scheduler.taskCacheTTL,
		"duplicate_task_window", scheduler.duplicateTaskWindow,
	)

	return scheduler, nil
}

// Start begins the scheduler's main polling and execution loop
func (s *TimeBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Starting time-based scheduler", "scheduler_id", s.schedulerID)

	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()
	// Poll and schedule tasks immediately on startup
	s.pollAndScheduleTasks()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler context cancelled, stopping")
			return
		case <-s.ctx.Done():
			s.logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.pollAndScheduleTasks()
		}
	}
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeTasksCount := len(s.activeTasks)

	s.cancel()

	duration := time.Since(startTime)

	s.logger.Info("Time-based scheduler stopped",
		"duration", duration,
		"active_tasks_stopped", activeTasksCount,
		"performer_lock_ttl", s.performerLockTTL,
		"task_cache_ttl", s.taskCacheTTL,
		"duplicate_task_window", s.duplicateTaskWindow,
	)
}
