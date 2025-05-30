package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval             = 1 * time.Second  // Poll every 1 second as requested
	workerTimeout            = 30 * time.Second // Timeout for worker operations
	maxRetries               = 3                // Max retries for failed operations
	requestTimeout           = 10 * time.Second // HTTP request timeout
	performerLockTTL         = 15 * time.Minute // Lock duration for condition monitoring
	valueCacheTTL            = 30 * time.Second // Cache TTL for API values
	conditionStateCacheTTL   = 5 * time.Minute  // Cache TTL for condition state
	duplicateConditionWindow = 10 * time.Second // Window to prevent duplicate condition processing
)

// Supported condition types
const (
	ConditionGreaterThan  = "greater_than"
	ConditionLessThan     = "less_than"
	ConditionBetween      = "between"
	ConditionEquals       = "equals"
	ConditionNotEquals    = "not_equals"
	ConditionGreaterEqual = "greater_equal"
	ConditionLessEqual    = "less_equal"
)

// Supported value source types
const (
	SourceTypeAPI    = "api"
	SourceTypeOracle = "oracle"
	SourceTypeStatic = "static"
)

// ConditionBasedScheduler manages individual job workers for condition monitoring
type ConditionBasedScheduler struct {
	ctx          context.Context
	cancel       context.CancelFunc
	logger       logging.Logger
	workers      map[int64]*ConditionWorker // jobID -> worker
	workersMutex sync.RWMutex
	dbClient     *client.DBServerClient
	cache        cache.Cache
	metrics      *metrics.Collector
	managerID    string
	httpClient   *http.Client
	maxWorkers   int
}

// ConditionWorker represents an individual worker monitoring a specific condition
type ConditionWorker struct {
	job          *schedulerTypes.ConditionJobData
	logger       logging.Logger
	dbClient     *client.DBServerClient
	cache        cache.Cache
	httpClient   *http.Client
	ctx          context.Context
	cancel       context.CancelFunc
	isRunning    bool
	mutex        sync.RWMutex
	lastValue    float64
	lastCheck    time.Time
	conditionMet int64 // Count of consecutive condition met checks
	managerID    string
}

// JobScheduleRequest represents the request to schedule a new condition job
type JobScheduleRequest struct {
	JobID                         int64    `json:"job_id" binding:"required"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	ConditionType                 string   `json:"condition_type" binding:"required"`
	UpperLimit                    float64  `json:"upper_limit"`
	LowerLimit                    float64  `json:"lower_limit"`
	ValueSourceType               string   `json:"value_source_type" binding:"required"`
	ValueSourceUrl                string   `json:"value_source_url" binding:"required"`
	TargetChainID                 string   `json:"target_chain_id" binding:"required"`
	TargetContractAddress         string   `json:"target_contract_address" binding:"required"`
	TargetFunction                string   `json:"target_function" binding:"required"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}

// ValueResponse represents a generic response structure for fetching values
type ValueResponse struct {
	Value     float64 `json:"value"`
	Price     float64 `json:"price"`     // Common for price APIs
	USD       float64 `json:"usd"`       // Common for CoinGecko-style APIs
	Rate      float64 `json:"rate"`      // Common for exchange rate APIs
	Result    float64 `json:"result"`    // Generic result field
	Data      float64 `json:"data"`      // Generic data field
	Timestamp int64   `json:"timestamp"` // Optional timestamp
}

// NewConditionBasedScheduler creates a new instance of ConditionBasedScheduler
func NewConditionBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*ConditionBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	// Initialize cache
	if err := cache.Init(); err != nil {
		logger.Warnf("Failed to initialize cache: %v", err)
	}

	cacheInstance, err := cache.GetCache()
	if err != nil {
		logger.Warnf("Cache not available, running without cache: %v", err)
	}

	// Test Redis connection
	if err := redisx.Ping(); err != nil {
		logger.Warnf("Redis not available, condition streaming disabled: %v", err)
	} else {
		logger.Info("Redis connection established, condition streaming enabled")
	}

	scheduler := &ConditionBasedScheduler{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		workers:   make(map[int64]*ConditionWorker),
		dbClient:  dbClient,
		cache:     cacheInstance,
		metrics:   metrics.NewCollector(),
		managerID: managerID,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	scheduler.logger.Info("Condition-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"cache_available", cacheInstance != nil,
	)

	return scheduler, nil
}

