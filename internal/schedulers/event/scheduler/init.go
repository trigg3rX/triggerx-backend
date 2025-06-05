package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// EventBasedScheduler manages individual job workers
type EventBasedScheduler struct {
	ctx          context.Context
	cancel       context.CancelFunc
	logger       logging.Logger
	workers      map[int64]*worker.EventWorker // jobID -> worker
	workersMutex sync.RWMutex
	chainClients map[string]*ethclient.Client // chainID -> client
	clientsMutex sync.RWMutex
	dbClient     *client.DBServerClient
	metrics      *metrics.Collector
	managerID    string
	maxWorkers   int
}

// NewEventBasedScheduler creates a new instance of EventBasedScheduler
func NewEventBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*EventBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	scheduler := &EventBasedScheduler{
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		workers:      make(map[int64]*worker.EventWorker),
		chainClients: make(map[string]*ethclient.Client),
		dbClient:     dbClient,
		metrics:      metrics.NewCollector(),
		managerID:    managerID,
		maxWorkers:   maxWorkers,
	}

	// Initialize chain clients
	if err := scheduler.initChainClients(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize chain clients: %w", err)
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Event-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"connected_chains", len(scheduler.chainClients),
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *EventBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Event-based scheduler ready for job scheduling", "manager_id", s.managerID)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info("Scheduler context cancelled, stopping all workers")
	s.Stop()
}

// Stop gracefully stops all job workers
func (s *EventBasedScheduler) Stop() {
	startTime := time.Now()

	s.logger.Info("Stopping event-based scheduler")

	// Capture statistics before shutdown
	s.workersMutex.RLock()
	totalWorkers := len(s.workers)
	runningWorkers := 0
	// workerDetails := make([]map[string]interface{}, 0, totalWorkers)

	// for jobID, worker := range s.workers {
	// 	isRunning := worker.IsRunning()
	// 	if isRunning {
	// 		runningWorkers++
	// 	}

		// workerDetails = append(workerDetails, map[string]interface{}{
		// 	"job_id":           jobID,
		// 	"is_running":       isRunning,
		// 	"trigger_chain_id": worker.Job.TriggerChainID,
		// 	"contract_address": worker.Job.TriggerContractAddress,
		// 	"trigger_event":    worker.Job.TriggerEvent,
		// 	"last_block":       worker.LastBlock,
		// })
	// }
	s.workersMutex.RUnlock()

	s.clientsMutex.RLock()
	connectedChains := len(s.chainClients)
	s.clientsMutex.RUnlock()

	s.cancel()

	// Stop all workers
	s.workersMutex.Lock()
	for jobID, worker := range s.workers {
		worker.Stop()
		s.logger.Info("Stopped worker", "job_id", jobID)
	}
	s.workers = make(map[int64]*worker.EventWorker)
	s.workersMutex.Unlock()

	// Close chain clients
	s.clientsMutex.Lock()
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info("Closed chain client", "chain_id", chainID)
	}
	s.chainClients = make(map[string]*ethclient.Client)
	s.clientsMutex.Unlock()

	duration := time.Since(startTime)

	s.logger.Info("Event-based scheduler stopped",
		"total_workers_stopped", totalWorkers,
		"running_workers_stopped", runningWorkers,
		"chains_disconnected", connectedChains,
		"duration", duration,
	)
}