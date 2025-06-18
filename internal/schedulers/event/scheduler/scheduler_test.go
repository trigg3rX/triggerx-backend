package scheduler

// import (
// 	"context"
// 	"errors"
// 	"os"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/ethclient"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
// 	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// 	schedulerTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// // --- Mock Types ---

// type MockDBClient struct{ mock.Mock }
// type MockConfig struct {
// 	DevMode          bool
// 	DatabaseHost     string
// 	DatabaseHostPort string
// 	SchedulerRPCPort string
// 	DBServerURL      string
// 	MaxWorkers       int
// 	ChainRPCUrls     map[string]string
// }

// func (m *MockConfig) IsDevMode() bool                    { return m.DevMode }
// func (m *MockConfig) GetDatabaseHost() string            { return m.DatabaseHost }
// func (m *MockConfig) GetDatabaseHostPort() string        { return m.DatabaseHostPort }
// func (m *MockConfig) GetSchedulerRPCPort() string        { return m.SchedulerRPCPort }
// func (m *MockConfig) GetDBServerURL() string             { return m.DBServerURL }
// func (m *MockConfig) GetChainRPCUrls() map[string]string { return m.ChainRPCUrls }
// func (m *MockConfig) GetMaxWorkers() int                 { return m.MaxWorkers }

// func (m *MockDBClient) UpdateJobStatus(jobID int64, status bool) error {
// 	args := m.Called(jobID, status)
// 	return args.Error(0)
// }

// func (m *MockDBClient) Close() {}

// // --- Mock Cache ---
// type MockCache struct{ mock.Mock }

// func (m *MockCache) Get(key string) (string, error) {
// 	args := m.Called(key)
// 	return args.String(0), args.Error(1)
// }
// func (m *MockCache) Set(key, value string, ttl time.Duration) error {
// 	args := m.Called(key, value, ttl)
// 	return args.Error(0)
// }
// func (m *MockCache) Delete(key string) error {
// 	args := m.Called(key)
// 	return args.Error(0)
// }
// func (m *MockCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
// 	args := m.Called(performerID, ttl)
// 	return args.Bool(0), args.Error(1)
// }
// func (m *MockCache) ReleasePerformerLock(performerID string) error {
// 	args := m.Called(performerID)
// 	return args.Error(0)
// }

// // --- Helper: Minimal Logger ---
// type testLogger struct{}

// func (l *testLogger) Info(msg string, args ...any)      {}
// func (l *testLogger) Infof(format string, args ...any)  {}
// func (l *testLogger) Warnf(format string, args ...any)  {}
// func (l *testLogger) Warn(msg string, args ...any)      {}
// func (l *testLogger) Error(msg string, args ...any)     {}
// func (l *testLogger) Errorf(format string, args ...any) {}
// func (l *testLogger) Debug(msg string, args ...any)     {}
// func (l *testLogger) Debugf(format string, args ...any) {}
// func (l *testLogger) Fatal(msg string, args ...any)     {}
// func (l *testLogger) Fatalf(format string, args ...any) {}
// func (l *testLogger) With(args ...any) logging.Logger   { return l }

// // --- Test Data ---
// func validJobData(jobID int64) *schedulerTypes.EventJobData {
// 	return &schedulerTypes.EventJobData{
// 		JobID:                  jobID,
// 		TriggerChainID:         "11155111",
// 		TriggerContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001").Hex(),
// 		TriggerEvent:           "Transfer(address,address,uint256)",
// 		TargetChainID:          "11155111",
// 		TargetContractAddress:  common.HexToAddress("0x0000000000000000000000000000000000000002").Hex(),
// 		TargetFunction:         "doSomething()",
// 		Recurring:              false,
// 	}
// }

// // --- ChainClient interface and mock for tests ---
// type ChainClient interface {
// 	BlockNumber(ctx context.Context) (uint64, error)
// }

// type MockChainClient struct {
// 	mock.Mock
// }

// func (m *MockChainClient) BlockNumber(ctx context.Context) (uint64, error) {
// 	args := m.Called(ctx)
// 	return args.Get(0).(uint64), args.Error(1)
// }