// ScheduleJob creates and starts a new condition worker
func (s *ConditionBasedScheduler) ScheduleJob(jobData *schedulerTypes.ConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Check if we've reached the maximum number of workers
	if len(s.workers) >= s.maxWorkers {
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Validate condition type
	if !isValidConditionType(jobData.ConditionType) {
		return fmt.Errorf("unsupported condition type: %s", jobData.ConditionType)
	}

	// Validate value source type
	if !isValidSourceType(jobData.ValueSourceType) {
		return fmt.Errorf("unsupported value source type: %s", jobData.ValueSourceType)
	}

	// Create condition worker
	worker, err := s.createConditionWorker(jobData)
	if err != nil {
		return fmt.Errorf("failed to create condition worker: %w", err)
	}

	// Store worker
	s.workers[jobData.JobID] = worker

	// Start worker
	go worker.start()

	// Update metrics
	metrics.JobsScheduled.Inc()
	metrics.JobsRunning.Inc()

	// Add job scheduling event to Redis stream
	jobContext := map[string]interface{}{
		"job_id":            jobData.JobID,
		"condition_type":    jobData.ConditionType,
		"upper_limit":       jobData.UpperLimit,
		"lower_limit":       jobData.LowerLimit,
		"value_source_type": jobData.ValueSourceType,
		"value_source_url":  jobData.ValueSourceUrl,
		"manager_id":        s.managerID,
		"scheduled_at":      time.Now().Unix(),
		"status":            "scheduled",
	}

	if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
		s.logger.Warnf("Failed to add condition job scheduling event to Redis stream: %v", err)
	}

	s.logger.Info("Condition job scheduled successfully",
		"job_id", jobData.JobID,
		"condition_type", jobData.ConditionType,
		"value_source", jobData.ValueSourceUrl,
		"upper_limit", jobData.UpperLimit,
		"lower_limit", jobData.LowerLimit,
		"active_workers", len(s.workers),
		"max_workers", s.maxWorkers,
	)

	return nil
}

// createConditionWorker creates a new condition worker instance
func (s *ConditionBasedScheduler) createConditionWorker(jobData *schedulerTypes.ConditionJobData) (*ConditionWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	worker := &ConditionWorker{
		job:        jobData,
		logger:     s.logger,
		dbClient:   s.dbClient,
		cache:      s.cache,
		httpClient: s.httpClient,
		ctx:        ctx,
		cancel:     cancel,
		isRunning:  false,
		lastCheck:  time.Now(),
		managerID:  s.managerID,
	}

	return worker, nil
}

// start begins the condition worker's monitoring loop
func (w *ConditionWorker) start() {
	w.mutex.Lock()
	w.isRunning = true
	w.mutex.Unlock()

	// Try to acquire performer lock
	lockKey := fmt.Sprintf("condition_job_%d", w.job.JobID)
	if w.cache != nil {
		acquired, err := w.cache.AcquirePerformerLock(lockKey, performerLockTTL)
		if err != nil {
			w.logger.Warnf("Failed to acquire performer lock for condition job %d: %v", w.job.JobID, err)
		} else if !acquired {
			w.logger.Warnf("Condition job %d is already being monitored by another worker, stopping", w.job.JobID)
			return
		}
		defer func() {
			if err := w.cache.ReleasePerformerLock(lockKey); err != nil {
				w.logger.Warnf("Failed to release performer lock for condition job %d: %v", w.job.JobID, err)
			}
		}()
	}

	w.logger.Info("Starting condition worker",
		"job_id", w.job.JobID,
		"condition_type", w.job.ConditionType,
		"value_source", w.job.ValueSourceUrl,
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Condition worker stopped", "job_id", w.job.JobID)
			return
		case <-ticker.C:
			if err := w.checkCondition(); err != nil {
				w.logger.Error("Error checking condition", "job_id", w.job.JobID, "error", err)
				metrics.JobsFailed.Inc()
			}
		}
	}
}

