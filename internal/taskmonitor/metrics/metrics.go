package metrics

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// System metrics
	uptimeSeconds     observability.Gauge
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// Service Health & Availability
	serviceStatus           *observability.GaugeVec
	isRedisUpstashAvailable observability.Gauge

	// Connection Management
	clientConnectionsTotal      *observability.CounterVec
	clientConnectionErrorsTotal *observability.CounterVec
	pingOperationsTotal         *observability.CounterVec
	pingDuration                observability.Histogram
	connectionChecksTotal       *observability.CounterVec

	// Core Stream Operations
	taskStreamLengths        *observability.GaugeVec
	jobStreamLengths         *observability.GaugeVec
	tasksAddedToStreamTotal  *observability.CounterVec
	tasksReadFromStreamTotal *observability.CounterVec
	jobsAddedToStreamTotal   *observability.CounterVec
	jobsReadFromStreamTotal  *observability.CounterVec

	// Task Lifecycle & Performance
	taskRetryOperationsTotal        observability.Counter
	taskMaxRetriesExceededTotal     observability.Counter
	tasksMovedToFailedStreamTotal   observability.Counter
	taskReadyToProcessingTotal      observability.Counter
	taskProcessingToCompletedTotal  observability.Counter
	taskLifecycleTransitionDuration *observability.HistogramVec

	// Redis Client Operation Metrics
	redisOperationsTotal      *observability.CounterVec
	redisOperationDuration    *observability.HistogramVec
	redisRetryAttempts        *observability.CounterVec
	redisConnectionRecoveries *observability.CounterVec
	redisConnectionHealth     *observability.GaugeVec
)

// Exported variables for use in other packages
var (
	UptimeSeconds                   observability.Gauge
	MemoryUsageBytes                observability.Gauge
	CPUUsagePercent                 observability.Gauge
	GoroutinesActive                observability.Gauge
	GCDurationSeconds               observability.Gauge
	ServiceStatus                   *observability.GaugeVec
	IsRedisUpstashAvailable         observability.Gauge
	ClientConnectionsTotal          *observability.CounterVec
	ClientConnectionErrorsTotal     *observability.CounterVec
	PingOperationsTotal             *observability.CounterVec
	PingDuration                    observability.Histogram
	ConnectionChecksTotal           *observability.CounterVec
	TaskStreamLengths               *observability.GaugeVec
	JobStreamLengths                *observability.GaugeVec
	TasksAddedToStreamTotal         *observability.CounterVec
	TasksReadFromStreamTotal        *observability.CounterVec
	JobsAddedToStreamTotal          *observability.CounterVec
	JobsReadFromStreamTotal         *observability.CounterVec
	TaskRetryOperationsTotal        observability.Counter
	TaskMaxRetriesExceededTotal     observability.Counter
	TasksMovedToFailedStreamTotal   observability.Counter
	TaskReadyToProcessingTotal      observability.Counter
	TaskProcessingToCompletedTotal  observability.Counter
	TaskLifecycleTransitionDuration *observability.HistogramVec
	RedisOperationsTotal            *observability.CounterVec
	RedisOperationDuration          *observability.HistogramVec
	RedisRetryAttempts              *observability.CounterVec
	RedisConnectionRecoveries       *observability.CounterVec
	RedisConnectionHealth           *observability.GaugeVec
)

