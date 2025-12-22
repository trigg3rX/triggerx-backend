package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

const (
	// DuplicateEventWindow is the window to prevent duplicate event processing
	// This matches the value in scheduler/worker package to avoid circular import
	DuplicateEventWindow = 30 * time.Second
)

var (
	// Internal tracking variables for performance calculations
	conditionStatsLock       sync.RWMutex
	conditionCheckTimes      []float64
	conditionsLastMinute     map[string]int64 // chain_id -> count
	lastMinuteReset          time.Time
	workerStartTimes         map[string]time.Time // job_id -> start_time
	workerMemoryUsage        map[string]int64     // job_id -> memory_bytes
	totalConditions          int64
	successfulConditions     int64
	totalActions             int64
	successfulActions        int64
	lastConfigUpdate         time.Time
	configUpdateInterval     = 30 * time.Second
	conditionEvaluationTimes []float64
	actionExecutionTimes     map[string][]float64 // job_id -> execution times
)

func init() {
	conditionsLastMinute = make(map[string]int64)
	workerStartTimes = make(map[string]time.Time)
	workerMemoryUsage = make(map[string]int64)
	actionExecutionTimes = make(map[string][]float64)
	lastMinuteReset = time.Now()
	lastConfigUpdate = time.Now()
}

// Collects system resource metrics
func collectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage (current allocated bytes)
	if memoryUsageBytes != nil {
		memoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}

	// Update CPU usage (using system memory as a proxy)
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		if cpuUsagePercent != nil {
			cpuUsagePercent.Set(ctx, cpuPercent[0])
		}
	} else {
		// Fallback to 0.0 if CPU monitoring fails
		if cpuUsagePercent != nil {
			cpuUsagePercent.Set(ctx, 0.0)
		}
	}

	// Update active goroutines count
	if goroutinesActive != nil {
		goroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
	}

	// Update garbage collection duration (total pause time in seconds)
	if gcDurationSeconds != nil {
		gcDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
	}
}

// Collects configuration-based metrics
func collectConfigurationMetrics() {
	now := time.Now()
	if now.Sub(lastConfigUpdate) < configUpdateInterval {
		return
	}
	lastConfigUpdate = now

	// Set duplicate condition window from configuration
	if duplicateConditionWindowSeconds != nil {
		duplicateConditionWindowSeconds.Set(ctx, getDuplicateConditionWindowSeconds())
	}
}

// Collects performance-related metrics
func collectPerformanceMetrics() {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	// Update conditions per minute for each chain
	now := time.Now()
	if now.Sub(lastMinuteReset) >= time.Minute {
		// Take a snapshot of current counts before resetting
		conditionsSnapshot := make(map[string]int64)
		for chainID, count := range conditionsLastMinute {
			conditionsSnapshot[chainID] = count
			if eventsPerMinute != nil {
				eventsPerMinute.WithLabelValues(chainID).Set(ctx, float64(count))
			}
		}

		// Reset for next minute in a separate goroutine to avoid blocking
		go func() {
			conditionStatsLock.Lock()
			defer conditionStatsLock.Unlock()

			// Only reset if enough time has passed (prevent race conditions)
			if time.Since(lastMinuteReset) >= time.Minute {
				conditionsLastMinute = make(map[string]int64)
				lastMinuteReset = time.Now()
			}
		}()
	}

	// Calculate average condition check time
	if len(conditionCheckTimes) > 0 {
		var sum float64
		for _, duration := range conditionCheckTimes {
			sum += duration
		}
		avgTime := sum / float64(len(conditionCheckTimes))
		if averageConditionCheckTimeSeconds != nil {
			averageConditionCheckTimeSeconds.Set(ctx, avgTime)
		}
	}
}

// Collects worker-related metrics
func collectWorkerMetrics() {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	// Update worker uptime for each active worker
	now := time.Now()
	for jobID, startTime := range workerStartTimes {
		uptime := now.Sub(startTime).Seconds()
		if workerUptimeSeconds != nil {
			workerUptimeSeconds.WithLabelValues(jobID).Set(ctx, uptime)
		}
	}

	// Update worker memory usage
	for jobID, memUsage := range workerMemoryUsage {
		if workerMemoryUsageBytes != nil {
			workerMemoryUsageBytes.WithLabelValues(jobID).Set(ctx, float64(memUsage))
		}
	}
}

