package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client/eventmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/repository"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring and event watching
type ConditionBasedScheduler struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	logger               observability.Logger
	tracer               observability.Tracer
	conditionWorkers     map[string]*worker.ConditionWorker         // jobID -> condition worker
	jobDataStore         map[string]*types.ScheduleConditionJobData // jobID -> job data for trigger notifications
	lastTriggerTime      map[string]time.Time                       // jobID -> last trigger timestamp for cooldown
	workersMutex         sync.RWMutex
	notificationMutex    sync.Mutex                        // Protect job data during notification processing
	chainClients         map[string]*nodeclient.NodeClient // chainID -> client
	HTTPClient           *httppkg.HTTPClient
	taskRepository       repository.TaskRepository
	taskDispatcherClient *rpcclient.Client        // RPC client for task dispatcher
	eventMonitorClient   *eventmonitor.GRPCClient // Event Monitor Service gRPC client
	metrics              *metrics.Collector
	maxWorkers           int
	schedulerID          string
	webhookURL           string        // Webhook URL for receiving event notifications
	cooldownPeriod       time.Duration // Cooldown period between task creations for recurring jobs
	jobStatusChecker     *JobStatusChecker
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(logger observability.Logger, tracer observability.Tracer, obsMetrics observability.Metrics, taskRepo repository.TaskRepository, eventJobRepo repository.EventJobRepository, conditionJobRepo repository.ConditionJobRepository) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize RPC client for task dispatcher
	taskDispatcherClient := rpcclient.NewClient(rpcclient.Config{
		ServiceName: config.GetTaskDispatcherRPCUrl(),
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger, tracer)

	// Initialize Event Monitor Service gRPC client
	eventMonitorURL := config.GetEventMonitorRPCUrl()
	eventMonitorClient, err := eventmonitor.NewGRPCClient(eventMonitorURL, logger, tracer)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize Event Monitor Service gRPC client: %w", err)
	}

	// Initialize HTTP client for condition workers
	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize HTTP client: %w", err)
	}

	// Build gRPC service URL for receiving event notifications
	// Format: host:port (e.g., localhost:9016)
	webhookURL := fmt.Sprintf("localhost:%s", config.GetGRPCPort())

	scheduler := &ConditionBasedScheduler{
		ctx:                  ctx,
		cancel:               cancel,
		logger:               logger,
		tracer:               tracer,
		conditionWorkers:     make(map[string]*worker.ConditionWorker),
		jobDataStore:         make(map[string]*types.ScheduleConditionJobData),
		lastTriggerTime:      make(map[string]time.Time),
		chainClients:         make(map[string]*nodeclient.NodeClient),
		HTTPClient:           httpClient,
		taskRepository:       taskRepo,
		taskDispatcherClient: taskDispatcherClient,
		eventMonitorClient:   eventMonitorClient,
		metrics:              metrics.NewCollector(obsMetrics),
		maxWorkers:           config.GetMaxWorkers(),
		schedulerID:          config.GetSchedulerID(),
		webhookURL:           webhookURL,
		cooldownPeriod:       30 * time.Second, // Default 30 seconds cooldown between task creations
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Initialize job status checker with scheduler reference for stopping workers/unregistering events
	scheduler.jobStatusChecker = NewJobStatusChecker(eventJobRepo, conditionJobRepo, scheduler, logger)

	scheduler.logger.Info(ctx, "Condition-based scheduler initialized",
		observability.Int("max_workers", scheduler.maxWorkers),
		observability.String("scheduler_id", scheduler.schedulerID),
		observability.String("task_dispatcher_url", config.GetTaskDispatcherRPCUrl()),
		observability.Int("connected_chains", len(scheduler.chainClients)),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info(ctx, "Condition-based scheduler ready for job scheduling",
		observability.String("scheduler_id", s.schedulerID),
	)

	// Start background cleanup goroutine for expired event jobs
	go s.cleanupExpiredEventJobs(ctx)

	// Start job status checker in background
	go s.jobStatusChecker.StartStatusCheckLoop(ctx)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info(ctx, "Scheduler context cancelled, stopping all workers")
	s.Stop(ctx)
}

// Stop gracefully stops all condition workers
func (s *ConditionBasedScheduler) Stop(ctx context.Context) {
	s.cancel()

	// Stop all workers and unregister event jobs
	s.workersMutex.Lock()
	for jobID, worker := range s.conditionWorkers {
		worker.Stop(ctx)
		s.logger.Info(ctx, "Stopped condition worker", observability.String("job_id", jobID))
	}

	// Unregister all event jobs from Event Monitor Service
	s.conditionWorkers = make(map[string]*worker.ConditionWorker)
	s.jobDataStore = make(map[string]*types.ScheduleConditionJobData)
	s.workersMutex.Unlock()

	// Close chain clients
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info(ctx, "Closed chain client", observability.String("chain_id", chainID))
	}
	s.chainClients = make(map[string]*nodeclient.NodeClient)

	// Close task dispatcher RPC client
	if s.taskDispatcherClient != nil {
		if err := s.taskDispatcherClient.Close(ctx); err != nil {
			s.logger.Error(ctx, "Failed to close task dispatcher RPC client", observability.Error(err))
		} else {
			s.logger.Info(ctx, "Closed task dispatcher RPC client")
		}
	}

	// Close Event Monitor Service gRPC client
	if s.eventMonitorClient != nil {
		s.eventMonitorClient.Close()
		s.logger.Info(ctx, "Closed Event Monitor Service gRPC client")
	}
}

// GetSchedulerID returns the scheduler ID
func (s *ConditionBasedScheduler) GetSchedulerID() string {
	return s.schedulerID
}

// cleanupExpiredEventJobs periodically checks for expired event jobs and unregisters them
// This handles the case where jobs expire without ever triggering an event
func (s *ConditionBasedScheduler) cleanupExpiredEventJobs(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info(ctx, "Stopping expired event jobs cleanup")
			return
		case <-ticker.C:
			now := time.Now()
			// Store expired job data
			expiredJobData := make([]*types.ScheduleConditionJobData, 0)

			// Find expired event jobs from jobDataStore (event jobs are stored there, not in conditionWorkers)
			s.workersMutex.RLock()
			for _, jobData := range s.jobDataStore {
				// Only process event jobs (task definition ID 3 or 4)
				if (jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4) &&
					!jobData.EventWorkerData.ExpirationTime.IsZero() &&
					jobData.EventWorkerData.ExpirationTime.Before(now) {
					expiredJobData = append(expiredJobData, jobData)
				}
			}
			s.workersMutex.RUnlock()

			// Unregister expired jobs
			if len(expiredJobData) > 0 {
				for _, jobData := range expiredJobData {
					if err := s.UnregisterEventJob(ctx, jobData.JobID); err != nil {
						s.logger.Warn(ctx, "Failed to unregister expired event job",
							observability.String("job_id", jobData.JobID),
							observability.Error(err))
					} else {
						s.logger.Debug(ctx, "Unregistered expired event job",
							observability.String("job_id", jobData.JobID))
					}
				}
				s.logger.Info(ctx, "Cleaned up expired event jobs",
					observability.Int("count", len(expiredJobData)))
			}
		}
	}
}
