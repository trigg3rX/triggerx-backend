package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval       = 30 * time.Second // Poll every 30 seconds
	executionWindow    = 5 * time.Minute  // Look ahead 5 minutes
	batchSize          = 50               // Process jobs in batches
	performerLockTTL   = 10 * time.Minute // Lock duration for job execution
	jobCacheTTL        = 5 * time.Minute  // Cache TTL for job data
	duplicateJobWindow = 1 * time.Minute  // Window to prevent duplicate job execution
)

type TimeBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	workerPool chan struct{}
	activeJobs map[int64]*types.TimeJobData
	jobQueue   chan *types.TimeJobData
	dbClient   *client.DBServerClient
	cache      cache.Cache
	metrics    *metrics.Collector
	managerID  string
	maxWorkers int
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*TimeBasedScheduler, error) {
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

	scheduler := &TimeBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		workerPool: make(chan struct{}, maxWorkers),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 100),
		dbClient:   dbClient,
		cache:      cacheInstance,
		metrics:    metrics.NewCollector(),
		managerID:  managerID,
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Start the worker pool
	for i := 0; i < maxWorkers; i++ {
		go scheduler.worker()
	}

	// Add scheduler startup event to Redis stream (Redis is already initialized in main.go)
	if redisx.IsAvailable() {
		startupEvent := map[string]interface{}{
			"event_type":      "scheduler_startup",
			"manager_id":      managerID,
			"max_workers":     maxWorkers,
			"cache_available": cacheInstance != nil,
			"redis_available": redisx.IsAvailable(),
			"poll_interval":   pollInterval.String(),
			"started_at":      time.Now().Unix(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, startupEvent); err != nil {
			logger.Warnf("Failed to add scheduler startup event to Redis stream: %v", err)
		} else {
			logger.Info("Scheduler startup event added to Redis stream")
		}
	}

	scheduler.logger.Info("Time-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"cache_available", cacheInstance != nil,
		"redis_available", redisx.IsAvailable(),
		"poll_interval", pollInterval,
	)

	return scheduler, nil
}