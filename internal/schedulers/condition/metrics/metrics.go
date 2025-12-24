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
	uptimeSeconds               observability.Gauge
	memoryUsageBytes            observability.Gauge
	cpuUsagePercent             observability.Gauge
	goroutinesActive            observability.Gauge
	gcDurationSeconds           observability.Gauge
	eventsPerMinute             *observability.GaugeVec
	jobsScheduled               observability.Counter
	jobsCompleted               *observability.CounterVec
	activeWorkers               observability.Gauge
	chainConnectionsTotal       *observability.CounterVec
	rpcRequestsTotal            *observability.CounterVec
	conditionEvaluationDuration observability.Histogram
	conditionsByTypeTotal       *observability.CounterVec
	conditionsBySourceTotal     *observability.CounterVec
	apiResponseStatusTotal      *observability.CounterVec
	valueParsingErrorsTotal     *observability.CounterVec
	dbRequestsTotal             *observability.CounterVec
	dbConnectionErrorsTotal     observability.Counter
	dbRetriesTotal              *observability.CounterVec
	// actionExecutionsTotal             *observability.CounterVec
	actionExecutionDuration         observability.Histogram
	workerUptimeSeconds             observability.Gauge
	workerErrorsTotal               *observability.CounterVec
	workerMemoryUsageBytes          observability.Gauge
	httpRequestsTotal               *observability.CounterVec
	httpClientConnectionErrorsTotal observability.Counter
	duplicateConditionWindowSeconds observability.Gauge
	// duplicateEventWindowSeconds       observability.Gauge
	averageConditionCheckTimeSeconds  observability.Gauge
	connectionFailuresTotal           *observability.CounterVec
	averageEventProcessingTimeSeconds observability.Gauge
	timeoutsTotal                     *observability.CounterVec
	criticalErrorsTotal               *observability.CounterVec
	invalidValuesTotal                *observability.CounterVec
)

// InitializeMetrics initializes all metrics using the observability metrics instance
// This must be called before using any metrics
func InitializeMetrics(obsMetrics observability.Metrics) {
	// Simple gauges
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.uptime_seconds",
		observability.WithDescription("Time passed since Condition Scheduler started in seconds"),
		observability.WithUnit("s"),
	)

	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.condition_scheduler.memory_usage_bytes",
		observability.WithDescription("Total memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.condition_scheduler.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.condition_scheduler.goroutines_active",
		observability.WithDescription("Number of active goroutines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	// Labeled gauges
	eventsPerMinute = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.condition_scheduler.events_per_minute",
		[]string{"chain_id"},
		observability.WithDescription("Event detection rate per chain"),
	)

	// Counters
	jobsScheduled = obsMetrics.Counter(
		"triggerx.condition_scheduler.jobs_scheduled_total",
		observability.WithDescription("Total number of jobs scheduled"),
	)

	jobsCompleted = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.jobs_completed_total",
		[]string{"status"},
		observability.WithDescription("Total number of jobs completed successfully or failed"),
	)

	activeWorkers = obsMetrics.Gauge(
		"triggerx.condition_scheduler.active_workers",
		observability.WithDescription("Number of active job workers currently running"),
	)

	chainConnectionsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.chain_connections_total",
		[]string{"chain_id", "status"},
		observability.WithDescription("Blockchain connection attempts"),
	)

	rpcRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.rpc_requests_total",
		[]string{"chain_id", "method", "status"},
		observability.WithDescription("RPC requests to blockchain nodes"),
	)

	conditionEvaluationDuration = obsMetrics.Histogram(
		"triggerx.condition_scheduler.condition_evaluation_duration_seconds",
		observability.WithDescription("Time taken to evaluate conditions"),
		observability.WithUnit("s"),
	)

	conditionsByTypeTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.conditions_by_type_total",
		[]string{"condition_type"},
		observability.WithDescription("Conditions monitored by type"),
	)

	conditionsBySourceTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.conditions_by_source_total",
		[]string{"source_type"},
		observability.WithDescription("Conditions monitored by source type"),
	)

	apiResponseStatusTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.api_response_status_total",
		[]string{"status_code"},
		observability.WithDescription("API response status codes"),
	)

	valueParsingErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.value_parsing_errors_total",
		[]string{"source_type"},
		observability.WithDescription("Value parsing errors by source type"),
	)

	dbRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.db_requests_total",
		[]string{"method", "endpoint", "status"},
		observability.WithDescription("Database client HTTP requests"),
	)

	dbConnectionErrorsTotal = obsMetrics.Counter(
		"triggerx.condition_scheduler.db_connection_errors_total",
		observability.WithDescription("Database connection failures"),
	)

	dbRetriesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.db_retries_total",
		[]string{"endpoint"},
		observability.WithDescription("Database request retry attempts"),
	)

	actionExecutionDuration = obsMetrics.Histogram(
		"triggerx.condition_scheduler.action_execution_duration_seconds",
		observability.WithDescription("Time taken to execute actions"),
		observability.WithUnit("s"),
	)

	workerUptimeSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.worker_uptime_seconds",
		observability.WithDescription("Average worker uptime"),
		observability.WithUnit("s"),
	)

	workerErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.worker_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Worker errors by type"),
	)

	workerMemoryUsageBytes = obsMetrics.Gauge(
		"triggerx.condition_scheduler.worker_memory_usage_bytes",
		observability.WithDescription("Total memory usage across all workers"),
		observability.WithUnit("By"),
	)

	httpRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.http_requests_total",
		[]string{"method", "endpoint", "status_code"},
		observability.WithDescription("HTTP API requests received"),
	)

	httpClientConnectionErrorsTotal = obsMetrics.Counter(
		"triggerx.condition_scheduler.http_client_connection_errors_total",
		observability.WithDescription("HTTP client connection errors"),
	)

	duplicateConditionWindowSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.duplicate_condition_window_seconds",
		observability.WithDescription("Duplicate condition detection window"),
		observability.WithUnit("s"),
	)

	averageConditionCheckTimeSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.average_condition_check_time_seconds",
		observability.WithDescription("Mean condition check time"),
		observability.WithUnit("s"),
	)

	connectionFailuresTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.connection_failures_total",
		[]string{"chain_id"},
		observability.WithDescription("Blockchain connection failures"),
	)

	averageEventProcessingTimeSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.average_event_processing_time_seconds",
		observability.WithDescription("Mean event processing time"),
		observability.WithUnit("s"),
	)

	timeoutsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.timeouts_total",
		[]string{"operation"},
		observability.WithDescription("Operation timeouts"),
	)

	criticalErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.critical_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Critical system errors"),
	)

	invalidValuesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.invalid_values_total",
		[]string{"source"},
		observability.WithDescription("Invalid/unparseable values received"),
	)
}
