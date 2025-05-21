package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	// "time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	// "github.com/trigg3rX/triggerx-backend/internal/manager/cache"
	"github.com/trigg3rX/triggerx-backend/internal/manager/client/database"
	"github.com/trigg3rX/triggerx-backend/internal/manager/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/workers"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobScheduler struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	logger logging.Logger

	cronScheduler *cron.Cron

	// cache    cache.Cache
	// balancer *LoadBalancer

	workers      map[int64]workers.Worker
	workerCtx    context.Context
	workerCancel context.CancelFunc

	jobChainStatus map[int64]string
	chainMutex     sync.RWMutex

	dbClient         *database.DatabaseClient
	aggregatorClient *aggregator.AggregatorClient
}

// NewJobScheduler creates a new instance of JobScheduler
// func NewJobScheduler(logger logging.Logger, jobCache cache.Cache, dbClient *database.DatabaseClient, aggregatorClient *aggregator.AggregatorClient) (*JobScheduler, error) {
func NewJobScheduler(logger logging.Logger, dbClient *database.DatabaseClient, aggregatorClient *aggregator.AggregatorClient) (*JobScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &JobScheduler{
		ctx:              ctx,
		cancel:           cancel,
		logger:           logger,
		cronScheduler:    cron.New(cron.WithSeconds()),
		// cache:            jobCache,
		// balancer:         NewLoadBalancer(),
		workers:          make(map[int64]workers.Worker),
		workerCtx:        ctx,
		workerCancel:     cancel,
		jobChainStatus:   make(map[int64]string),
		mu:               sync.RWMutex{},
		chainMutex:       sync.RWMutex{},
		dbClient:         dbClient,
		aggregatorClient: aggregatorClient,
	}

	// go scheduler.balancer.MonitorResources()
	scheduler.cronScheduler.Start()

	return scheduler, nil
}

func (s *JobScheduler) canAcceptNewJob() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// return s.balancer.CheckResourceAvailability()
	return true
}

func (s *JobScheduler) StartTimeBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		// s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewTimeBasedWorker(jobData, fmt.Sprintf("@every %ds", jobData.TimeInterval), s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	// state := &cache.JobState{
	// 	Created:      jobData.CreatedAt,
	// 	LastExecuted: jobData.LastExecutedAt,
	// 	Status:       "running",
	// 	Type:         "time-based",
	// 	Metadata: map[string]interface{}{
	// 		"timeframe": jobData.TimeFrame,
	// 		"interval":  jobData.TimeInterval,
	// 	},
	// }

	// if err := s.cache.Set(s.ctx, jobData.JobID, state); err != nil {
	// 	s.logger.Errorf("Failed to cache job state: %v", err)
	// }

	go worker.Start(s.workerCtx)

	s.logger.Infof("Started time-based job %d", jobData.JobID)
	return nil
}

func (s *JobScheduler) StartEventBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		// s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewEventBasedWorker(jobData, s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	// state := &cache.JobState{
	// 	Created:      jobData.CreatedAt,
	// 	LastExecuted: jobData.LastExecutedAt,
	// 	Status:       "running",
	// 	Type:         "event-based",
	// 	Metadata: map[string]interface{}{
	// 		"chain_id":  jobData.TriggerChainID,
	// 		"recurring": jobData.Recurring,
	// 	},
	// }

	// if err := s.cache.Set(s.ctx, jobData.JobID, state); err != nil {
	// 	s.logger.Errorf("Failed to cache job state: %v", err)
	// }

	go worker.Start(s.workerCtx)

	s.logger.Infof("Started event-based job %d", jobData.JobID)
	return nil
}

func (s *JobScheduler) StartConditionBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		// s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewConditionBasedWorker(jobData, s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	// state := &cache.JobState{
	// 	Created:      jobData.CreatedAt,
	// 	LastExecuted: jobData.LastExecutedAt,
	// 	Status:       "running",
	// 	Type:         "condition-based",
	// 	Metadata: map[string]interface{}{
	// 		"script_url": jobData.ScriptIPFSUrl,
	// 		"condition":  jobData.TargetFunction,
	// 	},
	// }

	// if err := s.cache.Set(s.ctx, jobData.JobID, state); err != nil {
	// 	s.logger.Errorf("Failed to cache job state: %v", err)
	// }

	go worker.Start(s.workerCtx)

	s.logger.Infof("Started condition-based job %d", jobData.JobID)
	return nil
}

func (s *JobScheduler) UpdateJobChainStatus(jobID int64, status string) {
	s.chainMutex.Lock()
	defer s.chainMutex.Unlock()

	s.jobChainStatus[jobID] = status
}

func (s *JobScheduler) SendDataToDatabase(route string, callType int, data interface{}) (bool, error) {
	err := godotenv.Load()
	if err != nil {
		return false, fmt.Errorf("error loading .env file: %v", err)
	}

	databaseURL := fmt.Sprintf("%s/%s", config.GetDatabaseRPCAddress(), route)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	var resp *http.Response
	if callType == 1 {
		resp, err = http.Post(databaseURL, "application/json", bytes.NewBuffer(jsonData))
	} else if callType == 2 {
		resp, err = http.Get(databaseURL)
	}

	if err != nil {
		return false, fmt.Errorf("error sending request to database: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("database service error (status=%d): %s", resp.StatusCode, string(body))
	}

	return true, nil
}

func (s *JobScheduler) Logger() logging.Logger {
	return s.logger
}

func (s *JobScheduler) GetJobDetails(jobID int64) (*types.HandleCreateJobData, error) {
	jobDetails, err := s.dbClient.GetJobDetails(jobID)
	if err != nil {
		return nil, err
	}
	return &jobDetails, nil
}

func (s *JobScheduler) GetWorker(jobID int64) workers.Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil
	}

	return worker
}

// func (s *JobScheduler) UpdateJobLastExecutedTime(jobID int64, timestamp time.Time) error {
// 	if err := s.cache.Update(s.ctx, jobID, "last_executed", timestamp); err != nil {
// 		return fmt.Errorf("failed to update job last executed time: %w", err)
// 	}
// 	return nil
// }

// func (s *JobScheduler) UpdateJobStateCache(jobID int64, field string, value interface{}) error {
// 	if err := s.cache.Update(s.ctx, jobID, field, value); err != nil {
// 		return fmt.Errorf("failed to update job state cache: %w", err)
// 	}
// 	return nil
// }

// RemoveJob removes a job from the scheduler and cache
func (s *JobScheduler) RemoveJob(jobID int64) {
	s.mu.Lock()
	if worker, exists := s.workers[jobID]; exists {
		worker.Stop()
		delete(s.workers, jobID)
	}
	s.mu.Unlock()

	// if err := s.cache.Delete(s.ctx, jobID); err != nil {
	// 	s.logger.Errorf("failed to remove job from cache: %v", err)
	// }
}

func (s *JobScheduler) GetDatabaseClient() *database.DatabaseClient {
	return s.dbClient
}

func (s *JobScheduler) GetAggregatorClient() *aggregator.AggregatorClient {
	return s.aggregatorClient
}
