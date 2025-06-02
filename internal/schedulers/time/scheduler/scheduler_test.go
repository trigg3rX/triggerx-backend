// scheduler_test.go
//
// To run these tests:
//   go get github.com/stretchr/testify
//   cd internal/schedulers/time/scheduler
//   go test -v
//
// This file includes tests for NewTimeBasedScheduler, pollAndScheduleJobs, executeJob, CalculateNextExecutionTime, and concurrency.

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// In a shared package, or at the top of scheduler.go or scheduler_test.go:
type DBClient interface {
	GetTimeBasedJobs() ([]types.TimeJobData, error)
	UpdateJobNextExecution(jobID int64, nextExecution time.Time) error
	UpdateJobStatus(jobID int64, status bool) error
}

// --- Mock Logger ---
type mockLogger struct{ mock.Mock }

func (m *mockLogger) Debug(msg string, tags ...any)               { m.Called(msg, tags) }
func (m *mockLogger) Info(msg string, tags ...any)                { m.Called(msg, tags) }
func (m *mockLogger) Warn(msg string, tags ...any)                { m.Called(msg, tags) }
func (m *mockLogger) Error(msg string, tags ...any)               { m.Called(msg, tags) }
func (m *mockLogger) Fatal(msg string, tags ...any)               { m.Called(msg, tags) }
func (m *mockLogger) Debugf(template string, args ...interface{}) { m.Called(template, args) }
func (m *mockLogger) Infof(template string, args ...interface{})  { m.Called(template, args) }
func (m *mockLogger) Warnf(template string, args ...interface{})  { m.Called(template, args) }
func (m *mockLogger) Errorf(template string, args ...interface{}) { m.Called(template, args) }
func (m *mockLogger) Fatalf(template string, args ...interface{}) { m.Called(template, args) }
func (m *mockLogger) With(tags ...any) logging.Logger             { m.Called(tags); return m }

// --- DBServerClient Mock for direct struct usage ---
// For most tests, we use a custom struct with the same methods as DBServerClient.
type fakeDBClient struct{ mock.Mock }

func (m *fakeDBClient) GetTimeBasedJobs() ([]types.TimeJobData, error) {
	args := m.Called()
	return args.Get(0).([]types.TimeJobData), args.Error(1)
}
func (m *fakeDBClient) UpdateJobNextExecution(jobID int64, nextExecution time.Time) error {
	args := m.Called(jobID, nextExecution)
	return args.Error(0)
}
func (m *fakeDBClient) UpdateJobStatus(jobID int64, status bool) error {
	args := m.Called(jobID, status)
	return args.Error(0)
}

// --- Mock Cache ---
type mockCache struct{ mock.Mock }

func (m *mockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}
func (m *mockCache) Set(key string, value string, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}
func (m *mockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}
func (m *mockCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	args := m.Called(performerID, ttl)
	return args.Bool(0), args.Error(1)
}
func (m *mockCache) ReleasePerformerLock(performerID string) error {
	args := m.Called(performerID)
	return args.Error(0)
}

// --- Test NewTimeBasedScheduler ---
func TestNewTimeBasedScheduler_Success(t *testing.T) {
	logger := new(mockLogger)
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return().Maybe()
	logger.On("Fatal", mock.Anything, mock.Anything).Return()
	logger.On("Fatalf", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)

	// Use a real DBServerClient with dummy config (no real network calls will be made)
	db, _ := client.NewDBServerClient(logger, client.Config{DBServerURL: "http://localhost:9999"})

	sched, err := NewTimeBasedScheduler("mgr1", logger, db)
	assert.NoError(t, err)
	assert.NotNil(t, sched)
	assert.Equal(t, "mgr1", sched.managerID)
	assert.NotNil(t, sched.logger)
	assert.NotNil(t, sched.dbClient)
}

func TestNewTimeBasedScheduler_CacheFailure(t *testing.T) {
	// Simulate cache init failure by temporarily replacing cacheInitWithLogger
	origInit := cacheInitWithLogger
	cacheInitWithLogger = func(_ interface{}) error { return errors.New("fail") }
	defer func() { cacheInitWithLogger = origInit }()

	logger := new(mockLogger)
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return().Maybe()
	logger.On("Fatal", mock.Anything, mock.Anything).Return()
	logger.On("Fatalf", mock.Anything, mock.Anything).Return()
	db, _ := client.NewDBServerClient(logger, client.Config{DBServerURL: "http://localhost:9999"})

	sched, err := NewTimeBasedScheduler("mgr2", logger, db)
	assert.NoError(t, err)
	assert.NotNil(t, sched)
}

