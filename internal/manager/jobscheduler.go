package manager

import (
	"context"
	"fmt"

	"sync"

	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
)

// JobScheduler orchestrates different types of jobs (time-based, event-based, condition-based)
// and manages their lifecycle, state persistence, and resource allocation
type JobScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	dbClient   *database.Connection
	logger     logging.Logger
	p2pNetwork *network.P2PHost

	cronScheduler *cron.Cron
	eventWatchers map[int64]context.CancelFunc
	conditions    map[int64]*ConditionMonitor

	stateCache map[int64]interface{}
	cacheMutex sync.RWMutex

	balancer     *LoadBalancer
	cacheManager *CacheManager

	workers      map[int64]Worker
	workerCtx    context.Context
	workerCancel context.CancelFunc
}

// ConditionMonitor tracks external conditions for condition-based jobs
// and triggers job execution when conditions are met
type ConditionMonitor struct {
	ctx      context.Context
	jobID    string
	scriptIPFSUrl string
}

func NewConditionMonitor(ctx context.Context, jobID string, scriptIPFSUrl string) *ConditionMonitor {
	return &ConditionMonitor{
		ctx:           ctx,
		jobID:         jobID,
		scriptIPFSUrl: scriptIPFSUrl,
	}
}

func (c *ConditionMonitor) Start(callback func()) {
	// TODO: Implement condition monitoring logic
}

// NewJobScheduler initializes a new scheduler with resource monitoring,
// state persistence, and job management capabilities
func NewJobScheduler(db *database.Connection, logger logging.Logger, p2p *network.P2PHost) (*JobScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &JobScheduler{
		ctx:           ctx,
		cancel:        cancel,
		dbClient:      db,
		logger:        logger,
		p2pNetwork:    p2p,
		cronScheduler: cron.New(cron.WithSeconds()),
		eventWatchers: make(map[int64]context.CancelFunc),
		conditions:    make(map[int64]*ConditionMonitor),
		stateCache:    make(map[int64]interface{}),
		balancer:      NewLoadBalancer(),
		workers:       make(map[int64]Worker),
		workerCtx:     ctx,
		workerCancel:  cancel,
	}

	cacheManager, err := NewCacheManager(scheduler)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache manager: %v", err)
	}
	scheduler.cacheManager = cacheManager

	go scheduler.balancer.MonitorResources()

	scheduler.cronScheduler.Start()

	return scheduler, nil
}

// canAcceptNewJob checks if system has enough resources to handle a new job
func (s *JobScheduler) canAcceptNewJob() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.balancer.CheckResourceAvailability()
}

// StartTimeBasedJob initializes and runs a job that executes on a time interval.
// Jobs that can't be started due to resource constraints are queued.
func (s *JobScheduler) StartTimeBasedJob(jobID int64) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobID)
		s.balancer.AddJobToQueue(jobID, 1)
		return nil
	}

	go func() {
		jobDetails, err := s.GetJobDetails(jobID)
		if err != nil {
			s.logger.Errorf("Failed to fetch job details: %v", err)
			return
		}

		cronExpr := fmt.Sprintf("@every %ds", jobDetails.TimeInterval)

		s.mu.Lock()
		worker := NewTimeBasedWorker(jobDetails, cronExpr, s)
		s.workers[jobID] = worker
		s.mu.Unlock()

		s.cacheMutex.Lock()
		s.stateCache[jobID] = map[string]interface{}{
			"created":       jobDetails.CreatedAt,
			"last_executed": jobDetails.LastExecuted,
			"status":        "running",
			"type":          "time-based",
			"timeframe":     jobDetails.TimeFrame,
			"interval":      jobDetails.TimeInterval,
		}
		s.cacheMutex.Unlock()

		go worker.Start(s.workerCtx)

		if err := s.cacheManager.SaveState(); err != nil {
			s.logger.Errorf("Failed to save state: %v", err)
			return
		}

		s.logger.Infof("Started time-based job %d", jobID)
	}()

	return nil
}

// StartEventBasedJob initializes and runs a job that executes in response to blockchain events.
// Includes state persistence and resource management.
func (s *JobScheduler) StartEventBasedJob(jobID int64) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobID)
		s.balancer.AddJobToQueue(jobID, 1)
		return nil
	}

	go func() {
		jobDetails, err := s.GetJobDetails(jobID)
		if err != nil {
			s.logger.Errorf("Failed to fetch job details: %v", err)
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		worker := NewEventBasedWorker(jobDetails, s)
		s.workers[jobID] = worker

		s.cacheMutex.Lock()
		s.stateCache[jobID] = map[string]interface{}{
			"created":       jobDetails.CreatedAt,
			"last_executed": jobDetails.LastExecuted,
			"status":        "running",
			"type":          "event-based",
			"chain_id":      jobDetails.TriggerChainID,
			"recurring":     jobDetails.Recurring,
		}
		s.cacheMutex.Unlock()

		go worker.Start(s.workerCtx)

		if err := s.cacheManager.SaveState(); err != nil {
			s.logger.Errorf("Failed to save state: %v", err)
			return
		}

		s.logger.Infof("Started event-based job %d", jobID)
	}()

	return nil
}

// StartConditionBasedJob initializes and runs a job that executes when specific conditions are met.
// Conditions are monitored via external scripts or APIs.
func (s *JobScheduler) StartConditionBasedJob(jobID int64) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobID)
		s.balancer.AddJobToQueue(jobID, 1)
		return nil
	}

	go func() {
		jobDetails, err := s.GetJobDetails(jobID)
		if err != nil {
			s.logger.Errorf("Failed to fetch job details: %v", err)
			return
		}

		s.mu.Lock()
		worker := NewConditionBasedWorker(jobDetails, s)
		s.workers[jobID] = worker
		s.mu.Unlock()

		s.cacheMutex.Lock()
		s.stateCache[jobID] = map[string]interface{}{
			"created":       jobDetails.CreatedAt,
			"last_executed": jobDetails.LastExecuted,
			"status":        "running",
			"type":          "condition-based",
			"script_url":    jobDetails.ScriptIPFSUrl,
			"condition":     jobDetails.TargetFunction,
		}
		s.cacheMutex.Unlock()

		go worker.Start(s.workerCtx)

		if err := s.cacheManager.SaveState(); err != nil {
			s.logger.Errorf("Failed to save state: %v", err)
			return
		}

		s.logger.Infof("Started condition-based job %d", jobID)
	}()

	return nil
}
