package manager

import (
	"context"
	"fmt"
	// "strconv"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	// "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	dbClient   *database.Connection
	logger     logging.Logger
	p2pNetwork *network.P2PHost

	// Job handling components
	cronScheduler *cron.Cron
	eventWatchers map[int64]context.CancelFunc
	conditions    map[int64]*ConditionMonitor

	// Cache management
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

// NewJobScheduler creates and initializes a new job scheduler
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

	// Initialize cache manager
	cacheManager, err := NewCacheManager(scheduler)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache manager: %v", err)
	}
	scheduler.cacheManager = cacheManager

	// Start the resource monitor
	go scheduler.balancer.MonitorResources()

	// Start the cron scheduler
	scheduler.cronScheduler.Start()

	return scheduler, nil
}

// canAcceptNewJob checks if the system can handle a new job
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

	// Fetch job details from database
	jobDetails, err := s.GetJobDetails(jobID)
	if err != nil {
		s.logger.Errorf("Failed to fetch job details: %v", err)
		return err
	}

	// Convert time interval to cron expression
	interval := jobDetails.TimeInterval
	cronExpr := fmt.Sprintf("@every %ds", interval)

	s.mu.Lock()
	defer s.mu.Unlock()

	worker := NewTimeBasedWorker(jobID, cronExpr, s)
	s.workers[jobID] = worker

	go worker.Start(s.workerCtx)

	s.cacheMutex.Lock()
	s.stateCache[jobID] = map[string]interface{}{
		"created":       jobDetails.CreatedAt,
		"last_executed": jobDetails.LastExecuted,
		"status":        "running",
		"type":          "time-based",
		"time_interval": interval,
	}
	s.cacheMutex.Unlock()

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state: %v", err)
		return err
	}

	return nil
}

func (s *JobScheduler) StartEventBasedJob(jobID int64) error {
	return nil
}

func (s *JobScheduler) StartConditionBasedJob(jobID int64) error {

	return nil
}

func (s *JobScheduler) StartCustomScriptJob(jobID int64) error {
	return nil
}