// checkCondition fetches the current value and checks if condition is satisfied
func (w *ConditionWorker) checkCondition() error {
	startTime := time.Now()

	// Track condition check
	metrics.ConditionsChecked.Inc()

	// Check for duplicate condition check prevention
	checkKey := fmt.Sprintf("condition_check_%d_%d", w.job.JobID, startTime.Unix())
	if w.cache != nil {
		if _, err := w.cache.Get(checkKey); err == nil {
			w.logger.Debug("Condition check already performed recently, skipping", "job_id", w.job.JobID)
			return nil
		}
		// Mark this check time to prevent duplicates
		w.cache.Set(checkKey, time.Now().Format(time.RFC3339), duplicateConditionWindow)
	}

	// Fetch current value from source (with caching)
	currentValue, err := w.fetchValueWithCache()
	if err != nil {
		return fmt.Errorf("failed to fetch value: %w", err)
	}

	w.lastValue = currentValue
	w.lastCheck = time.Now()

	// Create condition check context for Redis streaming
	conditionContext := map[string]interface{}{
		"job_id":         w.job.JobID,
		"manager_id":     w.managerID,
		"current_value":  currentValue,
		"condition_type": w.job.ConditionType,
		"upper_limit":    w.job.UpperLimit,
		"lower_limit":    w.job.LowerLimit,
		"checked_at":     startTime.Unix(),
	}

	// Check if condition is satisfied
	satisfied, err := w.evaluateCondition(currentValue)
	if err != nil {
		conditionContext["status"] = "evaluation_error"
		conditionContext["error"] = err.Error()
		redisx.AddJobToStream(redisx.JobsRetryEventStream, conditionContext)
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	if satisfied {
		w.conditionMet++
		metrics.ConditionsSatisfied.Inc()

		conditionContext["status"] = "satisfied"
		conditionContext["consecutive_checks"] = w.conditionMet

		w.logger.Info("Condition satisfied",
			"job_id", w.job.JobID,
			"current_value", currentValue,
			"condition_type", w.job.ConditionType,
			"upper_limit", w.job.UpperLimit,
			"lower_limit", w.job.LowerLimit,
			"consecutive_checks", w.conditionMet,
		)

		// Add satisfied condition to Redis stream
		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, conditionContext); err != nil {
			w.logger.Warnf("Failed to add condition satisfied event to Redis stream: %v", err)
		}

		// Execute action
		executionSuccess := w.performActionExecution(currentValue)

		duration := time.Since(startTime)
		conditionContext["duration_ms"] = duration.Milliseconds()
		conditionContext["completed_at"] = time.Now().Unix()

		if executionSuccess {
			conditionContext["action_status"] = "completed"
			if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, conditionContext); err != nil {
				w.logger.Warnf("Failed to add condition action completion to Redis stream: %v", err)
			}
		} else {
			conditionContext["action_status"] = "failed"
			if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, conditionContext); err != nil {
				w.logger.Warnf("Failed to add condition action failure to Redis stream: %v", err)
			}
		}
	} else {
		w.conditionMet = 0
		conditionContext["status"] = "not_satisfied"

		w.logger.Debug("Condition not satisfied",
			"job_id", w.job.JobID,
			"current_value", currentValue,
			"condition_type", w.job.ConditionType,
		)

		// Periodically log condition checks for monitoring
		if time.Now().Unix()%60 == 0 { // Every minute
			if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, conditionContext); err != nil {
				w.logger.Warnf("Failed to add condition check status to Redis stream: %v", err)
			}
		}
	}

	// Cache the condition state
	if w.cache != nil {
		w.cacheConditionState(currentValue, satisfied)
	}

	return nil
}

// fetchValueWithCache retrieves the current value with caching support
func (w *ConditionWorker) fetchValueWithCache() (float64, error) {
	// Try to get cached value first
	if w.cache != nil {
		cacheKey := fmt.Sprintf("value_%d_%s", w.job.JobID, w.job.ValueSourceUrl)
		if cached, err := w.cache.Get(cacheKey); err == nil {
			var cachedValue float64
			if _, err := fmt.Sscanf(cached, "%f", &cachedValue); err == nil {
				w.logger.Debug("Using cached value", "job_id", w.job.JobID, "value", cachedValue)
				return cachedValue, nil
			}
		}
	}

	// Fetch fresh value
	currentValue, err := w.fetchValue()
	if err != nil {
		return 0, err
	}

	// Cache the value
	if w.cache != nil {
		cacheKey := fmt.Sprintf("value_%d_%s", w.job.JobID, w.job.ValueSourceUrl)
		w.cache.Set(cacheKey, fmt.Sprintf("%f", currentValue), valueCacheTTL)
	}

	return currentValue, nil
}

