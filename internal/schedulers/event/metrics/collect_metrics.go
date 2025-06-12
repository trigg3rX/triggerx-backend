package metrics

import (
	"runtime"
	"sync"
	"time"
)

type HealthChecker interface {
	HealthCheck() error
}

var (
	dbChecker HealthChecker

	// Internal tracking variables for performance calculations
	eventStatsLock       sync.RWMutex
	eventProcessingTimes []float64
	eventsLastMinute     map[string]int64 // chain_id -> count
	lastMinuteReset      time.Time
	workerStartTimes     map[string]time.Time // job_id -> start_time
	workerMemoryUsage    map[string]int64     // job_id -> memory_bytes
	totalEvents          int64
	successfulEvents     int64
	totalActions         int64
	successfulActions    int64
	lastConfigUpdate     time.Time
	configUpdateInterval = 30 * time.Second
)

func init() {
	eventsLastMinute = make(map[string]int64)
	workerStartTimes = make(map[string]time.Time)
	workerMemoryUsage = make(map[string]int64)
	lastMinuteReset = time.Now()
	lastConfigUpdate = time.Now()
}

// SetHealthChecker sets the health checker for database status monitoring
func SetHealthChecker(checker HealthChecker) {
	dbChecker = checker
}

// Collects system resource metrics
func collectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage (current allocated bytes)
	MemoryUsageBytes.Set(float64(memStats.Alloc))

	// Update CPU usage (using system memory as a proxy)
	CPUUsagePercent.Set(float64(memStats.Sys))

	// Update active goroutines count
	GoroutinesActive.Set(float64(runtime.NumGoroutine()))

	// Update garbage collection duration (total pause time in seconds)
	GCDurationSeconds.Set(float64(memStats.PauseTotalNs) / 1e9)
}

// Collects configuration-based metrics
func collectConfigurationMetrics() {
	now := time.Now()
	if now.Sub(lastConfigUpdate) < configUpdateInterval {
		return
	}
	lastConfigUpdate = now

	// Set duplicate event window from configuration
	DuplicateEventWindowSeconds.Set(getDuplicateEventWindowSeconds())
}

// Collects performance-related metrics
func collectPerformanceMetrics() {
	eventStatsLock.RLock()
	defer eventStatsLock.RUnlock()

	// Update events per minute for each chain
	if time.Since(lastMinuteReset) >= time.Minute {
		for chainID, count := range eventsLastMinute {
			EventsPerMinute.WithLabelValues(chainID).Set(float64(count))
		}

		// Reset for next minute in a separate goroutine to avoid blocking
		go func() {
			eventStatsLock.Lock()
			eventsLastMinute = make(map[string]int64)
			lastMinuteReset = time.Now()
			eventStatsLock.Unlock()
		}()
	}

	// Calculate average event processing time
	if len(eventProcessingTimes) > 0 {
		var sum float64
		for _, duration := range eventProcessingTimes {
			sum += duration
		}
		avgTime := sum / float64(len(eventProcessingTimes))
		AverageEventProcessingTimeSeconds.Set(avgTime)
	}
}

// Collects worker-related metrics
func collectWorkerMetrics() {
	eventStatsLock.RLock()
	defer eventStatsLock.RUnlock()

	// Update worker uptime for each active worker
	now := time.Now()
	for jobID, startTime := range workerStartTimes {
		uptime := now.Sub(startTime).Seconds()
		WorkerUptimeSeconds.WithLabelValues(jobID).Set(uptime)
	}

	// Update worker memory usage
	for jobID, memUsage := range workerMemoryUsage {
		WorkerMemoryUsageBytes.WithLabelValues(jobID).Set(float64(memUsage))
	}
}

// Resets metrics that should be reset daily
func resetDailyMetrics() {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()

	// Reset daily counters
	for chainID := range eventsLastMinute {
		EventsPerMinute.WithLabelValues(chainID).Set(0)
	}
	AverageEventProcessingTimeSeconds.Set(0)

	// Reset tracking variables
	eventProcessingTimes = nil
	eventsLastMinute = make(map[string]int64)
	totalEvents = 0
	successfulEvents = 0
	totalActions = 0
	successfulActions = 0
	lastMinuteReset = time.Now()
}

// Helper functions to get configuration values
func getDuplicateEventWindowSeconds() float64 {
	return 30.0 // Default to 30 seconds, can be made configurable
}

// HTTP and API tracking functions

