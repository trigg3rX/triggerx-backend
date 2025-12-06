package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client/eventmonitor"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring and event watching
type ConditionBasedScheduler struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	logger               logging.Logger
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
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *dbserver.DBServerClient) (*ConditionBasedScheduler, error) {
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
	if err := scheduler.initChainClients(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize chain clients: %w", err)
	}

	if err := scheduler.initRetryClient(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize retry client: %w", err)
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Condition-based scheduler initialized",
		"max_workers", scheduler.maxWorkers,
		"scheduler_id", scheduler.schedulerID,
		"task_dispatcher_url", config.GetTaskDispatcherRPCUrl(),
		"connected_chains", len(scheduler.chainClients),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Condition-based scheduler ready for job scheduling",
		"scheduler_id", s.schedulerID)

	// Start background cleanup goroutine for expired event jobs
	go s.cleanupExpiredEventJobs(ctx)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info("Scheduler context cancelled, stopping all workers")
	s.Stop()
}

// Stop gracefully stops all condition workers
func (s *ConditionBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping condition-based scheduler")

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
		worker.Stop()
		s.logger.Info("Stopped condition worker", "job_id", jobID)
	}

	// Unregister all event jobs from Event Monitor Service
	for jobID, worker := range s.eventWorkers {
		// If using Event Monitor Service (worker is nil), unregister
		if worker == nil && s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(jobID.String()); err != nil {
				s.logger.Warn("Failed to unregister event job from Event Monitor Service during shutdown",
					"job_id", jobID,
					"error", err)
			} else {
				s.logger.Info("Unregistered event job from Event Monitor Service", "job_id", jobID)
			}
		} else if worker != nil {
			// Stop local worker if it exists (for backward compatibility)
			worker.Stop()
			s.logger.Info("Stopped event worker", "job_id", jobID)
		}
	}
	s.conditionWorkers = make(map[*types.BigInt]*worker.ConditionWorker)
	s.eventWorkers = make(map[*types.BigInt]*worker.EventWorker)
	s.jobDataStore = make(map[string]*types.ScheduleConditionJobData)
	s.workersMutex.Unlock()

	// Close chain clients
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info("Closed chain client", "chain_id", chainID)
	}
	s.chainClients = make(map[string]*nodeclient.NodeClient)

	// Close task dispatcher RPC client
	if s.taskDispatcherClient != nil {
		if err := s.taskDispatcherClient.Close(); err != nil {
			s.logger.Error("Failed to close task dispatcher RPC client", "error", err)
		} else {
			s.logger.Info("Closed task dispatcher RPC client")
		}
	}

	// Close Event Monitor Service client
	if s.eventMonitorClient != nil {
		s.eventMonitorClient.Close()
		s.logger.Info("Closed Event Monitor Service client")
	}

	duration := time.Since(startTime)

	s.logger.Info("Condition-based scheduler stopped",
		"duration", duration,
		"total_condition_workers_stopped", totalConditionWorkers,
		"total_event_workers_stopped", totalEventWorkers,
		"chains_disconnected", connectedChains,
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
			s.logger.Info("Stopping expired event jobs cleanup")
			return
		case <-ticker.C:
			now := time.Now()
			expiredJobIDs := make([]*big.Int, 0)

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
							expiredJobIDs = append(expiredJobIDs, jobIDBigInt.ToBigInt())
						}
					}
				}
			}
			s.workersMutex.RUnlock()

			// Unregister expired jobs
			for _, jobID := range expiredJobIDs {
				s.logger.Info("Found expired event job, unregistering from Event Monitor Service",
					"job_id", jobID)

				if err := s.UnregisterEventJob(jobID); err != nil {
					s.logger.Error("Failed to unregister expired event job",
						"job_id", jobID,
						"error", err)
				} else {
					s.logger.Info("Successfully unregistered expired event job",
						"job_id", jobID)
				}
			}

			if len(expiredJobIDs) > 0 {
				s.logger.Info("Cleaned up expired event jobs",
					"count", len(expiredJobIDs))
			}
		}
	}
}
