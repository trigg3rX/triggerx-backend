package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval    = 30 * time.Second // Poll every 30 seconds
	executionWindow = 5 * time.Minute  // Look ahead 5 minutes
	workerPoolSize  = 10               // Number of concurrent workers
	batchSize       = 50               // Process jobs in batches
)

type EventBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	workerPool chan struct{}
	activeJobs map[int64]*types.EventJobData
	jobQueue   chan *types.EventJobData
	dbClient   *client.DBServerClient
	metrics    *metrics.Collector
	managerID  string
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewEventBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*EventBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &EventBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		workerPool: make(chan struct{}, workerPoolSize),
		activeJobs: make(map[int64]*types.EventJobData),
		jobQueue:   make(chan *types.EventJobData, 100),
		dbClient:   dbClient,
		metrics:    metrics.NewCollector(),
		managerID:  managerID,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	return scheduler, nil
}

// Start begins the scheduler's main polling and execution loop
func (s *EventBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Starting event-based scheduler", "manager_id", s.managerID)

}

// Stop gracefully stops the scheduler
func (s *EventBasedScheduler) Stop() {
	s.logger.Info("Stopping event-based scheduler")
	s.cancel()

	// Close job queue
	close(s.jobQueue)

	// Wait for workers to finish (with timeout)
	timeout := time.After(30 * time.Second)
	for len(s.workerPool) < workerPoolSize {
		select {
		case <-timeout:
			s.logger.Warn("Timeout waiting for workers to finish")
			return
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	s.logger.Info("Event-based scheduler stopped")
}

// GetStats returns current scheduler statistics
func (s *EventBasedScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"manager_id":   s.managerID,
		"active_jobs":  len(s.activeJobs),
		"queue_length": len(s.jobQueue),
		"worker_pool":  len(s.workerPool),
		"max_workers":  workerPoolSize,
	}
}
