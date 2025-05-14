package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
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
	eventWatchers map[int64]*EventWatcher
	conditions    map[int64]*ConditionMonitor

	stateCache map[int64]interface{}
	cacheMutex sync.RWMutex

	balancer     *LoadBalancer
	cacheManager *CacheManager

	workers      map[int64]workers.Worker
	workerCtx    context.Context
	workerCancel context.CancelFunc

	jobChainStatus map[int64]string
	chainMutex     sync.RWMutex
}

type ConditionMonitor struct {
	ctx           context.Context
	jobID         string
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
}

func (c *ConditionMonitor) Stop() {
}

type EventWatcher struct {
	ctx             context.Context
	jobID           string
	chainID         string
	contractAddress string
	eventName       string
}

func NewEventWatcher(ctx context.Context, jobID string, chainID string, contractAddress string, eventName string) *EventWatcher {
	return &EventWatcher{
		ctx:             ctx,
		jobID:           jobID,
		chainID:         chainID,
		contractAddress: contractAddress,
		eventName:       eventName,
	}
}

func (e *EventWatcher) Start() {
}

func (e *EventWatcher) Stop() {
}

func NewJobScheduler(logger logging.Logger) (*JobScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &JobScheduler{
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		cronScheduler:  cron.New(cron.WithSeconds()),
		eventWatchers:  make(map[int64]*EventWatcher),
		conditions:     make(map[int64]*ConditionMonitor),
		stateCache:     make(map[int64]interface{}),
		balancer:       NewLoadBalancer(),
		workers:        make(map[int64]workers.Worker),
		workerCtx:      ctx,
		workerCancel:   cancel,
		jobChainStatus: make(map[int64]string),
		mu:             sync.RWMutex{},
		cacheMutex:     sync.RWMutex{},
		chainMutex:     sync.RWMutex{},
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

func (s *JobScheduler) StartTimeBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewTimeBasedWorker(jobData, fmt.Sprintf("@every %ds", jobData.TimeInterval), s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	s.cacheMutex.Lock()
	s.stateCache[jobData.JobID] = map[string]interface{}{
		"created":       jobData.CreatedAt,
		"last_executed": jobData.LastExecutedAt,
		"status":        "running",
		"type":          "time-based",
		"timeframe":     jobData.TimeFrame,
		"interval":      jobData.TimeInterval,
	}
	s.cacheMutex.Unlock()

	go worker.Start(s.workerCtx)

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state: %v", err)
	}

	s.logger.Infof("Started time-based job %d", jobData.JobID)
	return nil
}

func (s *JobScheduler) StartEventBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewEventBasedWorker(jobData, s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	s.cacheMutex.Lock()
	s.stateCache[jobData.JobID] = map[string]interface{}{
		"created":       jobData.CreatedAt,
		"last_executed": jobData.LastExecutedAt,
		"status":        "running",
		"type":          "event-based",
		"chain_id":      jobData.TriggerChainID,
		"recurring":     jobData.Recurring,
	}
	s.cacheMutex.Unlock()

	go worker.Start(s.workerCtx)

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state: %v", err)
	}

	s.logger.Infof("Started event-based job %d", jobData.JobID)
	return nil
}

func (s *JobScheduler) StartConditionBasedJob(jobData types.HandleCreateJobData) error {
	if !s.canAcceptNewJob() {
		s.logger.Warnf("System resources exceeded, queueing job %d", jobData.JobID)
		s.balancer.AddJobToQueue(jobData.JobID, 1)
		return nil
	}

	s.mu.Lock()
	worker := workers.NewConditionBasedWorker(jobData, s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	s.cacheMutex.Lock()
	s.stateCache[jobData.JobID] = map[string]interface{}{
		"created":       jobData.CreatedAt,
		"last_executed": jobData.LastExecutedAt,
		"status":        "running",
		"type":          "condition-based",
		"script_url":    jobData.ScriptIPFSUrl,
		"condition":     jobData.TargetFunction,
	}
	s.cacheMutex.Unlock()

	go worker.Start(s.workerCtx)

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state: %v", err)
	}

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

	databaseURL := fmt.Sprintf("%s/%s", config.DatabaseRPCAddress, route)

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
	jobDetails, err := services.GetJobDetails(jobID)
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

func (s *JobScheduler) UpdateJobLastExecutedTime(jobID int64, timestamp time.Time) error {
	s.mu.RLock()
	worker, exists := s.workers[jobID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("worker for job %d not found", jobID)
	}

	if timeWorker, ok := worker.(*workers.TimeBasedWorker); ok {
		timeWorker.UpdateLastExecutedTime(timestamp)
		s.logger.Infof("Updated last executed time for time-based job %d to %v", jobID, timestamp)
		return nil
	}

	if eventWorker, ok := worker.(*workers.EventBasedWorker); ok {
		eventWorker.UpdateLastExecutedTime(timestamp)
		s.logger.Infof("Updated last executed time for event-based job %d to %v", jobID, timestamp)
		return nil
	}

	if conditionWorker, ok := worker.(*workers.ConditionBasedWorker); ok {
		conditionWorker.UpdateLastExecutedTime(timestamp)
		s.logger.Infof("Updated last executed time for condition-based job %d to %v", jobID, timestamp)
		return nil
	}

	return fmt.Errorf("unsupported worker type for job %d", jobID)
}

func (s *JobScheduler) UpdateJobStateCache(jobID int64, field string, value interface{}) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	jobState, exists := s.stateCache[jobID]
	if !exists {
		return fmt.Errorf("no state cache for job %d", jobID)
	}

	if jobStateMap, ok := jobState.(map[string]interface{}); ok {
		jobStateMap[field] = value
		s.stateCache[jobID] = jobStateMap

		if err := s.cacheManager.SaveState(); err != nil {
			return fmt.Errorf("failed to save state after update: %w", err)
		}

		return nil
	}

	return fmt.Errorf("invalid state cache format for job %d", jobID)
}
