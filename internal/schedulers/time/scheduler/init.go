package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval       = 30 * time.Second // Poll every 30 seconds
	executionWindow    = 40 * time.Minute // Look ahead 40 minutes
	batchSize          = 100              // Process jobs in batches
	performerLockTTL   = 2 * time.Minute  // Lock duration for job execution
	jobCacheTTL        = 1 * time.Minute  // Cache TTL for job data
	duplicateJobWindow = 1 * time.Minute  // Window to prevent duplicate job execution
)

// DBClient interface for the scheduler
type DBClient interface {
	GetTimeBasedJobs() ([]types.ScheduleTimeJobData, error)
	UpdateJobNextExecution(jobID int64, nextExecution time.Time) error
	UpdateJobStatus(jobID int64, status bool) error
	HealthCheck() error
	Close()
}

type TimeBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	activeJobs map[int64]*types.ScheduleTimeJobData
	jobQueue   chan *types.ScheduleTimeJobData
	dbClient   DBClient
	aggClient  *aggregator.AggregatorClient
	metrics    *metrics.Collector
	managerID  string
	maxWorkers int
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient DBClient, aggClient *aggregator.AggregatorClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	scheduler := &TimeBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		activeJobs: make(map[int64]*types.ScheduleTimeJobData),
		jobQueue:   make(chan *types.ScheduleTimeJobData, 100),
		dbClient:   dbClient,
		aggClient:  aggClient,
		metrics:    metrics.NewCollector(),
		managerID:  managerID,
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Time-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"poll_interval", pollInterval,
	)

	return scheduler, nil
}