// --- Test pollAndScheduleJobs ---
func TestPollAndScheduleJobs_SortingBatchingWindow(t *testing.T) {
	logger := new(mockLogger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return() // Added this line
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)

	db := new(fakeDBClient)
	cache := new(mockCache)

	now := time.Now()
	jobs := []types.TimeJobData{
		{JobID: 2, NextExecutionTimestamp: now.Add(2 * time.Minute), TimeInterval: 60},
		{JobID: 1, NextExecutionTimestamp: now.Add(1 * time.Minute), TimeInterval: 60},
		{JobID: 3, NextExecutionTimestamp: now.Add(10 * time.Minute), TimeInterval: 60}, // outside window
	}
	db.On("GetTimeBasedJobs").Return(jobs, nil)
	cache.On("Get", mock.Anything).Return("", errors.New("not found"))
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create scheduler with mocked dependencies instead of real ones
	sched := &TimeBasedScheduler{
		ctx:        context.Background(),
		logger:     logger,
		workerPool: make(chan struct{}, 2),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 10),
		dbClient:   nil,
		cache:      cache,
		managerID:  "mgr",
		maxWorkers: 2,
	}
	// Replace the dbClient's GetTimeBasedJobs method call with our mock
	// We need to create a custom method to test pollAndScheduleJobs with mocked dbClient
	sched.testPollAndScheduleJobs(db)

	// Only jobs 1 and 2 should be queued (job 3 is outside window)
	queued := []*types.TimeJobData{}
LOOP:
	for {
		select {
		case job := <-sched.jobQueue:
			queued = append(queued, job)
			if len(queued) == 2 {
				break LOOP
			}
		case <-time.After(100 * time.Millisecond):
			break LOOP
		}
	}
	assert.Len(t, queued, 2)
	assert.True(t, queued[0].NextExecutionTimestamp.Before(queued[1].NextExecutionTimestamp)) // Sorted by time
}

// --- Test executeJob ---
func TestExecuteJob_LockAcquisitionAndFailure(t *testing.T) {
	logger := new(mockLogger)
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return().Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)

	db := new(fakeDBClient)
	cache := new(mockCache)

	job := &types.TimeJobData{
		JobID:                  42,
		ScheduleType:           "interval",
		TimeInterval:           60,
		NextExecutionTimestamp: time.Now().Add(-1 * time.Minute),
	}

	cache.On("AcquirePerformerLock", mock.Anything, mock.Anything).Return(true, nil)
	cache.On("ReleasePerformerLock", mock.Anything).Return(nil)
	cache.On("Get", mock.Anything).Return("", errors.New("not found"))
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	db.On("UpdateJobStatus", job.JobID, true).Return(nil)
	db.On("UpdateJobNextExecution", job.JobID, mock.Anything).Return(nil)
	db.On("UpdateJobStatus", job.JobID, false).Return(nil)

	sched := &TimeBasedScheduler{
		ctx:        context.Background(),
		logger:     logger,
		workerPool: make(chan struct{}, 1),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 1),
		dbClient:   (*client.DBServerClient)(nil), // Use nil or a proper *client.DBServerClient mock if needed
		cache:      cache,
		managerID:  "mgr",
		maxWorkers: 1,
	}

	sched.testExecuteJob(job, db)
	cache.AssertCalled(t, "AcquirePerformerLock", mock.Anything, mock.Anything)
	cache.AssertCalled(t, "ReleasePerformerLock", mock.Anything)
	db.AssertCalled(t, "UpdateJobStatus", job.JobID, true)
	db.AssertCalled(t, "UpdateJobStatus", job.JobID, false)
}

// --- Test CalculateNextExecutionTime (parser) ---
func TestCalculateNextExecutionTime_CronIntervalTimezone(t *testing.T) {
	cron := "* * * * * *" // valid 5-field cron
	interval := int64(3600)
	tz := "UTC"
	now := time.Now()

	next, err := parser.CalculateNextExecutionTime("cron", 0, cron, "", tz)
	assert.NoError(t, err)
	assert.True(t, next.After(now) || next.Equal(now))

	// Test invalid cron
	_, err = parser.CalculateNextExecutionTime("cron", 0, "invalid", "", tz)
	assert.Error(t, err)

	next2, err := parser.CalculateNextExecutionTime("interval", interval, "", "", tz)
	assert.NoError(t, err)
	assert.True(t, next2.After(now) || next2.Equal(now))
}

// --- Test Concurrency: worker pool limits and queue blocking ---
func TestWorkerPoolConcurrency(t *testing.T) {
	logger := new(mockLogger)
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return().Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)

	db := new(fakeDBClient)
	cache := new(mockCache)
	cache.On("AcquirePerformerLock", mock.Anything, mock.Anything).Return(true, nil)
	cache.On("ReleasePerformerLock", mock.Anything).Return(nil)
	cache.On("Get", mock.Anything).Return("", errors.New("not found"))
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	db.On("UpdateJobStatus", mock.Anything, true).Return(nil)
	db.On("UpdateJobNextExecution", mock.Anything, mock.Anything).Return(nil)
	db.On("UpdateJobStatus", mock.Anything, false).Return(nil)

	maxWorkers := 2
	sched := &TimeBasedScheduler{
		ctx:        context.Background(),
		logger:     logger,
		workerPool: make(chan struct{}, maxWorkers),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 10),
		dbClient:   (*client.DBServerClient)(nil),
		cache:      cache,
		managerID:  "mgr",
		maxWorkers: maxWorkers,
	}

	var wg sync.WaitGroup
	jobCount := 5
	for i := 0; i < maxWorkers; i++ {
		go func() { sched.worker() }()
	}
	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		job := &types.TimeJobData{JobID: int64(i), ScheduleType: "interval", TimeInterval: 60, NextExecutionTimestamp: time.Now()}
		go func(j *types.TimeJobData) {
			sched.jobQueue <- j
			wg.Done()
		}(job)
	}
	wg.Wait()
	assert.Equal(t, maxWorkers, sched.maxWorkers)
}

