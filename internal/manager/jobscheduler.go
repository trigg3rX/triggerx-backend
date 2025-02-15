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

type ConditionMonitor struct {
	ctx      context.Context
	jobID    string
	dbClient *database.Connection
}

func NewConditionMonitor(ctx context.Context, jobID string, db *database.Connection) *ConditionMonitor {
	return &ConditionMonitor{
		ctx:      ctx,
		jobID:    jobID,
		dbClient: db,
	}
}

func (c *ConditionMonitor) Start(callback func()) {
	// TODO: Implement condition monitoring logic
}

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

func (s *JobScheduler) canAcceptNewJob() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.balancer.CheckResourceAvailability()
}

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
			"chain_id":      11155420,
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
			"script_url":    jobDetails.ScriptIpfsUrl,
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

func (s *JobScheduler) getJobStatus(jobID int64) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if worker, exists := s.workers[jobID]; exists {
		return map[string]interface{}{
			"status":        worker.GetStatus(),
			"error":         worker.GetError(),
			"current_retry": worker.GetRetries(),
			"max_retries":   3,
		}
	}
	return nil
}

func (s *JobScheduler) updateJobCache(jobID int64) {
	status := s.getJobStatus(jobID)
	if status == nil {
		return
	}

	s.cacheMutex.Lock()
	if jobData, exists := s.stateCache[jobID]; exists {
		data := jobData.(map[string]interface{})
		for k, v := range status {
			data[k] = v
		}
	}
	s.cacheMutex.Unlock()

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state for job %d: %v", jobID, err)
	}
}
