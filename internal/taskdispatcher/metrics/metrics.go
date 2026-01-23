package metrics

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
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

	// HTTP Dispatch to Keepers
	tasksDispatchedToKeeperTotal *observability.CounterVec
	keeperDispatchDuration       *observability.HistogramVec
)

// Export variables for use in other packages
var (
	ServiceStatus                   = serviceStatus
	IsRedisUpstashAvailable         = isRedisUpstashAvailable
	ClientConnectionsTotal          = clientConnectionsTotal
	ClientConnectionErrorsTotal     = clientConnectionErrorsTotal
	PingOperationsTotal             = pingOperationsTotal
	ConnectionChecksTotal           = connectionChecksTotal
	TaskStreamLengths               = taskStreamLengths
	JobStreamLengths                = jobStreamLengths
	TasksAddedToStreamTotal         = tasksAddedToStreamTotal
	TasksReadFromStreamTotal        = tasksReadFromStreamTotal
	JobsAddedToStreamTotal          = jobsAddedToStreamTotal
	JobsReadFromStreamTotal         = jobsReadFromStreamTotal
	TaskRetryOperationsTotal        = taskRetryOperationsTotal
	TaskMaxRetriesExceededTotal     = taskMaxRetriesExceededTotal
	TasksMovedToFailedStreamTotal   = tasksMovedToFailedStreamTotal
	TaskReadyToProcessingTotal      = taskReadyToProcessingTotal
	TaskProcessingToCompletedTotal  = taskProcessingToCompletedTotal
	TaskLifecycleTransitionDuration = taskLifecycleTransitionDuration
	TasksDispatchedToKeeperTotal    = tasksDispatchedToKeeperTotal
	KeeperDispatchDuration          = keeperDispatchDuration
)

