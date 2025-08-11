package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

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
	conditionWorkers     map[*big.Int]*worker.ConditionWorker       // jobID -> condition worker
	eventWorkers         map[*big.Int]*worker.EventWorker           // jobID -> event worker
	jobDataStore         map[string]*types.ScheduleConditionJobData // jobID -> job data for trigger notifications
	workersMutex         sync.RWMutex
	notificationMutex    sync.Mutex                   // Protect job data during notification processing
	chainClients         map[string]*ethclient.Client // chainID -> client
	HTTPClient           *httppkg.HTTPClient
	dbClient             *dbserver.DBServerClient
	taskDispatcherClient *rpcclient.Client // RPC client for task dispatcher
	metrics              *metrics.Collector
	maxWorkers           int
	schedulerID          int
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

	scheduler := &ConditionBasedScheduler{
		ctx:                  ctx,
		cancel:               cancel,
		logger:               logger,
		conditionWorkers:     make(map[*big.Int]*worker.ConditionWorker),
		eventWorkers:         make(map[*big.Int]*worker.EventWorker),
		jobDataStore:         make(map[string]*types.ScheduleConditionJobData),
		chainClients:         make(map[string]*ethclient.Client),
		dbClient:             dbClient,
		taskDispatcherClient: taskDispatcherClient,
		metrics:              metrics.NewCollector(),
		maxWorkers:           config.GetMaxWorkers(),
		schedulerID:          config.GetSchedulerID(),
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

	// Stop all workers
	s.workersMutex.Lock()
	for jobID, worker := range s.conditionWorkers {
		worker.Stop()
		s.logger.Info("Stopped condition worker", "job_id", jobID)
	}
	for jobID, worker := range s.eventWorkers {
		worker.Stop()
		s.logger.Info("Stopped event worker", "job_id", jobID)
	}
	s.conditionWorkers = make(map[*big.Int]*worker.ConditionWorker)
	s.eventWorkers = make(map[*big.Int]*worker.EventWorker)
	s.jobDataStore = make(map[string]*types.ScheduleConditionJobData)
	s.workersMutex.Unlock()

	// Close chain clients
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info("Closed chain client", "chain_id", chainID)
	}
	s.chainClients = make(map[string]*ethclient.Client)

	// Close task dispatcher RPC client
	if s.taskDispatcherClient != nil {
		if err := s.taskDispatcherClient.Close(); err != nil {
			s.logger.Error("Failed to close task dispatcher RPC client", "error", err)
		} else {
			s.logger.Info("Closed task dispatcher RPC client")
		}
	}

	duration := time.Since(startTime)

	s.logger.Info("Condition-based scheduler stopped",
		"duration", duration,
		"total_condition_workers_stopped", totalConditionWorkers,
		"total_event_workers_stopped", totalEventWorkers,
		"chains_disconnected", connectedChains,
	)
}