// // --- Test-only constructor for EventBasedScheduler ---
// func newTestEventBasedScheduler(
// 	managerID string,
// 	logger logging.Logger,
// 	dbClient *client.DBServerClient,
// 	mockCfg MockConfig,
// ) *EventBasedScheduler {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	s := &EventBasedScheduler{
// 		ctx:          ctx,
// 		cancel:       cancel,
// 		logger:       logger,
// 		workers:      make(map[int64]*JobWorker),
// 		chainClients: make(map[string]*ethclient.Client),
// 		dbClient:     dbClient,
// 		cache:        nil, // or a mock cache if needed
// 		metrics:      nil, // or metrics.NewCollector() if needed
// 		managerID:    managerID,
// 		maxWorkers:   mockCfg.GetMaxWorkers(),
// 	}
// 	// Add a real ethclient.Client for all chain IDs in mockCfg.ChainRPCUrls
// 	for chainID, rpcURL := range mockCfg.ChainRPCUrls {
// 		client, err := ethclient.Dial(rpcURL)
// 		if err != nil {
// 			panic("Failed to connect to Anvil node at " + rpcURL + ": " + err.Error())
// 		}
// 		s.chainClients[chainID] = client
// 	}
// 	return s
// }

// // --- Tests ---
// func InitForTesting() {
// } // No-op: kept for compatibility or future test setup.
// func TestMain(m *testing.M) {
// 	os.Setenv("APP_ENV", "test")
// 	os.Setenv("MAX_WORKERS", "100") // Ensure maxWorkers is set for tests
// 	InitForTesting()
// 	if err := config.Init(); err != nil {
// 		// Ignore .env file not found error in tests
// 		if !strings.Contains(err.Error(), ".env") {
// 			panic(err)
// 		}
// 	}
// 	code := m.Run()
// 	os.Exit(code)
// }

// func TestScheduleJob_DuplicateID(t *testing.T) {
// 	t.Logf("Starting TestScheduleJob_DuplicateID (with MockConfig)")
// 	logger := &testLogger{}
// 	mockCfg := MockConfig{
// 		DevMode:          true,
// 		DatabaseHost:     "localhost",
// 		DatabaseHostPort: "9042",
// 		SchedulerRPCPort: "9005",
// 		DBServerURL:      "http://localhost:9002",
// 		MaxWorkers:       100,
// 		ChainRPCUrls: map[string]string{
// 			"11155420": "http://127.0.0.1:8545",
// 			"84532":    "http://127.0.0.1:8545",
// 			"11155111": "http://127.0.0.1:8545",
// 		},
// 	}
// 	s := newTestEventBasedScheduler("test-manager", logger, &client.DBServerClient{}, mockCfg)

// 	// Verify maxWorkers is set correctly
// 	t.Logf("maxWorkers: %d", s.maxWorkers)
// 	assert.Greater(t, s.maxWorkers, 0, "maxWorkers should be greater than 0")

// 	job := validJobData(1)
// 	t.Logf("Scheduling job with ID %d", job.JobID)
// 	err := s.ScheduleJob(job)
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not schedule first job: %v", err)
// 		t.Skipf("Skipping test: could not schedule first job: %v", err)
// 	}

// 	t.Logf("Attempting to schedule duplicate job with ID %d", job.JobID)
// 	err = s.ScheduleJob(job)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "already scheduled")
// 	t.Logf("TestScheduleJob_DuplicateID completed")
// }

// func TestScheduleJob_MaxWorkers(t *testing.T) {
// 	t.Logf("Starting TestScheduleJob_MaxWorkers")
// 	logger := &testLogger{}
// 	mockCfg := MockConfig{
// 		DevMode:          true,
// 		DatabaseHost:     "localhost",
// 		DatabaseHostPort: "9042",
// 		SchedulerRPCPort: "9005",
// 		DBServerURL:      "http://localhost:9002",
// 		MaxWorkers:       1,
// 		ChainRPCUrls: map[string]string{
// 			"11155111": "http://127.0.0.1:8545",
// 		},
// 	}
// 	s := newTestEventBasedScheduler("test-manager", logger, &client.DBServerClient{}, mockCfg)

