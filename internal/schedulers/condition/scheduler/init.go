package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring and event watching
type ConditionBasedScheduler struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	logger                  logging.Logger
	conditionWorkers        map[int64]*worker.ConditionWorker // jobID -> condition worker
	eventWorkers            map[int64]*worker.EventWorker     // jobID -> event worker
	workersMutex            sync.RWMutex
	chainClients            map[string]*ethclient.Client // chainID -> client
	HTTPClient              *retry.HTTPClient
	dbClient                *dbserver.DBServerClient
	aggClient               *aggregator.AggregatorClient
	metrics                 *metrics.Collector
	maxWorkers              int
	schedulerSigningAddress string
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *dbserver.DBServerClient, aggClient *aggregator.AggregatorClient) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &ConditionBasedScheduler{
		ctx:                     ctx,
		cancel:                  cancel,
		logger:                  logger,
		conditionWorkers:        make(map[int64]*worker.ConditionWorker),
		eventWorkers:            make(map[int64]*worker.EventWorker),
		chainClients:            make(map[string]*ethclient.Client),
		dbClient:                dbClient,
		aggClient:               aggClient,
		metrics:                 metrics.NewCollector(),
		maxWorkers:              config.GetMaxWorkers(),
		schedulerSigningAddress: config.GetSchedulerSigningAddress(),
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

	// Set up health checker for database monitoring
	metrics.SetHealthChecker(dbClient)

	scheduler.logger.Info("Condition-based scheduler initialized",
		"max_workers", scheduler.maxWorkers,
		"scheduler_signing_address", scheduler.schedulerSigningAddress,
		"connected_chains", len(scheduler.chainClients),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Condition-based scheduler ready for job scheduling", "scheduler_signing_address", s.schedulerSigningAddress)

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
	s.conditionWorkers = make(map[int64]*worker.ConditionWorker)
	s.eventWorkers = make(map[int64]*worker.EventWorker)
	s.workersMutex.Unlock()

	// Close chain clients
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info("Closed chain client", "chain_id", chainID)
	}
	s.chainClients = make(map[string]*ethclient.Client)

	duration := time.Since(startTime)

	s.logger.Info("Condition-based scheduler stopped",
		"duration", duration,
		"total_condition_workers_stopped", totalConditionWorkers,
		"total_event_workers_stopped", totalEventWorkers,
		"chains_disconnected", connectedChains,
	)
}
