package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/client/dbserver"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/retry"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring and event watching
type ConditionBasedScheduler struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	logger                  logging.Logger
	conditionWorkers        map[int64]*worker.ConditionWorker         // jobID -> condition worker
	eventWorkers            map[int64]*worker.EventWorker             // jobID -> event worker
	jobDataStore            map[int64]*types.ScheduleConditionJobData // jobID -> job data for trigger notifications
	workersMutex            sync.RWMutex
	chainClients            map[string]*ethclient.Client // chainID -> client
	HTTPClient              *retry.HTTPClient
	dbClient                *dbserver.DBServerClient
	httpClient              *http.Client // For Redis API calls
	redisAPIURL             string
	metrics                 *metrics.Collector
	maxWorkers              int
	schedulerID             int
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *dbserver.DBServerClient) (*ConditionBasedScheduler, error) {
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

	scheduler := &ConditionBasedScheduler{
		ctx:                     ctx,
		cancel:                  cancel,
		logger:                  logger,
		conditionWorkers:        make(map[int64]*worker.ConditionWorker),
		eventWorkers:            make(map[int64]*worker.EventWorker),
		jobDataStore:            make(map[int64]*types.ScheduleConditionJobData),
		chainClients:            make(map[string]*ethclient.Client),
		dbClient:                dbClient,
		httpClient:              httpClient,
		redisAPIURL:             config.GetRedisRPCUrl(),
		metrics:                 metrics.NewCollector(),
		maxWorkers:              config.GetMaxWorkers(),
		schedulerID:             config.GetSchedulerID(),
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
		"redis_api_url", scheduler.redisAPIURL,
		"connected_chains", len(scheduler.chainClients),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Condition-based scheduler ready for job scheduling",
		"scheduler_id", s.schedulerID,)

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
	s.jobDataStore = make(map[int64]*types.ScheduleConditionJobData)
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