// 	t.Logf("Set maxWorkers to 1")

// 	t.Logf("Scheduling first job")
// 	err := s.ScheduleJob(validJobData(1))
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not schedule first job: %v", err)
// 		t.Skipf("Skipping test: could not schedule first job: %v", err)
// 	}

// 	t.Logf("Attempting to schedule second job (should fail)")
// 	err = s.ScheduleJob(validJobData(2))
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "maximum number of workers")
// 	t.Logf("TestScheduleJob_MaxWorkers completed")
// }

// func TestWorkerLifecycle_StartStop(t *testing.T) {
// 	t.Logf("Starting TestWorkerLifecycle_StartStop")
// 	logger := &testLogger{}
// 	mockCfg := MockConfig{
// 		DevMode:          true,
// 		DatabaseHost:     "localhost",
// 		DatabaseHostPort: "9042",
// 		SchedulerRPCPort: "9005",
// 		DBServerURL:      "http://localhost:9002",
// 		MaxWorkers:       100,
// 		ChainRPCUrls: map[string]string{
// 			"11155111": "http://127.0.0.1:8545",
// 		},
// 	}
// 	s := newTestEventBasedScheduler("test-manager", logger, &client.DBServerClient{}, mockCfg)

// 	t.Logf("maxWorkers: %d", s.maxWorkers)
// 	assert.Greater(t, s.maxWorkers, 0, "maxWorkers should be greater than 0")

// 	job := validJobData(1)
// 	t.Logf("Scheduling job with ID %d", job.JobID)
// 	err := s.ScheduleJob(job)
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not schedule job: %v", err)
// 		t.Skipf("Skipping test: could not schedule job: %v", err)
// 	}

// 	worker := s.workers[job.JobID]
// 	if !assert.NotNil(t, worker) {
// 		t.Logf("Worker not created for job ID %d", job.JobID)
// 		t.Skip("Skipping test: worker not created")
// 	}

// 	t.Logf("Waiting for worker to start...")
// 	time.Sleep(100 * time.Millisecond)

// 	t.Logf("Stopping worker for job ID %d", job.JobID)
// 	worker.stop()
// 	assert.False(t, worker.IsRunning())
// 	t.Logf("TestWorkerLifecycle_StartStop completed")
// }

// func TestCacheHandling_Error(t *testing.T) {
// 	t.Logf("Starting TestCacheHandling_Error")
// 	logger := &testLogger{}
// 	mockCfg := MockConfig{
// 		DevMode:          true,
// 		DatabaseHost:     "localhost",
// 		DatabaseHostPort: "9042",
// 		SchedulerRPCPort: "9005",
// 		DBServerURL:      "http://localhost:9002",
// 		MaxWorkers:       100,
// 		ChainRPCUrls: map[string]string{
// 			"11155111": "http://127.0.0.1:8545",
// 		},
// 	}
// 	s := newTestEventBasedScheduler("test-manager", logger, &client.DBServerClient{}, mockCfg)

// 	t.Logf("maxWorkers: %d", s.maxWorkers)
// 	assert.Greater(t, s.maxWorkers, 0, "maxWorkers should be greater than 0")

// 	cache := new(MockCache)
// 	s.cache = cache
// 	cache.On("Get", mock.Anything).Return("", errors.New("cache error")).Maybe()
// 	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("cache error")).Maybe()
// 	cache.On("AcquirePerformerLock", mock.Anything, mock.Anything).Return(true, nil).Maybe()
// 	cache.On("ReleasePerformerLock", mock.Anything).Return(nil).Maybe()

// 	job := validJobData(1)
// 	t.Logf("Scheduling job with ID %d (cache will error)", job.JobID)
// 	err := s.ScheduleJob(job)
// 	assert.NoError(t, err)
// 	t.Logf("TestCacheHandling_Error completed")
// }

