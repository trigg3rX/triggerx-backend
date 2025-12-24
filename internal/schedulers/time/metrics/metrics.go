package metrics

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// Metrics instances
	uptimeSeconds                    observability.Gauge
	memoryUsageBytes                 observability.Gauge
	cpuUsagePercent                  observability.Gauge
	goroutinesActive                 observability.Gauge
	gcDurationSeconds                observability.Gauge
	tasksPerMinute                   observability.Gauge
	averageTaskCompletionTimeSeconds observability.Gauge
	taskSuccessRatePercent           observability.Gauge
	tasksScheduled                   observability.Gauge
	tasksCompleted                   *observability.CounterVec
	taskExecutionTime                observability.Histogram
	tasksExpiredTotal                observability.Counter
	taskBatchSize                    observability.Gauge
	dbRequestsTotal                  *observability.CounterVec
	httpRequestsTotal                *observability.CounterVec
	dbConnectionErrorsTotal          observability.Counter
	dbRetriesTotal                   *observability.CounterVec
	taskBroadcastsTotal              *observability.CounterVec
	tasksByScheduleTypeTotal         *observability.CounterVec
	duplicateTaskWindowSeconds       observability.Gauge
)

// InitializeMetrics initializes all metrics using the observability metrics instance
// This must be called before using any metrics
func InitializeMetrics(obsMetrics observability.Metrics) {
	// Simple gauges
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

	tasksPerMinute = obsMetrics.Gauge(
		"triggerx.time_scheduler.tasks_per_minute",
		observability.WithDescription("Task throughput rate"),
		observability.WithUnit("1/m"),
	)

	averageTaskCompletionTimeSeconds = obsMetrics.Gauge(
		"triggerx.time_scheduler.average_task_completion_time_seconds",
		observability.WithDescription("Mean task completion time"),
		observability.WithUnit("s"),
	)

	taskSuccessRatePercent = obsMetrics.Gauge(
		"triggerx.time_scheduler.task_success_rate_percent",
		observability.WithDescription("Overall task success percentage"),
		observability.WithUnit("%"),
	)

	tasksScheduled = obsMetrics.Gauge(
		"triggerx.time_scheduler.scheduler_tasks_scheduled",
		observability.WithDescription("Total number of jobs currently scheduled"),
	)

	tasksCompleted = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.scheduler_tasks_completed",
		[]string{"status"},
		observability.WithDescription("Total number of jobs completed (success/fail)"),
	)

	taskExecutionTime = obsMetrics.Histogram(
		"triggerx.time_scheduler.scheduler_task_execution_time",
		observability.WithDescription("Time taken to execute a task in seconds"),
		observability.WithUnit("s"),
	)

	tasksExpiredTotal = obsMetrics.Counter(
		"triggerx.time_scheduler.tasks_expired_total",
		observability.WithDescription("Tasks that expired before execution"),
	)

	taskBatchSize = obsMetrics.Gauge(
		"triggerx.time_scheduler.job_batch_size",
		observability.WithDescription("Number of jobs processed per batch"),
	)

	dbRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.db_requests_total",
		[]string{"method", "endpoint", "status"},
		observability.WithDescription("Database client requests"),
	)

	httpRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.http_requests_total",
		[]string{"method", "endpoint", "status_code"},
		observability.WithDescription("HTTP API requests received"),
	)

	dbConnectionErrorsTotal = obsMetrics.Counter(
		"triggerx.time_scheduler.db_connection_errors_total",
		observability.WithDescription("Database connection failures"),
	)

	dbRetriesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.db_retries_total",
		[]string{"endpoint"},
		observability.WithDescription("Database request retry attempts"),
	)

	taskBroadcastsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.task_broadcasts_total",
		[]string{"status"},
		observability.WithDescription("Task broadcasts to performers"),
	)

	tasksByScheduleTypeTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.time_scheduler.tasks_by_schedule_type_total",
		[]string{"type"},
		observability.WithDescription("Tasks processed by schedule type"),
	)

	duplicateTaskWindowSeconds = obsMetrics.Gauge(
		"triggerx.time_scheduler.duplicate_task_window_seconds",
		observability.WithDescription("Duplicate task detection window"),
		observability.WithUnit("s"),
	)
}