// Resets metrics that should be reset daily
func resetDailyMetrics() {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()

	// Reset daily counters
	for chainID := range conditionsLastMinute {
		if eventsPerMinute != nil {
			eventsPerMinute.WithLabelValues(chainID).Set(ctx, 0)
		}
	}
	if averageConditionCheckTimeSeconds != nil {
		averageConditionCheckTimeSeconds.Set(ctx, 0)
	}

	// Reset tracking variables
	conditionCheckTimes = nil
	conditionsLastMinute = make(map[string]int64)
	conditionEvaluationTimes = nil
	actionExecutionTimes = make(map[string][]float64)
	totalConditions = 0
	successfulConditions = 0
	totalActions = 0
	successfulActions = 0
	lastMinuteReset = time.Now()
}

// Helper functions to get configuration values
func getDuplicateConditionWindowSeconds() float64 {
	// Use the local constant to avoid circular import
	return DuplicateEventWindow.Seconds()
}

// HTTP and API tracking functions

// TrackHTTPRequest tracks HTTP request metrics
func TrackHTTPRequest(method, endpoint, statusCode string) {
	if httpRequestsTotal != nil {
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc(ctx)
	}
}

// TrackHTTPClientConnectionError tracks HTTP client connection errors
func TrackHTTPClientConnectionError() {
	if httpClientConnectionErrorsTotal != nil {
		httpClientConnectionErrorsTotal.Inc(ctx)
	}
}

// Database tracking functions

// TrackDBRequest tracks database request metrics
func TrackDBRequest(method, endpoint, status string) {
	if dbRequestsTotal != nil {
		dbRequestsTotal.WithLabelValues(method, endpoint, status).Inc(ctx)
	}
}

// TrackDBConnectionError tracks database connection errors
func TrackDBConnectionError() {
	if dbConnectionErrorsTotal != nil {
		dbConnectionErrorsTotal.Inc(ctx)
	}
}

// TrackDBRetry tracks database retry attempts
func TrackDBRetry(endpoint string) {
	if dbRetriesTotal != nil {
		dbRetriesTotal.WithLabelValues(endpoint).Inc(ctx)
	}
}

// Job and Worker tracking functions

// TrackJobScheduled tracks when a job is scheduled
func TrackJobScheduled() {
	if jobsScheduled != nil {
		jobsScheduled.Inc(ctx)
	}
}

// TrackJobCompleted tracks when a job completes
func TrackJobCompleted(status string) {
	if jobsCompleted != nil {
		jobsCompleted.WithLabelValues(status).Inc(ctx)
	}
}

// UpdateActiveWorkers updates the count of active workers
func UpdateActiveWorkers(count int) {
	if activeWorkers != nil {
		activeWorkers.Set(ctx, float64(count))
	}
}

// TrackWorkerStart tracks when a worker starts
func TrackWorkerStart(jobID string) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	workerStartTimes[jobID] = time.Now()
}

// TrackWorkerStop tracks when a worker stops
func TrackWorkerStop(jobID string) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	delete(workerStartTimes, jobID)
	delete(workerMemoryUsage, jobID)
	delete(actionExecutionTimes, jobID)
}

// TrackWorkerMemoryUsage updates worker memory usage
func TrackWorkerMemoryUsage(jobID string, memoryBytes int64) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	workerMemoryUsage[jobID] = memoryBytes
}

// Condition-specific tracking functions

// TrackConditionEvaluation tracks condition evaluation metrics
func TrackConditionEvaluation(duration time.Duration) {
	if conditionEvaluationDuration != nil {
		conditionEvaluationDuration.Record(ctx, duration.Seconds())
	}

	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	conditionEvaluationTimes = append(conditionEvaluationTimes, duration.Seconds())
}

// TrackConditionByType tracks conditions by their type
func TrackConditionByType(conditionType string) {
	if conditionsByTypeTotal != nil {
		conditionsByTypeTotal.WithLabelValues(conditionType).Inc(ctx)
	}
}