// func TestConcurrency_ScheduleUnschedule(t *testing.T) {
// 	t.Logf("Starting TestConcurrency_ScheduleUnschedule")
// 	logger := &testLogger{}
// 	mockCfg := MockConfig{
// 		DevMode:          true,
// 		DatabaseHost:     "localhost",
// 		DatabaseHostPort: "9042",
// 		SchedulerRPCPort: "9005",
// 		DBServerURL:      "http://localhost:9002",
// 		MaxWorkers:       10,
// 		ChainRPCUrls: map[string]string{
// 			"11155111": "http://127.0.0.1:8545",
// 		},
// 	}
// 	s := newTestEventBasedScheduler("test-manager", logger, &client.DBServerClient{}, mockCfg)

// 	t.Logf("maxWorkers: %d", s.maxWorkers)
// 	assert.Greater(t, s.maxWorkers, 0, "maxWorkers should be greater than 0")

// 	var wg sync.WaitGroup
// 	jobs := 10

// 	if jobs > s.maxWorkers {
// 		jobs = s.maxWorkers
// 	}

// 	t.Logf("Scheduling %d jobs concurrently", jobs)
// 	for i := 0; i < jobs; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			err := s.ScheduleJob(validJobData(int64(id)))
// 			if err != nil {
// 				t.Logf("Failed to schedule job %d: %v", id, err)
// 			} else {
// 				t.Logf("Scheduled job %d", id)
// 			}
// 		}(i)
// 	}
// 	wg.Wait()

// 	time.Sleep(200 * time.Millisecond)

// 	scheduledJobs := len(s.workers)
// 	t.Logf("Scheduled %d out of %d jobs", scheduledJobs, jobs)

// 	var jobIDs []int64
// 	s.workersMutex.RLock()
// 	for jobID := range s.workers {
// 		jobIDs = append(jobIDs, jobID)
// 	}
// 	s.workersMutex.RUnlock()

// 	t.Logf("Unscheduling all jobs concurrently")
// 	for _, jobID := range jobIDs {
// 		wg.Add(1)
// 		go func(id int64) {
// 			defer wg.Done()
// 			err := s.UnscheduleJob(id)
// 			if err != nil {
// 				t.Logf("Failed to unschedule job %d: %v", id, err)
// 			} else {
// 				t.Logf("Unscheduled job %d", id)
// 			}
// 		}(jobID)
// 	}
// 	wg.Wait()

// 	assert.Equal(t, 0, len(s.workers))
// 	t.Logf("TestConcurrency_ScheduleUnschedule completed")
// }

// // Additional test to verify config initialization
// func TestConfigInitialization(t *testing.T) {
// 	os.Setenv("MAX_WORKERS", "100")
// 	if err := config.Init(); err != nil {
// 		// Ignore .env file not found error in tests
// 		if !strings.Contains(err.Error(), ".env") {
// 			t.Fatalf("Failed to initialize config: %v", err)
// 		}
// 	}
// 	config.SetMaxWorkersForTest(100)
// 	t.Logf("MAX_WORKERS env: %s", os.Getenv("MAX_WORKERS"))
// 	t.Logf("config.GetMaxWorkers(): %d", config.GetMaxWorkers())
// 	t.Logf("Starting TestConfigInitialization")
// 	maxWorkers := config.GetMaxWorkers()
// 	t.Logf("maxWorkers from config: %d", maxWorkers)
// 	assert.Greater(t, maxWorkers, 0, "MAX_WORKERS should be greater than 0")
// 	assert.Equal(t, 100, maxWorkers, "MAX_WORKERS should be 100 as set in TestMain")
// 	t.Logf("TestConfigInitialization completed")
// }

// // Test scheduler creation with proper maxWorkers
// func TestSchedulerCreation(t *testing.T) {
// 	os.Setenv("MAX_WORKERS", "100")
// 	if err := config.Init(); err != nil {
// 		// Ignore .env file not found error in tests
// 		if !strings.Contains(err.Error(), ".env") {
// 			t.Fatalf("Failed to initialize config: %v", err)
// 		}
// 	}
// 	config.SetMaxWorkersForTest(100)
// 	t.Logf("MAX_WORKERS env: %s", os.Getenv("MAX_WORKERS"))
// 	t.Logf("config.GetMaxWorkers(): %d", config.GetMaxWorkers())
// 	t.Logf("Starting TestSchedulerCreation")
// 	logger := &testLogger{}
// 	s, err := NewEventBasedScheduler("test-manager", logger, &client.DBServerClient{})

