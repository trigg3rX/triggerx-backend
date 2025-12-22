package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
)

var (
	// Internal tracking variables for performance calculations
	taskStatsLock       sync.RWMutex
	taskCompletionTimes []float64
	successfulTasks     int64
	totalTasks          int64
	tasksLastMinute     int64
	lastMinuteReset     time.Time
)

func init() {
	lastMinuteReset = time.Now()
}

// Starts collecting metrics
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
	// Set job batch size from configuration
	if taskBatchSize != nil {
		taskBatchSize.Set(ctx, float64(getTaskBatchSize()))
	}

	// Set duplicate task window from configuration
	if duplicateTaskWindowSeconds != nil {
		duplicateTaskWindowSeconds.Set(ctx, getDuplicateTaskWindowSeconds())
	}
}

// Collects performance-related metrics
func collectPerformanceMetrics() {
	taskStatsLock.RLock()
	defer taskStatsLock.RUnlock()

	// Update tasks per minute
	if time.Since(lastMinuteReset) >= time.Minute {
		if tasksPerMinute != nil {
			tasksPerMinute.Set(ctx, float64(tasksLastMinute))
		}
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
		if taskSuccessRatePercent != nil {
			taskSuccessRatePercent.Set(ctx, successRate)
		}
	}

	// Calculate average task completion time
	if len(taskCompletionTimes) > 0 {
		var sum float64
		for _, duration := range taskCompletionTimes {
			sum += duration
		}
		avgTime := sum / float64(len(taskCompletionTimes))
		if averageTaskCompletionTimeSeconds != nil {
			averageTaskCompletionTimeSeconds.Set(ctx, avgTime)
		}
	}
}

// Resets metrics that should be reset daily
func resetDailyMetrics() {
	taskStatsLock.Lock()
	defer taskStatsLock.Unlock()

	// Reset daily counters
	if tasksPerMinute != nil {
		tasksPerMinute.Set(ctx, 0)
	}
	if taskSuccessRatePercent != nil {
		taskSuccessRatePercent.Set(ctx, 0)
	}
	if averageTaskCompletionTimeSeconds != nil {
		averageTaskCompletionTimeSeconds.Set(ctx, 0)
	}

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
	if httpRequestsTotal != nil {
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc(ctx)
	}
}

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

// TrackTaskBroadcast tracks task broadcasts to performers
func TrackTaskBroadcast(status string) {
	if taskBroadcastsTotal != nil {
		taskBroadcastsTotal.WithLabelValues(status).Inc(ctx)
	}
}

// TrackTaskByScheduleType tracks tasks by their schedule type
func TrackTaskByScheduleType(scheduleType string) {
	if tasksByScheduleTypeTotal != nil {
		tasksByScheduleTypeTotal.WithLabelValues(scheduleType).Inc(ctx)
	}
}

// TrackTaskExpired tracks expired tasks
func TrackTaskExpired() {
	if tasksExpiredTotal != nil {
		tasksExpiredTotal.Inc(ctx)
	}
}

// UpdateTasksPerMinute updates the tasks per minute metric
func UpdateTasksPerMinute(count float64) {
	if tasksPerMinute != nil {
		tasksPerMinute.Set(ctx, count)
	}
}

// UpdateAverageTaskCompletionTime updates the average task completion time
func UpdateAverageTaskCompletionTime(seconds float64) {
	if averageTaskCompletionTimeSeconds != nil {
		averageTaskCompletionTimeSeconds.Set(ctx, seconds)
	}
}

// UpdateTaskSuccessRate updates the task success rate percentage
func UpdateTaskSuccessRate(percentage float64) {
	if taskSuccessRatePercent != nil {
		taskSuccessRatePercent.Set(ctx, percentage)
	}
}

// UpdateTasksScheduled updates the number of tasks scheduled
func UpdateTasksScheduled(count float64) {
	if tasksScheduled != nil {
		tasksScheduled.Set(ctx, count)
	}
}

// UpdateTaskBatchSize updates the task batch size metric
func UpdateTaskBatchSize(size float64) {
	if taskBatchSize != nil {
		taskBatchSize.Set(ctx, size)
	}
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
	if taskExecutionTime != nil {
		taskExecutionTime.Record(ctx, duration)
	}
}

// TrackTaskCompletion tracks when a task completes (wrapper for scheduler)
func TrackTaskCompletion(success bool, duration time.Duration) {
	status := "failed"
	if success {
		status = "success"
	}

	if tasksCompleted != nil {
		tasksCompleted.WithLabelValues(status).Inc(ctx)
	}
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