// TrackConditionBySource tracks conditions by their source type
func TrackConditionBySource(sourceType string) {
	if conditionsBySourceTotal != nil {
		conditionsBySourceTotal.WithLabelValues(sourceType).Inc(ctx)
	}
}

// TrackAPIResponse tracks API response status
func TrackAPIResponse(sourceURL, statusCode string) {
	if apiResponseStatusTotal != nil {
		apiResponseStatusTotal.WithLabelValues(sourceURL, statusCode).Inc(ctx)
	}
}

// TrackValueParsingError tracks value parsing errors
func TrackValueParsingError(sourceType string) {
	if valueParsingErrorsTotal != nil {
		valueParsingErrorsTotal.WithLabelValues(sourceType).Inc(ctx)
	}
}

// TrackInvalidValue tracks invalid values received
func TrackInvalidValue(source string) {
	if invalidValuesTotal != nil {
		invalidValuesTotal.WithLabelValues(source).Inc(ctx)
	}
}

// Action execution tracking functions

// TrackActionExecution tracks action execution with duration
func TrackActionExecution(jobID string, duration time.Duration) {
	if actionExecutionDuration != nil {
		actionExecutionDuration.WithLabelValues(jobID).Record(ctx, duration.Seconds())
	}

	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	if actionExecutionTimes[jobID] == nil {
		actionExecutionTimes[jobID] = make([]float64, 0)
	}
	actionExecutionTimes[jobID] = append(actionExecutionTimes[jobID], duration.Seconds())
}

// Error and recovery tracking functions

// TrackTimeout tracks operation timeouts
func TrackTimeout(operation string) {
	if timeoutsTotal != nil {
		timeoutsTotal.WithLabelValues(operation).Inc(ctx)
	}
}

// TrackCriticalError tracks critical system errors
func TrackCriticalError(errorType string) {
	if criticalErrorsTotal != nil {
		criticalErrorsTotal.WithLabelValues(errorType).Inc(ctx)
	}
}

// Condition processing tracking functions

// TrackConditionCheck tracks condition checks with duration and success
func TrackConditionCheck(chainID string, duration time.Duration, success bool) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()

	// Update conditions per minute
	if conditionsLastMinute[chainID] == 0 {
		conditionsLastMinute[chainID] = 0
	}
	conditionsLastMinute[chainID]++

	// Track condition check time
	conditionCheckTimes = append(conditionCheckTimes, duration.Seconds())

	// Update totals
	totalConditions++
	if success {
		successfulConditions++
	}
}

// Statistics and reporting functions

// GetConditionStats returns condition processing statistics
func GetConditionStats() (total, successful int64, avgCheckTime float64) {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	total = totalConditions
	successful = successfulConditions

	if len(conditionCheckTimes) > 0 {
		var sum float64
		for _, duration := range conditionCheckTimes {
			sum += duration
		}
		avgCheckTime = sum / float64(len(conditionCheckTimes))
	}

	return total, successful, avgCheckTime
}

// GetActionStats returns action execution statistics
func GetActionStats() (total, successful int64) {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()
	return totalActions, successfulActions
}

// GetWorkerCount returns the current number of active workers
func GetWorkerCount() int {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()
	return len(workerStartTimes)
}

// GetConditionEvaluationStats returns condition evaluation performance statistics
func GetConditionEvaluationStats() (count int, avgDuration float64, maxDuration float64) {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	count = len(conditionEvaluationTimes)
	if count == 0 {
		return 0, 0, 0
	}

	var sum, max float64
	for _, duration := range conditionEvaluationTimes {
		sum += duration
		if duration > max {
			max = duration
		}
	}

	avgDuration = sum / float64(count)
	maxDuration = max

	return count, avgDuration, maxDuration
}

// GetSchedulerHealthStatus returns overall scheduler health status
func GetSchedulerHealthStatus() map[string]interface{} {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	total, successful, avgCheckTime := GetConditionStats()
	totalActions, successfulActions := GetActionStats()
	workerCount := GetWorkerCount()
	evalCount, avgEvalDuration, maxEvalDuration := GetConditionEvaluationStats()

	return map[string]interface{}{
		"total_conditions":        total,
		"successful_conditions":   successful,
		"condition_success_rate":  float64(successful) / float64(total) * 100,
		"average_check_time":      avgCheckTime,
		"total_actions":           totalActions,
		"successful_actions":      successfulActions,
		"action_success_rate":     float64(successfulActions) / float64(totalActions) * 100,
		"active_workers":          workerCount,
		"evaluations_count":       evalCount,
		"average_evaluation_time": avgEvalDuration,
		"max_evaluation_time":     maxEvalDuration,
		"conditions_last_minute":  len(conditionsLastMinute),
	}
}