// InitializeMetrics initializes all metrics using the observability metrics instance
func InitializeMetrics(obsMetrics observability.Metrics) {
	// System metrics
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.taskdispatcher.uptime_seconds",
		observability.WithDescription("Time passed since Task Dispatcher Service started in seconds"),
		observability.WithUnit("s"),
	)

	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.taskdispatcher.memory_usage_bytes",
		observability.WithDescription("Service memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.taskdispatcher.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.taskdispatcher.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.taskdispatcher.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	// Service Health & Availability
	serviceStatus = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskdispatcher.service_status",
		[]string{"component"},
		observability.WithDescription("Service component health status"),
	)

	isRedisUpstashAvailable = obsMetrics.Gauge(
		"triggerx.taskdispatcher.is_upstash_available",
		observability.WithDescription("Whether Upstash Redis is available and being used (1=Upstash, 0=Local)"),
	)

	// Connection Management
	clientConnectionsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.client_connections_total",
		[]string{"status"},
		observability.WithDescription("Redis client connections"),
	)

	clientConnectionErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.client_connection_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Redis client connection errors"),
	)

	pingOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.ping_operations_total",
		[]string{"status"},
		observability.WithDescription("Redis ping operations"),
	)

	pingDuration = obsMetrics.Histogram(
		"triggerx.taskdispatcher.ping_duration_seconds",
		observability.WithDescription("Redis ping response time"),
		observability.WithUnit("s"),
	)

	connectionChecksTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.connection_checks_total",
		[]string{"status"},
		observability.WithDescription("Connection health checks"),
	)

	// Core Stream Operations
	taskStreamLengths = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskdispatcher.task_stream_lengths",
		[]string{"stream"},
		observability.WithDescription("Current task stream lengths"),
	)

	jobStreamLengths = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskdispatcher.job_stream_lengths",
		[]string{"stream"},
		observability.WithDescription("Current job stream lengths"),
	)

	tasksAddedToStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.tasks_added_to_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Tasks added to streams"),
	)

	tasksReadFromStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.tasks_read_from_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Tasks read from streams"),
	)

	jobsAddedToStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.jobs_added_to_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Jobs added to streams"),
	)

	jobsReadFromStreamTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.jobs_read_from_stream_total",
		[]string{"stream", "status"},
		observability.WithDescription("Jobs read from streams"),
	)

	// Task Lifecycle & Performance
	taskRetryOperationsTotal = obsMetrics.Counter(
		"triggerx.taskdispatcher.task_retry_operations_total",
		observability.WithDescription("Task retry operations"),
	)

	taskMaxRetriesExceededTotal = obsMetrics.Counter(
		"triggerx.taskdispatcher.task_max_retries_exceeded_total",
		observability.WithDescription("Tasks exceeding max retry attempts"),
	)

	tasksMovedToFailedStreamTotal = obsMetrics.Counter(
		"triggerx.taskdispatcher.tasks_moved_to_failed_stream_total",
		observability.WithDescription("Tasks permanently failed and moved to failed stream"),
	)

	taskReadyToProcessingTotal = obsMetrics.Counter(
		"triggerx.taskdispatcher.task_ready_to_processing_total",
		observability.WithDescription("Tasks moved from ready to processing stream"),
	)

	taskProcessingToCompletedTotal = obsMetrics.Counter(
		"triggerx.taskdispatcher.task_processing_to_completed_total",
		observability.WithDescription("Tasks moved from processing to completed stream"),
	)

	taskLifecycleTransitionDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.taskdispatcher.task_lifecycle_transition_duration_seconds",
		[]string{"from_stream", "to_stream"},
		observability.WithDescription("Task lifecycle transition time"),
		observability.WithUnit("s"),
	)

	// Redis Client Operation Metrics
	redisOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.redis_operations_total",
		[]string{"operation", "status"},
		observability.WithDescription("Total Redis operations performed"),
	)

	redisOperationDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.taskdispatcher.redis_operation_duration_seconds",
		[]string{"operation"},
		observability.WithDescription("Redis operation duration"),
		observability.WithUnit("s"),
	)

	redisRetryAttempts = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.redis_retry_attempts_total",
		[]string{"operation"},
		observability.WithDescription("Total Redis retry attempts"),
	)

	redisConnectionRecoveries = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.redis_connection_recoveries_total",
		[]string{"status"},
		observability.WithDescription("Total Redis connection recovery attempts"),
	)

	redisConnectionHealth = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.taskdispatcher.redis_connection_health",
		[]string{"type"},
		observability.WithDescription("Redis connection health status"),
	)

	// HTTP Dispatch to Keepers
	tasksDispatchedToKeeperTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.taskdispatcher.tasks_dispatched_to_keeper_total",
		[]string{"status", "network"},
		observability.WithDescription("Total tasks dispatched to keepers via HTTP"),
	)

	keeperDispatchDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.taskdispatcher.keeper_dispatch_duration_seconds",
		[]string{"status", "network"},
		observability.WithDescription("HTTP dispatch latency to keepers"),
		observability.WithUnit("s"),
	)

	// Update exported variables
	ServiceStatus = serviceStatus
	IsRedisUpstashAvailable = isRedisUpstashAvailable
	ClientConnectionsTotal = clientConnectionsTotal
	ClientConnectionErrorsTotal = clientConnectionErrorsTotal
	PingOperationsTotal = pingOperationsTotal
	ConnectionChecksTotal = connectionChecksTotal
	TaskStreamLengths = taskStreamLengths
	JobStreamLengths = jobStreamLengths
	TasksAddedToStreamTotal = tasksAddedToStreamTotal
	TasksReadFromStreamTotal = tasksReadFromStreamTotal
	JobsAddedToStreamTotal = jobsAddedToStreamTotal
	JobsReadFromStreamTotal = jobsReadFromStreamTotal
	TaskRetryOperationsTotal = taskRetryOperationsTotal
	TaskMaxRetriesExceededTotal = taskMaxRetriesExceededTotal
	TasksMovedToFailedStreamTotal = tasksMovedToFailedStreamTotal
	TaskReadyToProcessingTotal = taskReadyToProcessingTotal
	TaskProcessingToCompletedTotal = taskProcessingToCompletedTotal
	TaskLifecycleTransitionDuration = taskLifecycleTransitionDuration
	TasksDispatchedToKeeperTotal = tasksDispatchedToKeeperTotal
	KeeperDispatchDuration = keeperDispatchDuration
}

// TrackKeeperDispatch tracks a task dispatch to keeper
func TrackKeeperDispatch(ctx context.Context, success bool, network types.KeeperNetwork, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}

	if TasksDispatchedToKeeperTotal != nil {
		TasksDispatchedToKeeperTotal.WithLabelValues(status, string(network)).Inc(ctx)
	}
	if KeeperDispatchDuration != nil {
		KeeperDispatchDuration.WithLabelValues(status, string(network)).Record(ctx, duration.Seconds())
	}
}
