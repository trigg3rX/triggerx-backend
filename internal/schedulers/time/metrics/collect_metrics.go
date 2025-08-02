package metrics

import (
	"runtime"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
)

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
	// Set job batch size from configuration
	TaskBatchSize.Set(float64(getTaskBatchSize()))

	// Set duplicate task window from configuration
	DuplicateTaskWindowSeconds.Set(getDuplicateTaskWindowSeconds())
}

// Collects performance-related metrics
func collectPerformanceMetrics() {
	taskStatsLock.RLock()
	defer taskStatsLock.RUnlock()

	// Update tasks per minute
	if time.Since(lastMinuteReset) >= time.Minute {
		TasksPerMinute.Set(float64(tasksLastMinute))
		// Reset for next minute in a separate goroutine to avoid blocking
		go func() {
			taskStatsLock.Lock()
			tasksLastMinute = 0
			lastMinuteReset = time.Now()
			taskStatsLock.Unlock()
		}()
	}

	// Calculate success rate
	if totalTasks > 0 {
		successRate := (float64(successfulTasks) / float64(totalTasks)) * 100
		TaskSuccessRatePercent.Set(successRate)
	}

	// Calculate average task completion time
	if len(taskCompletionTimes) > 0 {
		var sum float64
		for _, duration := range taskCompletionTimes {
			sum += duration
		}
		avgTime := sum / float64(len(taskCompletionTimes))
		AverageTaskCompletionTimeSeconds.Set(avgTime)
	}
}

// Resets metrics that should be reset daily
func resetDailyMetrics() {
	taskStatsLock.Lock()
	defer taskStatsLock.Unlock()

	// Reset daily counters
	TasksPerMinute.Set(0)
	TaskSuccessRatePercent.Set(0)
	AverageTaskCompletionTimeSeconds.Set(0)

	// Reset tracking variables
	taskCompletionTimes = nil
	successfulTasks = 0
	totalTasks = 0
	tasksLastMinute = 0
	lastMinuteReset = time.Now()
}

// Helper functions to get configuration values
func getTaskBatchSize() int {
	return config.GetTaskBatchSize()
}

func getDuplicateTaskWindowSeconds() float64 {
	return config.GetDuplicateTaskWindow().Seconds()
}

// HTTP Middleware and tracking functions

// TrackHTTPRequest tracks HTTP request metrics
func TrackHTTPRequest(method, endpoint, statusCode string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

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

// TrackTaskBroadcast tracks task broadcasts to performers
func TrackTaskBroadcast(status string) {
	TaskBroadcastsTotal.WithLabelValues(status).Inc()
}

// TrackTaskByScheduleType tracks tasks by their schedule type
func TrackTaskByScheduleType(scheduleType string) {
	TasksByScheduleTypeTotal.WithLabelValues(scheduleType).Inc()
}

// TrackTaskExpired tracks expired tasks
func TrackTaskExpired() {
	TasksExpiredTotal.Inc()
}

// UpdateTasksPerMinute updates the tasks per minute metric
func UpdateTasksPerMinute(count float64) {
	TasksPerMinute.Set(count)
}

// UpdateAverageTaskCompletionTime updates the average task completion time
func UpdateAverageTaskCompletionTime(seconds float64) {
	AverageTaskCompletionTimeSeconds.Set(seconds)
}

// UpdateTaskSuccessRate updates the task success rate percentage
func UpdateTaskSuccessRate(percentage float64) {
	TaskSuccessRatePercent.Set(percentage)
}

// TrackTaskExecution tracks task execution with timing (use this when a task starts executing)
func TrackTaskExecution(duration float64, success bool) {
	taskStatsLock.Lock()
	defer taskStatsLock.Unlock()

	// Update total task count
	totalTasks++
	tasksLastMinute++

	// Track success/failure
	if success {
		successfulTasks++
	}

	// Track completion time (keep last 1000 entries to avoid memory growth)
	taskCompletionTimes = append(taskCompletionTimes, duration)
	if len(taskCompletionTimes) > 1000 {
		taskCompletionTimes = taskCompletionTimes[1:]
	}

	// Observe execution time in histogram
	TaskExecutionTime.Observe(duration)
}

// TrackTaskCompletion tracks when a task completes (wrapper for scheduler)
func TrackTaskCompletion(success bool, duration time.Duration) {
	status := "failed"
	if success {
		status = "success"
	}

	TasksCompleted.WithLabelValues(status).Inc()
	TrackTaskExecution(duration.Seconds(), success)
}

// GetTaskStats returns current task statistics (for debugging/monitoring)
func GetTaskStats() (total, successful int64, avgTime float64) {
	taskStatsLock.RLock()
	defer taskStatsLock.RUnlock()

	total = totalTasks
	successful = successfulTasks

	if len(taskCompletionTimes) > 0 {
		var sum float64
		for _, duration := range taskCompletionTimes {
			sum += duration
		}
		avgTime = sum / float64(len(taskCompletionTimes))
	}

	return
}
