// Package scheduler provides a time-based task scheduler that polls the database
// for scheduled tasks and dispatches them to the task dispatcher service.
//
// The scheduler operates in polling cycles, fetching tasks that are due for execution
// within a configurable look-ahead window, creating task records, and submitting
// them in batches to the task dispatcher via RPC.
package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/rpc/clients/taskdispatcher"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// TimeBasedScheduler is the main scheduler implementation that handles
// time-based task scheduling and execution.
type TimeBasedScheduler struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	logger                  observability.Logger
	tracer                  observability.Tracer
	timeJobRepository       repository.TimeJobRepository
	taskRepository          repository.TaskRepository
	scriptStorageRepository repository.ScriptStorageRepository
	taskDispatcherClient    *taskdispatcher.Client // RPC client for task dispatcher
	metrics                 *metrics.Collector
	schedulerID             string
	pollingInterval         time.Duration
	pollingLookAhead        time.Duration
	taskBatchSize           int
	performerLockTTL        time.Duration
	taskCacheTTL            time.Duration
	duplicateTaskWindow     time.Duration
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler.
//
// Parameters:
//   - logger: Logger for observability
//   - tracer: Tracer for distributed tracing
//   - obsMetrics: Metrics collector for observability
//   - timeJobRepo: Repository for time-based jobs (includes both traditional and agent jobs)
//   - scriptStorageRepo: Repository for script storage (for agent jobs)
//   - taskRepo: Repository for task data operations
//   - taskDispatcherClient: RPC client for task dispatcher service
//
// Returns a configured scheduler instance ready to start, or an error if initialization fails.
func NewTimeBasedScheduler(
	logger observability.Logger,
	tracer observability.Tracer,
	obsMetrics observability.Metrics,
	timeJobRepo repository.TimeJobRepository,
	scriptStorageRepo repository.ScriptStorageRepository,
	taskRepo repository.TaskRepository,
	taskDispatcherClient *taskdispatcher.Client,
) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &TimeBasedScheduler{
		ctx:                     ctx,
		cancel:                  cancel,
		logger:                  logger,
		tracer:                  tracer,
		timeJobRepository:       timeJobRepo,
		taskRepository:          taskRepo,
		scriptStorageRepository: scriptStorageRepo,
		taskDispatcherClient:    taskDispatcherClient,
		metrics:                 metrics.NewCollector(obsMetrics),
		schedulerID:             config.GetSchedulerID(),
		pollingInterval:         config.GetPollingInterval(),
		pollingLookAhead:        config.GetPollingLookAhead(),
		taskBatchSize:           config.GetTaskBatchSize(),
		performerLockTTL:        config.GetPerformerLockTTL(),
		taskCacheTTL:            config.GetTaskCacheTTL(),
		duplicateTaskWindow:     config.GetDuplicateTaskWindow(),
	}

	// Start metrics collection
	scheduler.metrics.Start()

	return scheduler, nil
}

// Start begins the scheduler's main polling and execution loop
func (s *TimeBasedScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()
	// Poll and schedule tasks immediately on startup
	s.pollAndScheduleTasks(ctx)

	for {
		select {
		case <-ctx.Done():
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
	// Cancel scheduler context to stop the polling loop
	s.cancel()

	// Close RPC client connection pool
	if err := s.taskDispatcherClient.Close(ctx); err != nil {
		s.logger.Warn(ctx, "Error closing task dispatcher client", observability.Error(err))
	}
}