// cacheConditionState caches the current condition state
func (w *ConditionWorker) cacheConditionState(value float64, satisfied bool) {
	if w.cache == nil {
		return
	}

	stateData := map[string]interface{}{
		"job_id":        w.job.JobID,
		"current_value": value,
		"satisfied":     satisfied,
		"condition_met": w.conditionMet,
		"last_check":    w.lastCheck.Unix(),
		"cached_at":     time.Now().Unix(),
	}

	if jsonData, err := json.Marshal(stateData); err == nil {
		cacheKey := fmt.Sprintf("condition_state_%d", w.job.JobID)
		w.cache.Set(cacheKey, string(jsonData), conditionStateCacheTTL)
	}
}

// performActionExecution handles the actual action execution logic
func (w *ConditionWorker) performActionExecution(triggerValue float64) bool {
	// TODO: Implement actual action execution logic
	// This should:
	// 1. Send task to manager/keeper for execution
	// 2. Handle response and update job status
	// 3. For non-recurring jobs, stop the worker

	w.logger.Info("Simulating action execution",
		"job_id", w.job.JobID,
		"trigger_value", triggerValue,
		"target_chain", w.job.TargetChainID,
		"target_contract", w.job.TargetContractAddress,
		"target_function", w.job.TargetFunction,
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// For non-recurring jobs, stop the worker after execution
	if !w.job.Recurring {
		w.logger.Info("Non-recurring job completed, stopping worker", "job_id", w.job.JobID)
		go w.stop() // Stop in a goroutine to avoid deadlock
	}

	// Simulate 95% success rate
	return time.Now().UnixNano()%100 < 95
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Stop worker
	worker.stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update metrics
	metrics.JobsRunning.Dec()

	// Add job unscheduling event to Redis stream
	jobContext := map[string]interface{}{
		"job_id":         jobID,
		"manager_id":     s.managerID,
		"unscheduled_at": time.Now().Unix(),
		"status":         "unscheduled",
	}

	if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
		s.logger.Warnf("Failed to add condition job unscheduling event to Redis stream: %v", err)
	}

	s.logger.Info("Condition job unscheduled successfully", "job_id", jobID)
	return nil
}

// stop gracefully stops the condition worker
func (w *ConditionWorker) stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		w.cancel()
		w.isRunning = false
		w.logger.Info("Condition worker stopped", "job_id", w.job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *ConditionWorker) IsRunning() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.isRunning
}

// Start begins the scheduler's main loop (for compatibility)
func (s *ConditionBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Condition-based scheduler ready for job scheduling", "manager_id", s.managerID)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info("Scheduler context cancelled, stopping all workers")
	s.Stop()
}

// Stop gracefully stops all condition workers
func (s *ConditionBasedScheduler) Stop() {
	s.logger.Info("Stopping condition-based scheduler")

	s.cancel()

	// Stop all workers
	s.workersMutex.Lock()
	for jobID, worker := range s.workers {
		worker.stop()
		s.logger.Info("Stopped worker", "job_id", jobID)
	}
	s.workers = make(map[int64]*ConditionWorker)
	s.workersMutex.Unlock()

	s.logger.Info("Condition-based scheduler stopped")
}

// GetStats returns current scheduler statistics
func (s *ConditionBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	runningWorkers := 0
	for _, worker := range s.workers {
		if worker.IsRunning() {
			runningWorkers++
		}
	}

	return map[string]interface{}{
		"manager_id":        s.managerID,
		"total_workers":     len(s.workers),
		"running_workers":   runningWorkers,
		"max_workers":       s.maxWorkers,
		"cache_available":   s.cache != nil,
		"supported_sources": []string{"api", "oracle", "static"},
		"supported_conditions": []string{
			"greater_than", "less_than", "between",
			"equals", "not_equals", "greater_equal", "less_equal",
		},
		"poll_interval_seconds": pollInterval.Seconds(),
	}
}

// GetJobWorkerStats returns statistics for a specific condition worker
func (s *ConditionBasedScheduler) GetJobWorkerStats(jobID int64) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil, fmt.Errorf("job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":              worker.job.JobID,
		"is_running":          worker.IsRunning(),
		"condition_type":      worker.job.ConditionType,
		"upper_limit":         worker.job.UpperLimit,
		"lower_limit":         worker.job.LowerLimit,
		"value_source":        worker.job.ValueSourceUrl,
		"last_value":          worker.lastValue,
		"last_check":          worker.lastCheck,
		"condition_met_count": worker.conditionMet,
		"manager_id":          worker.managerID,
	}, nil
}