// TrackHTTPRequest tracks HTTP request metrics
func TrackHTTPRequest(method, endpoint, statusCode string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// Database tracking functions

// TrackDBRequest tracks database request metrics
func TrackDBRequest(method, endpoint, status string) {
	DBRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// TrackDBConnectionError tracks database connection errors
func TrackDBConnectionError() {
	DBConnectionErrorsTotal.Inc()
}

// TrackDBRetry tracks database retry attempts
func TrackDBRetry(endpoint string) {
	DBRetriesTotal.WithLabelValues(endpoint).Inc()
}

// Blockchain and RPC tracking functions

// TrackChainConnection tracks blockchain connection attempts
func TrackChainConnection(chainID, status string) {
	ChainConnectionsTotal.WithLabelValues(chainID, status).Inc()
}

// TrackRPCRequest tracks RPC requests to blockchain nodes
func TrackRPCRequest(chainID, method, status string) {
	RPCRequestsTotal.WithLabelValues(chainID, method, status).Inc()
}

// TrackConnectionFailure tracks blockchain connection failures
func TrackConnectionFailure(chainID string) {
	ConnectionFailuresTotal.WithLabelValues(chainID).Inc()
}

// Job and Worker tracking functions

// TrackJobScheduled tracks when a job is scheduled
func TrackJobScheduled() {
	JobsScheduled.Inc()
}

// TrackJobCompleted tracks when a job completes
func TrackJobCompleted(status string) {
	JobsCompleted.WithLabelValues(status).Inc()
}

// UpdateActiveWorkers updates the count of active workers
func UpdateActiveWorkers(count int) {
	ActiveWorkers.Set(float64(count))
}

// TrackWorkerStart tracks when a worker starts
func TrackWorkerStart(jobID string) {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()
	workerStartTimes[jobID] = time.Now()
}

// TrackWorkerStop tracks when a worker stops
func TrackWorkerStop(jobID string) {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()
	delete(workerStartTimes, jobID)
	delete(workerMemoryUsage, jobID)
}

// TrackWorkerError tracks worker errors
func TrackWorkerError(jobID, errorType string) {
	WorkerErrorsTotal.WithLabelValues(jobID, errorType).Inc()
}

// UpdateWorkerMemoryUsage updates worker memory usage
func UpdateWorkerMemoryUsage(jobID string, memoryBytes int64) {
	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()
	workerMemoryUsage[jobID] = memoryBytes
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

// Action tracking functions

// TrackActionExecution tracks action executions
func TrackActionExecution(jobID, status string) {
	ActionExecutionsTotal.WithLabelValues(jobID, status).Inc()

	eventStatsLock.Lock()
	defer eventStatsLock.Unlock()
	totalActions++
	if status == "success" {
		successfulActions++
	}
}

// Error and Recovery tracking functions

// TrackTimeout tracks operation timeouts
func TrackTimeout(operation string) {
	TimeoutsTotal.WithLabelValues(operation).Inc()
}

// TrackCriticalError tracks critical system errors
func TrackCriticalError(errorType string) {
	CriticalErrorsTotal.WithLabelValues(errorType).Inc()
}

// TrackRecoveryAttempt tracks automatic recovery attempts
func TrackRecoveryAttempt(component string) {
	RecoveryAttemptsTotal.WithLabelValues(component).Inc()
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

// GetActionStats returns current action statistics
func GetActionStats() (total, successful int64) {
	eventStatsLock.RLock()
	defer eventStatsLock.RUnlock()
	return totalActions, successfulActions
}

// GetWorkerCount returns the current number of active workers
func GetWorkerCount() int {
	eventStatsLock.RLock()
	defer eventStatsLock.RUnlock()
	return len(workerStartTimes)
}

// UpdateEventsPerMinute updates the events per minute metric for a specific chain
func UpdateEventsPerMinute(chainID string, count float64) {
	EventsPerMinute.WithLabelValues(chainID).Set(count)
}

// UpdateAverageEventProcessingTime updates the average event processing time
func UpdateAverageEventProcessingTime(seconds float64) {
	AverageEventProcessingTimeSeconds.Set(seconds)
}

// TrackEventProcessing tracks event processing with timing
func TrackEventProcessing(chainID string, duration time.Duration, success bool) {
	// Track the event
	TrackEvent(chainID, duration)

	// Track success/failure
	if success {
		TrackEventSuccess(chainID)
	}
}