var (
	// Internal tracking variables for performance calculations
	eventStatsLock       sync.RWMutex
	eventProcessingTimes []float64
	eventsLastMinute     map[string]int64 // chain_id -> count
	totalEvents          int64
	successfulEvents     int64
)

func init() {
	eventsLastMinute = make(map[string]int64)
	workerStartTimes = make(map[string]time.Time)
	workerMemoryUsage = make(map[string]int64)
	lastMinuteReset = time.Now()
	lastConfigUpdate = time.Now()
}

// Blockchain and RPC tracking functions

// TrackChainConnection tracks blockchain connection attempts
func TrackChainConnection(chainID, status string) {
	if chainConnectionsTotal != nil {
		chainConnectionsTotal.WithLabelValues(chainID, status).Inc(ctx)
	}
}

// TrackRPCRequest tracks RPC requests to blockchain nodes
func TrackRPCRequest(chainID, method, status string) {
	if rpcRequestsTotal != nil {
		rpcRequestsTotal.WithLabelValues(chainID, method, status).Inc(ctx)
	}
}

// TrackConnectionFailure tracks blockchain connection failures
func TrackConnectionFailure(chainID string) {
	if connectionFailuresTotal != nil {
		connectionFailuresTotal.WithLabelValues(chainID).Inc(ctx)
	}
}

// TrackWorkerError tracks worker errors
func TrackWorkerError(jobID, errorType string) {
	if workerErrorsTotal != nil {
		workerErrorsTotal.WithLabelValues(jobID, errorType).Inc(ctx)
	}
}

// Event tracking functions

// TrackEvent tracks when an event is detected
func TrackEvent(chainID string, processingTime time.Duration) {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()

	// Update events per minute counter
	eventsLastMinute[chainID]++
	totalEvents++

	// Track processing time (keep last 1000 entries to avoid memory growth)
	eventProcessingTimes = append(eventProcessingTimes, processingTime.Seconds())
	if len(eventProcessingTimes) > 1000 {
		eventProcessingTimes = eventProcessingTimes[1:]
	}
}

// TrackEventSuccess tracks successful event processing
func TrackEventSuccess(chainID string) {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()
	successfulEvents++
}

// Stats and utility functions

// GetEventStats returns current event statistics
func GetEventStats() (total, successful int64, avgProcessingTime float64) {
	eventStatsLock.RLock()
	defer eventStatsLock.RUnlock()

	total = totalEvents
	successful = successfulEvents

	if len(eventProcessingTimes) > 0 {
		var sum float64
		for _, duration := range eventProcessingTimes {
			sum += duration
		}
		avgProcessingTime = sum / float64(len(eventProcessingTimes))
	}

	return
}

// TrackEventWithDuration tracks event processing with comprehensive metrics
func TrackEventWithDuration(chainID string, duration time.Duration, success bool) {
	// Track the basic event
	TrackEvent(chainID, duration)

	// Track success/failure
	if success {
		TrackEventSuccess(chainID)
	}

	// Update average processing time immediately for real-time accuracy
	eventStatsLock.RLock()
	if len(eventProcessingTimes) > 0 {
		var sum float64
		for _, processingTime := range eventProcessingTimes {
			sum += processingTime
		}
		avgTime := sum / float64(len(eventProcessingTimes))
		if averageEventProcessingTimeSeconds != nil {
			averageEventProcessingTimeSeconds.Set(ctx, avgTime)
		}
	}
	eventStatsLock.RUnlock()
}

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if uptimeSeconds != nil {
				uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
			collectSystemMetrics()
			collectConfigurationMetrics()
			collectPerformanceMetrics()
			collectWorkerMetrics()
		}
	}()

	// Reset daily metrics every day at midnight
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			resetDailyMetrics()
		}
	}()
}
