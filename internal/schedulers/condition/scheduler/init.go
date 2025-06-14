package scheduler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring
type ConditionBasedScheduler struct {
	ctx          context.Context
	cancel       context.CancelFunc
	logger       logging.Logger
	workers      map[int64]*worker.ConditionWorker // jobID -> worker
	workersMutex sync.RWMutex
	dbClient     *client.DBServerClient
	metrics      *metrics.Collector
	managerID    string
	httpClient   *http.Client
	maxWorkers   int
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	scheduler := &ConditionBasedScheduler{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		workers:   make(map[int64]*worker.ConditionWorker),
		dbClient:  dbClient,
		metrics:   metrics.NewCollector(),
		managerID: managerID,
		httpClient: &http.Client{
			Timeout: schedulerTypes.RequestTimeout,
		},
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Set up health checker for database monitoring
	metrics.SetHealthChecker(dbClient)

	scheduler.logger.Info("Condition-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"poll_interval", schedulerTypes.PollInterval,
		"request_timeout", schedulerTypes.RequestTimeout,
	)

	return scheduler, nil
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Condition-based scheduler ready for job scheduling", "manager_id", s.managerID)

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
	totalWorkers := len(s.workers)
	runningWorkers := 0
	// workerDetails := make([]map[string]interface{}, 0, totalWorkers)

	// for jobID, worker := range s.workers {
	// 	isRunning := worker.IsRunning()
	// 	if isRunning {
	// 		runningWorkers++
	// 	}

	// 	workerDetails = append(workerDetails, map[string]interface{}{
	// 		"job_id":            jobID,
	// 		"is_running":        isRunning,
	// 		"condition_type":    worker.Job.ConditionType,
	// 		"value_source_type": worker.Job.ValueSourceType,
	// 		"value_source_url":  worker.Job.ValueSourceUrl,
	// 		"last_value":        worker.LastValue,
	// 		"condition_met":     worker.ConditionMet,
	// 	})
	// }
	s.workersMutex.RUnlock()

	s.cancel()

	// Stop all workers
	s.workersMutex.Lock()
	for jobID, worker := range s.workers {
		worker.Stop()
		s.logger.Info("Stopped worker", "job_id", jobID)
	}
	s.workers = make(map[int64]*worker.ConditionWorker)
	s.workersMutex.Unlock()

	duration := time.Since(startTime)

	s.logger.Info("Condition-based scheduler stopped",
		"duration", duration,
		"total_workers_stopped", totalWorkers,
		"running_workers_stopped", runningWorkers,
	)
}