// InitializeMetrics initializes all metrics using the observability metrics instance
func InitializeMetrics(obsMetrics observability.Metrics) {
	// System metrics
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.taskmonitor.uptime_seconds",
		observability.WithDescription("Time passed since Task Monitor Service started in seconds"),
		observability.WithUnit("s"),
	)
	UptimeSeconds = uptimeSeconds

	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.taskmonitor.memory_usage_bytes",
		observability.WithDescription("Service memory consumption"),
		observability.WithUnit("By"),
	)
	MemoryUsageBytes = memoryUsageBytes

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.taskmonitor.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)
	CPUUsagePercent = cpuUsagePercent

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.taskmonitor.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)
	GoroutinesActive = goroutinesActive

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.taskmonitor.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)
	GCDurationSeconds = gcDurationSeconds

	// Service Health & Availability
	serviceStatus = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskmonitor.service_status",
		[]string{"component"},
		observability.WithDescription("Service component health status"),
	)
	ServiceStatus = serviceStatus

	isRedisUpstashAvailable = obsMetrics.Gauge(
		"triggerx.taskmonitor.is_upstash_available",
		observability.WithDescription("Whether Upstash Redis is available and being used (1=Upstash, 0=Local)"),
	)
	IsRedisUpstashAvailable = isRedisUpstashAvailable

	// Connection Management
	clientConnectionsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.client_connections_total",
		[]string{"status"},
		observability.WithDescription("Redis client connections"),
	)
	ClientConnectionsTotal = clientConnectionsTotal

	clientConnectionErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.client_connection_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Redis client connection errors"),
	)
	ClientConnectionErrorsTotal = clientConnectionErrorsTotal

	pingOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.ping_operations_total",
		[]string{"status"},
		observability.WithDescription("Redis ping operations"),
	)
	PingOperationsTotal = pingOperationsTotal

	pingDuration = obsMetrics.Histogram(
		"triggerx.taskmonitor.ping_duration_seconds",
		observability.WithDescription("Redis ping response time"),
		observability.WithUnit("s"),
	)
	PingDuration = pingDuration

	connectionChecksTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.connection_checks_total",
		[]string{"status"},
		observability.WithDescription("Connection health checks"),
	)
	ConnectionChecksTotal = connectionChecksTotal

	// Core Stream Operations
	taskStreamLengths = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskmonitor.task_stream_lengths",
		[]string{"stream"},
		observability.WithDescription("Current task stream lengths"),
	)
	TaskStreamLengths = taskStreamLengths

	jobStreamLengths = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskmonitor.job_stream_lengths",
		[]string{"stream"},
		observability.WithDescription("Current job stream lengths"),
	)
	JobStreamLengths = jobStreamLengths

	tasksAddedToStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.tasks_added_to_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Tasks added to streams"),
	)
	TasksAddedToStreamTotal = tasksAddedToStreamTotal

	tasksReadFromStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.tasks_read_from_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Tasks read from streams"),
	)
	TasksReadFromStreamTotal = tasksReadFromStreamTotal

	jobsAddedToStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.jobs_added_to_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Jobs added to streams"),
	)
	JobsAddedToStreamTotal = jobsAddedToStreamTotal

	jobsReadFromStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.jobs_read_from_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Jobs read from streams"),
	)
	JobsReadFromStreamTotal = jobsReadFromStreamTotal

	// Task Lifecycle & Performance
	taskRetryOperationsTotal = obsMetrics.Counter(
		"triggerx.taskmonitor.task_retry_operations_total",
		observability.WithDescription("Task retry operations"),
	)
	TaskRetryOperationsTotal = taskRetryOperationsTotal

	taskMaxRetriesExceededTotal = obsMetrics.Counter(
		"triggerx.taskmonitor.task_max_retries_exceeded_total",
		observability.WithDescription("Tasks exceeding max retry attempts"),
	)
	TaskMaxRetriesExceededTotal = taskMaxRetriesExceededTotal

	tasksMovedToFailedStreamTotal = obsMetrics.Counter(
		"triggerx.taskmonitor.tasks_moved_to_failed_stream_total",
		observability.WithDescription("Tasks permanently failed and moved to failed stream"),
	)
	TasksMovedToFailedStreamTotal = tasksMovedToFailedStreamTotal

	taskReadyToProcessingTotal = obsMetrics.Counter(
		"triggerx.taskmonitor.task_ready_to_processing_total",
		observability.WithDescription("Tasks moved from ready to processing stream"),
	)
	TaskReadyToProcessingTotal = taskReadyToProcessingTotal

	taskProcessingToCompletedTotal = obsMetrics.Counter(
		"triggerx.taskmonitor.task_processing_to_completed_total",
		observability.WithDescription("Tasks moved from processing to completed stream"),
	)
	TaskProcessingToCompletedTotal = taskProcessingToCompletedTotal

	taskLifecycleTransitionDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.taskmonitor.task_lifecycle_transition_duration_seconds",
		[]string{"from_stream", "to_stream"},
		observability.WithDescription("Task lifecycle transition time"),
		observability.WithUnit("s"),
	)
	TaskLifecycleTransitionDuration = taskLifecycleTransitionDuration

	// Redis Client Operation Metrics
	redisOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.redis_operations_total",
		[]string{"operation", "status"},
		observability.WithDescription("Total Redis operations performed"),
	)
	RedisOperationsTotal = redisOperationsTotal

	redisOperationDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.taskmonitor.redis_operation_duration_seconds",
		[]string{"operation"},
		observability.WithDescription("Redis operation duration"),
		observability.WithUnit("s"),
	)
	RedisOperationDuration = redisOperationDuration

	redisRetryAttempts = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.redis_retry_attempts_total",
		[]string{"operation"},
		observability.WithDescription("Total Redis retry attempts"),
	)
	RedisRetryAttempts = redisRetryAttempts

	redisConnectionRecoveries = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskmonitor.redis_connection_recoveries_total",
		[]string{"status"},
		observability.WithDescription("Total Redis connection recovery attempts"),
	)
	RedisConnectionRecoveries = redisConnectionRecoveries

	redisConnectionHealth = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskmonitor.redis_connection_health",
		[]string{"type"},
		observability.WithDescription("Redis connection health status"),
	)
	RedisConnectionHealth = redisConnectionHealth
}