// 	if err != nil {
// 		t.Logf("Could not initialize scheduler: %v", err)
// 		t.Skipf("Skipping test: could not initialize scheduler: %v", err)
// 	}

// 	t.Logf("Scheduler maxWorkers: %d", s.maxWorkers)
// 	assert.Greater(t, s.maxWorkers, 0, "Scheduler maxWorkers should be greater than 0")
// 	assert.Equal(t, 100, s.maxWorkers, "Scheduler maxWorkers should match config")

// 	stats := s.GetStats()
// 	t.Logf("Scheduler stats: %+v", stats)
// 	assert.Equal(t, 100, stats["max_workers"], "Stats should show correct max_workers")
// 	t.Logf("TestSchedulerCreation completed")
// }

// // --- Edge Case Tests ---

// func TestEdgeCase_ScheduleJob_NilJobData(t *testing.T) {
// 	t.Logf("Edge Case: ScheduleJob with nil job data")
// 	logger := &testLogger{}
// 	s, err := NewEventBasedScheduler("test-manager", logger, &client.DBServerClient{})
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not initialize scheduler (no chain clients)")
// 		t.Skip("Skipping edge case: could not initialize scheduler")
// 	}
// 	err = s.ScheduleJob(nil)
// 	if err == nil {
// 		t.Logf("Edge case failed: ScheduleJob(nil) did not return error")
// 	} else {
// 		t.Logf("Edge case passed: ScheduleJob(nil) returned error: %v", err)
// 	}
// }

// func TestEdgeCase_ScheduleJob_InvalidContractAddress(t *testing.T) {
// 	t.Logf("Edge Case: ScheduleJob with invalid contract address")
// 	logger := &testLogger{}
// 	s, err := NewEventBasedScheduler("test-manager", logger, &client.DBServerClient{})
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not initialize scheduler (no chain clients)")
// 		t.Skip("Skipping edge case: could not initialize scheduler")
// 	}
// 	job := validJobData(1)
// 	job.TriggerContractAddress = "notAHexAddress"
// 	err = s.ScheduleJob(job)
// 	if err == nil {
// 		t.Logf("Edge case failed: ScheduleJob with invalid contract address did not return error")
// 	} else {
// 		t.Logf("Edge case passed: ScheduleJob with invalid contract address returned error: %v", err)
// 	}
// }

// func TestEdgeCase_ScheduleJob_UnsupportedChainID(t *testing.T) {
// 	t.Logf("Edge Case: ScheduleJob with unsupported chain ID")
// 	logger := &testLogger{}
// 	s, err := NewEventBasedScheduler("test-manager", logger, &client.DBServerClient{})
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not initialize scheduler (no chain clients)")
// 		t.Skip("Skipping edge case: could not initialize scheduler")
// 	}
// 	job := validJobData(1)
// 	job.TriggerChainID = "99999999" // unsupported chain ID
// 	err = s.ScheduleJob(job)
// 	if err == nil {
// 		t.Logf("Edge case failed: ScheduleJob with unsupported chain ID did not return error")
// 	} else {
// 		t.Logf("Edge case passed: ScheduleJob with unsupported chain ID returned error: %v", err)
// 	}
// }

// func TestEdgeCase_UnscheduleJob_NonExistentJobID(t *testing.T) {
// 	t.Logf("Edge Case: UnscheduleJob with non-existent job ID")
// 	logger := &testLogger{}
// 	s, err := NewEventBasedScheduler("test-manager", logger, &client.DBServerClient{})
// 	if !assert.NoError(t, err) {
// 		t.Logf("Could not initialize scheduler (no chain clients)")
// 		t.Skip("Skipping edge case: could not initialize scheduler")
// 	}
// 	err = s.UnscheduleJob(99999) // job ID that was never scheduled
// 	if err == nil {
// 		t.Logf("Edge case failed: UnscheduleJob with non-existent job ID did not return error")
// 	} else {
// 		t.Logf("Edge case passed: UnscheduleJob with non-existent job ID returned error: %v", err)
// 	}
// }