// Helper functions

func isValidConditionType(conditionType string) bool {
	validTypes := []string{
		ConditionGreaterThan, ConditionLessThan, ConditionBetween,
		ConditionEquals, ConditionNotEquals, ConditionGreaterEqual, ConditionLessEqual,
	}
	for _, valid := range validTypes {
		if conditionType == valid {
			return true
		}
	}
	return false
}

func isValidSourceType(sourceType string) bool {
	validTypes := []string{SourceTypeAPI, SourceTypeOracle, SourceTypeStatic}
	for _, valid := range validTypes {
		if sourceType == valid {
			return true
		}
	}
	return false
}

// fetchValue retrieves the current value from the configured source
func (w *ConditionWorker) fetchValue() (float64, error) {
	switch w.job.ValueSourceType {
	case SourceTypeAPI:
		return w.fetchFromAPI()
	case SourceTypeOracle:
		return w.fetchFromOracle()
	case SourceTypeStatic:
		return w.fetchStaticValue()
	default:
		return 0, fmt.Errorf("unsupported value source type: %s", w.job.ValueSourceType)
	}
}

// fetchFromAPI fetches value from an HTTP API endpoint
func (w *ConditionWorker) fetchFromAPI() (float64, error) {
	metrics.ValueSourceRequests.Inc()

	resp, err := w.httpClient.Get(w.job.ValueSourceUrl)
	if err != nil {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse response as ValueResponse struct
	var valueResp ValueResponse
	if err := json.Unmarshal(body, &valueResp); err == nil {
		// Check which field has a non-zero value
		if valueResp.Value != 0 {
			return valueResp.Value, nil
		}
		if valueResp.Price != 0 {
			return valueResp.Price, nil
		}
		if valueResp.USD != 0 {
			return valueResp.USD, nil
		}
		if valueResp.Rate != 0 {
			return valueResp.Rate, nil
		}
		if valueResp.Result != 0 {
			return valueResp.Result, nil
		}
		if valueResp.Data != 0 {
			return valueResp.Data, nil
		}
	}

	// Try to parse as a simple float
	var floatValue float64
	if err := json.Unmarshal(body, &floatValue); err == nil {
		return floatValue, nil
	}

	// Try to parse as a simple string and convert to float
	var stringValue string
	if err := json.Unmarshal(body, &stringValue); err == nil {
		if floatVal, parseErr := strconv.ParseFloat(stringValue, 64); parseErr == nil {
			return floatVal, nil
		}
	}

	metrics.ValueSourceErrors.Inc()
	return 0, fmt.Errorf("could not extract numeric value from response: %s", string(body))
}

// fetchFromOracle fetches value from an oracle (placeholder implementation)
func (w *ConditionWorker) fetchFromOracle() (float64, error) {
	// TODO: Implement oracle-specific logic
	// For now, treat as API endpoint
	return w.fetchFromAPI()
}

// fetchStaticValue returns a static value (for testing purposes)
func (w *ConditionWorker) fetchStaticValue() (float64, error) {
	// Parse URL as the static value
	value, err := strconv.ParseFloat(w.job.ValueSourceUrl, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid static value: %s", w.job.ValueSourceUrl)
	}
	return value, nil
}

// evaluateCondition checks if the current value satisfies the condition
func (w *ConditionWorker) evaluateCondition(currentValue float64) (bool, error) {
	switch w.job.ConditionType {
	case ConditionGreaterThan:
		return currentValue > w.job.LowerLimit, nil
	case ConditionLessThan:
		return currentValue < w.job.UpperLimit, nil
	case ConditionBetween:
		return currentValue >= w.job.LowerLimit && currentValue <= w.job.UpperLimit, nil
	case ConditionEquals:
		return currentValue == w.job.LowerLimit, nil
	case ConditionNotEquals:
		return currentValue != w.job.LowerLimit, nil
	case ConditionGreaterEqual:
		return currentValue >= w.job.LowerLimit, nil
	case ConditionLessEqual:
		return currentValue <= w.job.UpperLimit, nil
	default:
		return false, fmt.Errorf("unsupported condition type: %s", w.job.ConditionType)
	}
}
