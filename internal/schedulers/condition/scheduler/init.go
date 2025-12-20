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
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
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
	conditionWorkers     map[*types.BigInt]*worker.ConditionWorker  // jobID -> condition worker
	eventWorkers         map[*types.BigInt]*worker.EventWorker      // jobID -> event worker
	jobDataStore         map[string]*types.ScheduleConditionJobData // jobID -> job data for trigger notifications
	workersMutex         sync.RWMutex
	notificationMutex    sync.Mutex                        // Protect job data during notification processing
	chainClients         map[string]*nodeclient.NodeClient // chainID -> client
	HTTPClient           *httppkg.HTTPClient
	dbClient             *dbserver.DBServerClient
	taskDispatcherClient *rpcclient.Client    // RPC client for task dispatcher
	eventMonitorClient   *eventmonitor.Client // Event Monitor Service client
	metrics              *metrics.Collector
	maxWorkers           int
	schedulerID          int
	webhookURL           string // Webhook URL for receiving event notifications
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger observability.Logger, dbClient *dbserver.DBServerClient) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize RPC client for task dispatcher
	taskDispatcherClient := rpcclient.NewClient(rpcclient.Config{
		ServiceName: config.GetTaskDispatcherRPCUrl(),
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger)

	// Initialize Event Monitor Service client
	eventMonitorClient, err := eventmonitor.NewClient(config.GetEventMonitorServiceURL(), logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize Event Monitor Service client: %w", err)
	}

	// Build webhook URL for receiving event notifications
	webhookURL := fmt.Sprintf("http://localhost:%s/api/v1/events/notify", config.GetSchedulerRPCPort())

	scheduler := &ConditionBasedScheduler{
		ctx:                  ctx,
		cancel:               cancel,
		logger:               logger,
		conditionWorkers:     make(map[*types.BigInt]*worker.ConditionWorker),
		eventWorkers:         make(map[*types.BigInt]*worker.EventWorker),
		jobDataStore:         make(map[string]*types.ScheduleConditionJobData),
		chainClients:         make(map[string]*nodeclient.NodeClient),
		dbClient:             dbClient,
		taskDispatcherClient: taskDispatcherClient,
		eventMonitorClient:   eventMonitorClient,
		metrics:              metrics.NewCollector(),
		maxWorkers:           config.GetMaxWorkers(),
		schedulerID:          config.GetSchedulerID(),
		webhookURL:           webhookURL,
	}

	// Initialize chain clients for event workers
	if err := scheduler.initChainClients(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize chain clients: %w", err)
	}

	if err := scheduler.initRetryClient(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize retry client: %w", err)
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info(ctx, "Condition-based scheduler initialized",
		observability.Int("max_workers", scheduler.maxWorkers),
		observability.Int("scheduler_id", scheduler.schedulerID),
		observability.String("task_dispatcher_url", config.GetTaskDispatcherRPCUrl()),
		observability.Int("connected_chains", len(scheduler.chainClients)),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info(ctx, "Condition-based scheduler ready for job scheduling",
		observability.Int("scheduler_id", s.schedulerID),
	)

	// Start background cleanup goroutine for expired event jobs
	go s.cleanupExpiredEventJobs(ctx)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info(ctx, "Scheduler context cancelled, stopping all workers")
	s.Stop(ctx)
}

// Stop gracefully stops all condition workers
func (s *ConditionBasedScheduler) Stop(ctx context.Context) {
	startTime := time.Now()
	s.logger.Info(ctx, "Stopping condition-based scheduler")

	// Capture statistics before shutdown
	s.workersMutex.RLock()
	totalConditionWorkers := len(s.conditionWorkers)
	totalEventWorkers := len(s.eventWorkers)
	s.workersMutex.RUnlock()

	connectedChains := len(s.chainClients)

	s.cancel()

	// Stop all workers and unregister event jobs
	s.workersMutex.Lock()
	for jobID, worker := range s.conditionWorkers {
		worker.Stop(ctx)
		s.logger.Info(ctx, "Stopped condition worker", observability.String("job_id", jobID.String()))
	}

	// Unregister all event jobs from Event Monitor Service
	for jobID, worker := range s.eventWorkers {
		// If using Event Monitor Service (worker is nil), unregister
		if worker == nil && s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(ctx, jobID.String()); err != nil {
				s.logger.Warn(ctx, "Failed to unregister event job from Event Monitor Service during shutdown",
					observability.String("job_id", jobID.String()),
					observability.Error(err))
			} else {
				s.logger.Info(ctx, "Unregistered event job from Event Monitor Service", observability.String("job_id", jobID.String()))
			}
		} else if worker != nil {
			// Stop local worker if it exists (for backward compatibility)
			worker.Stop(ctx)
			s.logger.Info(ctx, "Stopped event worker", observability.String("job_id", jobID.String()))
		}
	}
	s.conditionWorkers = make(map[*types.BigInt]*worker.ConditionWorker)
	s.eventWorkers = make(map[*types.BigInt]*worker.EventWorker)
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

	// Close Event Monitor Service client
	if s.eventMonitorClient != nil {
		s.eventMonitorClient.Close()
		s.logger.Info(ctx, "Closed Event Monitor Service client")
	}

	duration := time.Since(startTime)

	s.logger.Info(ctx, "Condition-based scheduler stopped",
		observability.Duration("duration", duration),
		observability.Int("total_condition_workers_stopped", totalConditionWorkers),
		observability.Int("total_event_workers_stopped", totalEventWorkers),
		observability.Int("chains_disconnected", connectedChains),
	)
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
			// Store original *types.BigInt pointers to properly key into maps later
			expiredJobIDs := make([]*types.BigInt, 0)

			// Find expired event jobs
			s.workersMutex.RLock()
			for jobIDBigInt, eventWorker := range s.eventWorkers {
				// Only check jobs that are using Event Monitor Service (eventWorker is nil)
				if eventWorker == nil {
					jobIDStr := jobIDBigInt.String()
					jobData, exists := s.jobDataStore[jobIDStr]
					if exists && jobData != nil {
						// Check if job has expired
						if jobData.EventWorkerData.ExpirationTime.Before(now) {
							expiredJobIDs = append(expiredJobIDs, jobIDBigInt)
						}
					}
				}
			}
			s.workersMutex.RUnlock()

			// Unregister expired jobs
			for _, jobID := range expiredJobIDs {
				s.logger.Info(ctx, "Found expired event job, unregistering from Event Monitor Service",
					observability.String("job_id", jobID.String()))

				if err := s.unregisterEventJobByPointer(ctx, jobID); err != nil {
					// Only log as warning since the job may have been already unregistered
					// by another goroutine (e.g., event notification handler)
					s.logger.Warn(ctx, "Could not unregister expired event job (may already be unregistered)",
						observability.String("job_id", jobID.String()),
						observability.Error(err))
				} else {
					s.logger.Info(ctx, "Successfully unregistered expired event job",
						observability.String("job_id", jobID.String()))
				}
			}

			if len(expiredJobIDs) > 0 {
				s.logger.Info(ctx, "Cleaned up expired event jobs",
					observability.Int("count", len(expiredJobIDs)))
			}
		}
	}
}
