package scheduler

// import (
// 	"context"
// 	"errors"
// 	"net/http"
// 	"testing"
// 	"time"

// 	// "github.com/trigg3rX/triggerx-backend/internal/cache"
// 	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// 	schedulerTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// func init() {
// 	_ = logging.InitServiceLogger(logging.LoggerConfig{
// 		LogDir:          logging.BaseDataDir,
// 		ProcessName:     "test",
// 		Environment:     logging.Development,
// 		UseColors:       false,
// 		MinStdoutLevel:  logging.InfoLevel,
// 		MinFileLogLevel: logging.InfoLevel,
// 	})
// }

// // --- Mock Config ---
// var mockMaxWorkers = 2

// // --- Mock Cache ---
// type mockCache struct {
// 	lockAcquired bool
// 	lockErr      error
// 	getMap       map[string]string
// 	setErr       error
// 	getErr       error
// }

// func (m *mockCache) AcquirePerformerLock(key string, ttl time.Duration) (bool, error) {
// 	return m.lockAcquired, m.lockErr
// }
// func (m *mockCache) ReleasePerformerLock(key string) error { return nil }
// func (m *mockCache) Get(key string) (string, error) {
// 	if m.getErr != nil {
// 		return "", m.getErr
// 	}
// 	v, ok := m.getMap[key]
// 	if !ok {
// 		return "", errors.New("not found")
// 	}
// 	return v, nil
// }
// func (m *mockCache) Set(key, value string, ttl time.Duration) error { return m.setErr }
// func (m *mockCache) Delete(key string) error                        { delete(m.getMap, key); return nil }

// // --- Mock Logger ---
// type mockLogger struct{}

// func (m *mockLogger) Debug(msg string, tags ...any)               {}
// func (m *mockLogger) Info(msg string, tags ...any)                {}
// func (m *mockLogger) Warn(msg string, tags ...any)                {}
// func (m *mockLogger) Error(msg string, tags ...any)               {}
// func (m *mockLogger) Fatal(msg string, tags ...any)               {}
// func (m *mockLogger) Debugf(template string, args ...interface{}) {}
// func (m *mockLogger) Infof(template string, args ...interface{})  {}
// func (m *mockLogger) Warnf(template string, args ...interface{})  {}
// func (m *mockLogger) Errorf(template string, args ...interface{}) {}
// func (m *mockLogger) Fatalf(template string, args ...interface{}) {}
// func (m *mockLogger) With(tags ...any) logging.Logger             { return m }

// --- Helper to create a scheduler ---
// func newTestScheduler(cacheInst cache.Cache, dbClient *client.DBServerClient) *ConditionBasedScheduler {
// 	return &ConditionBasedScheduler{
// 		ctx:        context.Background(),
// 		cancel:     func() {},
// 		logger:     &mockLogger{},
// 		workers:    make(map[int64]*ConditionWorker),
// 		cache:      cacheInst,
// 		dbClient:   dbClient,
// 		managerID:  "test-manager",
// 		httpClient: &http.Client{Timeout: 2 * time.Second},
// 		maxWorkers: mockMaxWorkers,
// 		metrics:    nil,
// 	}
// }

// --- Test ScheduleJob edge cases ---
// func TestScheduleJob_EdgeCases(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{} // Only SendTaskToManager is used, so can use zero struct
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 1,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeAPI,
// 		ValueSourceUrl:        "http://test/api",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 	}

// 	// 1. Normal scheduling
// 	if err := s.ScheduleJob(job); err != nil {
// 		t.Errorf("expected no error, got %v", err)
// 	}
// 	// 2. Duplicate job
// 	if err := s.ScheduleJob(job); err == nil {
// 		t.Errorf("expected error for duplicate job, got nil")
// 	}
// 	// 3. Max workers
// 	job2 := *job
// 	job2.JobID = 2
// 	if err := s.ScheduleJob(&job2); err != nil {
// 		t.Errorf("expected no error, got %v", err)
// 	}
// 	job3 := *job
// 	job3.JobID = 3
// 	if err := s.ScheduleJob(&job3); err == nil {
// 		t.Errorf("expected error for max workers, got nil")
// 	}
// 	// 4. Invalid condition type
// 	job4 := *job
// 	job4.JobID = 4
// 	job4.ConditionType = "invalid"
// 	delete(s.workers, 2) // free up slot
// 	if err := s.ScheduleJob(&job4); err == nil {
// 		t.Errorf("expected error for invalid condition type, got nil")
// 	}
// 	// 5. Invalid value source type
// 	job5 := *job
// 	job5.JobID = 5
// 	job5.ValueSourceType = "invalid"
// 	if err := s.ScheduleJob(&job5); err == nil {
// 		t.Errorf("expected error for invalid value source type, got nil")
// 	}
// }

// --- Test UnscheduleJob edge cases ---
// func TestUnscheduleJob_EdgeCases(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)
// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 10,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeAPI,
// 		ValueSourceUrl:        "http://test/api",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 	}
// 	_ = s.ScheduleJob(job)
// 	// 1. Unschedule existing job
// 	if err := s.UnscheduleJob(10); err != nil {
// 		t.Errorf("expected no error, got %v", err)
// 	}
// 	// 2. Unschedule non-existent job
// 	if err := s.UnscheduleJob(999); err == nil {
// 		t.Errorf("expected error for unscheduling non-existent job, got nil")
// 	}
// }

// --- Worker lock acquisition edge cases ---
// func TestWorker_Start_LockEdgeCases(t *testing.T) {
// 	cache := &mockCache{lockAcquired: false, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 100,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeAPI,
// 		ValueSourceUrl:        "http://test/api",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 	}

