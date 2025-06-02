package scheduler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
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
	cache        cache.Cache
	metrics      *metrics.Collector
	managerID    string
	httpClient   *http.Client
	maxWorkers   int
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	// Initialize cache with enhanced Redis support
	if err := cache.InitWithLogger(logger); err != nil {
		logger.Warnf("Failed to initialize cache: %v", err)
	}

	cacheInstance, err := cache.GetCache()
	if err != nil {
		logger.Warnf("Cache not available, running without cache: %v", err)
		cacheInstance = nil
	} else {
		// Log cache type and Redis availability
		cacheInfo := cache.GetCacheInfo()
		logger.Infof("Cache initialized: type=%s, redis_available=%v",
			cacheInfo["type"], cacheInfo["redis_available"])
	}

	scheduler := &ConditionBasedScheduler{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		workers:   make(map[int64]*worker.ConditionWorker),
		dbClient:  dbClient,
		cache:     cacheInstance,
		metrics:   metrics.NewCollector(),
		managerID: managerID,
		httpClient: &http.Client{
			Timeout: schedulerTypes.RequestTimeout,
		},
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Add scheduler startup event to Redis stream (Redis is already initialized in main.go)
	if redisx.IsAvailable() {
		startupEvent := map[string]interface{}{
			"event_type":      "scheduler_startup",
			"manager_id":      managerID,
			"max_workers":     maxWorkers,
			"cache_available": cacheInstance != nil,
			"redis_available": redisx.IsAvailable(),
			"poll_interval":   schedulerTypes.PollInterval.String(),
			"request_timeout": schedulerTypes.RequestTimeout.String(),
			"started_at":      time.Now().Unix(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, startupEvent); err != nil {
			logger.Warnf("Failed to add scheduler startup event to Redis stream: %v", err)
		} else {
			logger.Info("Scheduler startup event added to Redis stream")
		}
	}

	scheduler.logger.Info("Condition-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"cache_available", cacheInstance != nil,
		"redis_available", redisx.IsAvailable(),
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
	workerDetails := make([]map[string]interface{}, 0, totalWorkers)

	for jobID, worker := range s.workers {
		isRunning := worker.IsRunning()
		if isRunning {
			runningWorkers++
		}

		workerDetails = append(workerDetails, map[string]interface{}{
			"job_id":            jobID,
			"is_running":        isRunning,
			"condition_type":    worker.Job.ConditionType,
			"value_source_type": worker.Job.ValueSourceType,
			"value_source_url":  worker.Job.ValueSourceUrl,
			"last_value":        worker.LastValue,
			"condition_met":     worker.ConditionMet,
		})
	}
	s.workersMutex.RUnlock()

	// Add comprehensive scheduler shutdown event to Redis stream
	if redisx.IsAvailable() {
		shutdownEvent := map[string]interface{}{
			"event_type":        "scheduler_shutdown",
			"manager_id":        s.managerID,
			"total_workers":     totalWorkers,
			"running_workers":   runningWorkers,
			"cache_available":   s.cache != nil,
			"worker_details":    workerDetails,
			"shutdown_at":       startTime.Unix(),
			"graceful_shutdown": true,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, shutdownEvent); err != nil {
			s.logger.Warnf("Failed to add scheduler shutdown event to Redis stream: %v", err)
		} else {
			s.logger.Info("Scheduler shutdown event added to Redis stream")
		}
	}

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

	// Add final shutdown completion event to Redis stream
	if redisx.IsAvailable() {
		completionEvent := map[string]interface{}{
			"event_type":      "scheduler_shutdown_complete",
			"manager_id":      s.managerID,
			"duration_ms":     duration.Milliseconds(),
			"completed_at":    time.Now().Unix(),
			"workers_stopped": totalWorkers,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, completionEvent); err != nil {
			s.logger.Warnf("Failed to add shutdown completion event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Condition-based scheduler stopped",
		"duration", duration,
		"total_workers_stopped", totalWorkers,
		"running_workers_stopped", runningWorkers,
	)
}