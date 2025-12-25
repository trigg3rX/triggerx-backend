package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// System Metrics
	uptimeSeconds     observability.Gauge
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// Performance Metrics
	tasksPerMinute       observability.Gauge
	tasksCreatedTotal    observability.Gauge
	tasksDispatchedTotal *observability.CounterVec
	taskBatchSize        observability.Gauge

	// Database Metrics
	dbRequestsTotal       *observability.CounterVec // Read and Write
	dbRequestsErrorsTotal *observability.CounterVec // Read and Write

	// Internal tracking variables for performance calculations
	taskStatsLock   sync.RWMutex
	tasksLastMinute int64
	lastMinuteReset time.Time
)

func init() {
	lastMinuteReset = time.Now()
}

// InitializeMetrics initializes all metrics using the observability metrics instance.
// This must be called before using any metrics.
func InitializeMetrics(obsMetrics observability.Metrics) {
	// System metrics
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.time_scheduler.uptime_seconds",
		observability.WithDescription("Time passed since Time Scheduler started in seconds"),
		observability.WithUnit("s"),
	)

	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.time_scheduler.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.time_scheduler.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.time_scheduler.goroutines_active",
		observability.WithDescription("Number of active goroutines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.time_scheduler.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	// Performance metrics
	tasksPerMinute = obsMetrics.Gauge(
		"triggerx.time_scheduler.tasks_per_minute",
		observability.WithDescription("Task throughput rate"),
		observability.WithUnit("1/m"),
	)

	taskBatchSize = obsMetrics.Gauge(
		"triggerx.time_scheduler.job_batch_size",
		observability.WithDescription("Number of jobs processed per batch"),
	)

	tasksCreatedTotal = obsMetrics.Gauge(
		"triggerx.time_scheduler.tasks_created_total",
		observability.WithDescription("Total number of tasks created"),
	)

	tasksDispatchedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.tasks_dispatched_total",
		[]string{"status"},
		observability.WithDescription("Total number of tasks dispatched to task dispatcher"),
	)

	// Database metrics
	dbRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.db_requests_total",
		[]string{"method", "endpoint", "status"},
		observability.WithDescription("Database client requests"),
	)

	dbRequestsErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.db_requests_errors_total",
		[]string{"method", "endpoint"},
		observability.WithDescription("Database request errors"),
	)
}

// Collector manages metrics collection for the time scheduler.
type Collector struct {
	metrics observability.Metrics
}

// NewCollector creates a new metrics collector.
// The metrics are managed through the observability package, which handles
// Prometheus export internally if configured.
func NewCollector(obsMetrics observability.Metrics) *Collector {
	return &Collector{
		metrics: obsMetrics,
	}
}

// Start starts the background metrics collection goroutines.
// This begins collecting system metrics, performance metrics, and configuration metrics.
func (c *Collector) Start() {
	// Update metrics every 15 seconds
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

// collectSystemMetrics collects system resource metrics
func collectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage (current allocated bytes)
	if memoryUsageBytes != nil {
		memoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}

	// Update CPU usage
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

// collectConfigurationMetrics collects configuration-based metrics
func collectConfigurationMetrics() {
	// Set job batch size from configuration
	if taskBatchSize != nil {
		taskBatchSize.Set(ctx, float64(config.GetTaskBatchSize()))
	}
}

// collectPerformanceMetrics collects performance-related metrics
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
}

// resetDailyMetrics resets metrics that should be reset daily
func resetDailyMetrics() {
	taskStatsLock.Lock()
	defer taskStatsLock.Unlock()

	// Reset tasks per minute
	if tasksPerMinute != nil {
		tasksPerMinute.Set(ctx, 0)
	}

	// Reset tracking variables
	tasksLastMinute = 0
	lastMinuteReset = time.Now()
}

// Tracking functions

// TrackDBRequest tracks database request metrics
func TrackDBRequest(method, endpoint, status string) {
	if dbRequestsTotal != nil {
		dbRequestsTotal.WithLabelValues(method, endpoint, status).Inc(ctx)
	}
}

// TrackDBRequestError tracks database request errors
func TrackDBRequestError(method, endpoint string) {
	if dbRequestsErrorsTotal != nil {
		dbRequestsErrorsTotal.WithLabelValues(method, endpoint).Inc(ctx)
	}
}

// TrackDBConnectionError tracks database connection errors (legacy function name)
func TrackDBConnectionError() {
	// Track as a database request error
	TrackDBRequestError("connection", "database")
}

// UpdateTasksCreated updates the total number of tasks created
func UpdateTasksCreated(count float64) {
	if tasksCreatedTotal != nil {
		tasksCreatedTotal.Set(ctx, count)
	}
}

// UpdateTasksDispatched updates the number of tasks dispatched with status
func UpdateTasksDispatched(status string) {
	if tasksDispatchedTotal != nil {
		tasksDispatchedTotal.WithLabelValues(status).Inc(ctx)
	}
}

// UpdateTasksPerMinute updates the tasks per minute metric
func UpdateTasksPerMinute(count float64) {
	if tasksPerMinute != nil {
		tasksPerMinute.Set(ctx, count)
	}
	taskStatsLock.Lock()
	tasksLastMinute = int64(count)
	taskStatsLock.Unlock()
}

// UpdateTaskBatchSize updates the task batch size metric
func UpdateTaskBatchSize(size float64) {
	if taskBatchSize != nil {
		taskBatchSize.Set(ctx, size)
	}
}

// TrackTaskExpired tracks expired tasks (increments tasks created but not dispatched)
func TrackTaskExpired() {
	// Expired tasks are created but not dispatched, so we track them separately
	// This could be added as a label to tasksDispatchedTotal if needed
}

// TrackTaskByScheduleType is deprecated - schedule type tracking removed
// Kept for backward compatibility but does nothing
func TrackTaskByScheduleType(scheduleType string) {
	// No longer tracking schedule types
}

// TrackTaskBroadcast tracks task broadcasts to task dispatcher
func TrackTaskBroadcast(status string) {
	UpdateTasksDispatched(status)
}

// TrackTaskCompletion is a legacy function - use UpdateTasksDispatched instead
func TrackTaskCompletion(success bool, duration time.Duration) {
	status := "failed"
	if success {
		status = "success"
	}
	UpdateTasksDispatched(status)
}

// UpdateTasksScheduled is a legacy function - use UpdateTasksCreated instead
func UpdateTasksScheduled(count float64) {
	UpdateTasksCreated(count)
}