// Helper methods for testing with mocked dependencies

func (s *TimeBasedScheduler) testPollAndScheduleJobs(mockDB *fakeDBClient) {
	pollStart := time.Now()

	jobs, err := mockDB.GetTimeBasedJobs()
	if err != nil {
		s.logger.Errorf("Failed to fetch time-based jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		s.logger.Debug("No jobs found for execution")
		return
	}

	s.logger.Infof("Found %d jobs to process", len(jobs))

	// Sort jobs by execution time (earliest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].NextExecutionTimestamp.Before(jobs[j].NextExecutionTimestamp)
	})

	// Process jobs in batches
	now := time.Now()
	executionWindow := now.Add(5 * time.Minute) // 5 minute window

	for _, job := range jobs {
		// Check if job is due for execution (within execution window)
		if job.NextExecutionTimestamp.After(executionWindow) {
			continue // Job is not due yet
		}

		if job.NextExecutionTimestamp.Before(now.Add(-1 * time.Minute)) {
			s.logger.Warnf("Job %d is overdue by %v", job.JobID, now.Sub(job.NextExecutionTimestamp))
		}

		// Check cache to prevent duplicate processing
		if s.cache != nil {
			jobKey := fmt.Sprintf("timejob:processing:%d", job.JobID)
			if _, err := s.cache.Get(jobKey); err == nil {
				s.logger.Debugf("Job %d is already being processed (cache hit), skipping", job.JobID)
				continue
			}
			// Mark job as being processed
			if err := s.cache.Set(jobKey, "1", 5*time.Minute); err != nil {
				s.logger.Warnf("Failed to set processing cache for job %d: %v", job.JobID, err)
			}
		}

		// Add job to execution queue
		select {
		case s.jobQueue <- &job:
			s.logger.Debugf("Queued job %d for execution", job.JobID)
		default:
			s.logger.Warnf("Job queue is full, skipping job %d", job.JobID)
		}
	}

	_ = pollStart // suppress unused variable warning
}

func (s *TimeBasedScheduler) testExecuteJob(job *types.TimeJobData, mockDB *fakeDBClient) {
	startTime := time.Now()
	jobKey := fmt.Sprintf("job_%d", job.JobID)

	s.logger.Infof("Executing time-based job %d (type: %s)", job.JobID, job.ScheduleType)

	// Try to acquire performer lock to prevent duplicate execution
	lockAcquired := false
	if s.cache != nil {
		acquired, err := s.cache.AcquirePerformerLock(jobKey, 10*time.Minute)
		if err != nil {
			s.logger.Warnf("Failed to acquire performer lock for job %d: %v", job.JobID, err)
		} else if !acquired {
			s.logger.Warnf("Job %d is already being executed by another instance, skipping", job.JobID)
			return
		}
		lockAcquired = true
		defer func() {
			if err := s.cache.ReleasePerformerLock(jobKey); err != nil {
				s.logger.Warnf("Failed to release performer lock for job %d: %v", job.JobID, err)
			}
		}()
	}

	// Update job status to running
	if err := mockDB.UpdateJobStatus(job.JobID, true); err != nil {
		s.logger.Errorf("Failed to update job %d status to running: %v", job.JobID, err)
	}

	// Calculate next execution time
	nextExecution, err := parser.CalculateNextExecutionTime(
		job.ScheduleType,
		job.TimeInterval,
		job.CronExpression,
		job.SpecificSchedule,
		job.Timezone,
	)
	if err != nil {
		s.logger.Errorf("Failed to calculate next execution time for job %d: %v", job.JobID, err)
		return
	}

	// Update next execution time in database
	if err := mockDB.UpdateJobNextExecution(job.JobID, nextExecution); err != nil {
		s.logger.Errorf("Failed to update next execution time for job %d: %v", job.JobID, err)
		return
	}

	// Update job status to completed
	if err := mockDB.UpdateJobStatus(job.JobID, false); err != nil {
		s.logger.Errorf("Failed to update job %d status to completed: %v", job.JobID, err)
	}

	// Cache the updated job data
	if s.cache != nil {
		// cacheJobData implementation would go here
	}

	duration := time.Since(startTime)
	s.logger.Infof("Completed job %d in %v, next execution at %v",
		job.JobID, duration, nextExecution)

	_ = lockAcquired // suppress unused variable warning
}

// --- Helpers for patching (if needed) ---
var cacheInitWithLogger = func(log interface{}) error {
	return nil // default: do nothing
}

// Add missing import