// 	worker, err := s.createConditionWorker(job)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	// Should be running even if lock not acquired (matches implementation)
// 	go worker.start()
// 	time.Sleep(100 * time.Millisecond)
// 	if !worker.IsRunning() {
// 		t.Errorf("worker should be running even if lock not acquired (matches implementation)")
// 	}

// 	// Simulate lock error
// 	cache.lockAcquired = false
// 	cache.lockErr = errors.New("lock error")
// 	worker2, _ := s.createConditionWorker(job)
// 	go worker2.start()
// 	time.Sleep(100 * time.Millisecond)
// 	if !worker2.IsRunning() {
// 		t.Errorf("worker should be running even if lock error (matches implementation)")
// 	}
// }

// --- Cache get/set failures ---
// func TestWorker_CacheFailures(t *testing.T) {
// 	cache := &mockCache{
// 		lockAcquired: true,
// 		getMap:       make(map[string]string),
// 		getErr:       errors.New("cache get error"),
// 		setErr:       errors.New("cache set error"),
// 	}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 200,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeStatic,
// 		ValueSourceUrl:        "42",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 	}

// 	worker, _ := s.createConditionWorker(job)
// 	// Should not panic on cache errors
// 	go worker.start()
// 	time.Sleep(200 * time.Millisecond)
// 	worker.stop()
// }

// --- Action execution with client failure ---
// func TestWorker_ActionExecution_ClientFailure(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	s := newTestScheduler(cache, nil)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 300,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeStatic,
// 		ValueSourceUrl:        "100",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 		Recurring:             false,
// 	}

// 	worker, _ := s.createConditionWorker(job)
// 	// Here, we just run the worker and ensure no panic occurs.
// 	go worker.start()
// 	time.Sleep(200 * time.Millisecond)
// 	worker.stop()
// }

// --- Value fetching with invalid static value ---
// func TestWorker_FetchValue_Errors(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	// Invalid static value
// 	job := &schedulerTypes.ConditionJobData{
// 		JobID:                 400,
// 		ConditionType:         ConditionGreaterThan,
// 		ValueSourceType:       SourceTypeStatic,
// 		ValueSourceUrl:        "not-a-number",
// 		TargetChainID:         "1",
// 		TargetContractAddress: "0xabc",
// 		TargetFunction:        "foo",
// 	}
// 	worker, _ := s.createConditionWorker(job)
// 	if _, err := worker.fetchValue(); err == nil {
// 		t.Errorf("expected error for invalid static value")
// 	}
// }

// --- Worker stop and restart ---
// func TestWorker_StopAndRestart(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID: 500, ConditionType: ConditionGreaterThan, ValueSourceType: SourceTypeStatic,
// 		ValueSourceUrl: "10", TargetChainID: "1", TargetContractAddress: "0xabc", TargetFunction: "foo",
// 	}
// 	worker, _ := s.createConditionWorker(job)
// 	go worker.start()
// 	time.Sleep(100 * time.Millisecond)
// 	worker.stop()
// 	time.Sleep(50 * time.Millisecond)
// 	if worker.IsRunning() {
// 		t.Errorf("worker should not be running after stop")
// 	}
// 	// Try to start again
// 	go worker.start()
// 	time.Sleep(100 * time.Millisecond)
// 	worker.stop()
// }

// // --- Condition evaluation error ---
// func TestWorker_ConditionEvaluationError(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID: 600, ConditionType: "unsupported", ValueSourceType: SourceTypeStatic,
// 		ValueSourceUrl: "10", TargetChainID: "1", TargetContractAddress: "0xabc", TargetFunction: "foo",
// 	}
// 	worker, _ := s.createConditionWorker(job)
// 	if _, err := worker.evaluateCondition(10); err == nil {
// 		t.Errorf("expected error for unsupported condition type")
// 	}
// }

// // --- API value fetch HTTP failure ---
// func TestWorker_FetchFromAPI_HTTPFailure(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID: 700, ConditionType: ConditionGreaterThan, ValueSourceType: SourceTypeAPI,
// 		ValueSourceUrl: "http://localhost:9999/doesnotexist",
// 		TargetChainID:  "1", TargetContractAddress: "0xabc", TargetFunction: "foo",
// 	}
// 	worker, _ := s.createConditionWorker(job)
// 	worker.httpClient = &http.Client{Timeout: 100 * time.Millisecond}
// 	if _, err := worker.fetchFromAPI(); err == nil {
// 		t.Errorf("expected error for HTTP failure")
// 	}
// }

// // --- Recurring job stops after many iterations ---
// func TestWorker_RecurringJobStop(t *testing.T) {
// 	cache := &mockCache{lockAcquired: true, getMap: make(map[string]string)}
// 	dbClient := &client.DBServerClient{}
// 	s := newTestScheduler(cache, dbClient)

// 	job := &schedulerTypes.ConditionJobData{
// 		JobID: 800, ConditionType: ConditionGreaterThan, ValueSourceType: SourceTypeStatic,
// 		ValueSourceUrl: "100", TargetChainID: "1", TargetContractAddress: "0xabc", TargetFunction: "foo",
// 		Recurring: true,
// 	}
// 	worker, _ := s.createConditionWorker(job)
// 	go worker.start()
// 	time.Sleep(500 * time.Millisecond)
// 	worker.stop()
// 	if worker.IsRunning() {
// 		t.Errorf("worker should not be running after stop")
// 	}
// }
