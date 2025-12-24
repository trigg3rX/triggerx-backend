package metrics

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// UptimeSeconds tracks the database server uptime in seconds
	uptimeSeconds observability.Gauge

	// Health check metrics with status
	healthChecksTotal *observability.CounterVec

	// Total HTTP Request metrics
	httpRequestsTotal *observability.CounterVec

	// HTTP Request duration metrics
	httpRequestDuration *observability.HistogramVec

	// Active HTTP requests
	activeRequests *observability.GaugeVec

	// Request throughput rate
	requestsPerSecond *observability.GaugeVec

	// Average response time
	averageResponseTime *observability.GaugeVec

	// Database operation metrics
	databaseOperationsTotal *observability.CounterVec

	// Database query duration metrics
	dbQueryDuration *observability.HistogramVec

	// Database slow queries metrics
	dbSlowQueriesTotal *observability.CounterVec

	// Connection metrics
	activeConnections observability.Gauge

	// Retry mechanism metrics
	retryAttemptsTotal *observability.CounterVec

	// Retry success metrics
	retrySuccessesTotal *observability.CounterVec

	// Retry failure metrics
	retryFailuresTotal *observability.CounterVec

	// Error and reliability metrics
	databaseErrorsTotal *observability.CounterVec

	// Panic recovery metrics
	panicRecoveriesTotal *observability.CounterVec

	// Request timeout metrics
	requestTimeoutsTotal *observability.CounterVec

	// Memory usage metrics
	memoryUsageBytes observability.Gauge

	// CPU usage metrics
	cpuUsagePercent observability.Gauge

	// Goroutines active metrics
	goroutinesActive observability.Gauge

	// Garbage collection duration metrics
	gcDurationSeconds observability.Gauge

	// Database operations by entity
	dbOperationsTotal *observability.CounterVec

	// Database operation duration metrics
	databaseOperationDuration *observability.HistogramVec

	// Database operations per second metrics
	dbOperationsPerSecond *observability.GaugeVec
)

// Exported variables for use in other packages
var (
	UptimeSeconds             observability.Gauge
	HealthChecksTotal         *observability.CounterVec
	HTTPRequestsTotal         *observability.CounterVec
	HTTPRequestDuration       *observability.HistogramVec
	ActiveRequests            *observability.GaugeVec
	RequestsPerSecond         *observability.GaugeVec
	AverageResponseTime       *observability.GaugeVec
	DatabaseOperationsTotal   *observability.CounterVec
	DBQueryDuration           *observability.HistogramVec
	DBSlowQueriesTotal        *observability.CounterVec
	ActiveConnections         observability.Gauge
	RetryAttemptsTotal        *observability.CounterVec
	RetrySuccessesTotal       *observability.CounterVec
	RetryFailuresTotal        *observability.CounterVec
	DatabaseErrorsTotal       *observability.CounterVec
	PanicRecoveriesTotal      *observability.CounterVec
	RequestTimeoutsTotal      *observability.CounterVec
	MemoryUsageBytes          observability.Gauge
	CPUUsagePercent           observability.Gauge
	GoroutinesActive          observability.Gauge
	GCDurationSeconds         observability.Gauge
	DBOperationsTotal         *observability.CounterVec
	DatabaseOperationDuration *observability.HistogramVec
	DBOperationsPerSecond     *observability.GaugeVec
)

// InitializeMetrics initializes all metrics using the observability metrics instance
func InitializeMetrics(obsMetrics observability.Metrics) {
	// UptimeSeconds tracks the database server uptime in seconds
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.db_server.uptime_seconds",
		observability.WithDescription("The uptime of the database server in seconds"),
		observability.WithUnit("s"),
	)
	UptimeSeconds = uptimeSeconds

	// Health check metrics with status
	healthChecksTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.health_checks_total",
		[]string{"status"},
		observability.WithDescription("Total health check requests"),
	)
	HealthChecksTotal = healthChecksTotal

	// Total HTTP Request metrics
	httpRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.http_requests_total",
		[]string{"method", "endpoint", "status"},
		observability.WithDescription("Total HTTP requests processed"),
	)
	HTTPRequestsTotal = httpRequestsTotal

	// HTTP Request duration metrics
	httpRequestDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.db_server.http_request_duration_seconds",
		[]string{"method", "endpoint"},
		observability.WithDescription("HTTP request duration in seconds"),
		observability.WithUnit("s"),
	)
	HTTPRequestDuration = httpRequestDuration

	// Active HTTP requests
	activeRequests = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.db_server.active_requests",
		[]string{"endpoint"},
		observability.WithDescription("Currently active HTTP requests"),
	)
	ActiveRequests = activeRequests

	// Request throughput rate
	requestsPerSecond = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.db_server.requests_per_second",
		[]string{"endpoint"},
		observability.WithDescription("Request throughput rate"),
	)
	RequestsPerSecond = requestsPerSecond

	// Average response time
	averageResponseTime = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.db_server.average_response_time_seconds",
		[]string{"endpoint"},
		observability.WithDescription("Mean response time"),
		observability.WithUnit("s"),
	)
	AverageResponseTime = averageResponseTime

	// Database operation metrics
	databaseOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.database_operations_total",
		[]string{"operation", "table", "status"},
		observability.WithDescription("Total database operations performed"),
	)
	DatabaseOperationsTotal = databaseOperationsTotal

	// Database query duration metrics
	dbQueryDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.db_server.db_query_duration_seconds",
		[]string{"query_type"},
		observability.WithDescription("Database query execution time"),
		observability.WithUnit("s"),
	)
	DBQueryDuration = dbQueryDuration

	// Database slow queries metrics
	dbSlowQueriesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.db_slow_queries_total",
		[]string{"threshold"},
		observability.WithDescription("Queries exceeding time threshold"),
	)
	DBSlowQueriesTotal = dbSlowQueriesTotal

	// Connection metrics
	activeConnections = obsMetrics.Gauge(
		"triggerx.db_server.active_connections",
		observability.WithDescription("Current number of active database connections"),
	)
	ActiveConnections = activeConnections

	// Retry mechanism metrics
	retryAttemptsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.retry_attempts_total",
		[]string{"endpoint", "attempt_number"},
		observability.WithDescription("Retry mechanism attempts"),
	)
	RetryAttemptsTotal = retryAttemptsTotal

	// Retry success metrics
	retrySuccessesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.retry_successes_total",
		[]string{"endpoint"},
		observability.WithDescription("Successful retries"),
	)
	RetrySuccessesTotal = retrySuccessesTotal

	// Retry failure metrics
	retryFailuresTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.retry_failures_total",
		[]string{"endpoint"},
		observability.WithDescription("Failed retry attempts"),
	)
	RetryFailuresTotal = retryFailuresTotal

	// Error and reliability metrics
	databaseErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.database_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Database errors"),
	)
	DatabaseErrorsTotal = databaseErrorsTotal

	// Panic recovery metrics
	panicRecoveriesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.panic_recoveries_total",
		[]string{"endpoint"},
		observability.WithDescription("Panic recovery instances"),
	)
	PanicRecoveriesTotal = panicRecoveriesTotal

	// Request timeout metrics
	requestTimeoutsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.request_timeouts_total",
		[]string{"endpoint"},
		observability.WithDescription("Request timeout occurrences"),
	)
	RequestTimeoutsTotal = requestTimeoutsTotal

	// Memory usage metrics
	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.db_server.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)
	MemoryUsageBytes = memoryUsageBytes

	// CPU usage metrics
	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.db_server.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)
	CPUUsagePercent = cpuUsagePercent

	// Goroutines active metrics
	goroutinesActive = obsMetrics.Gauge(
		"triggerx.db_server.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)
	GoroutinesActive = goroutinesActive

	// Garbage collection duration metrics
	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.db_server.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)
	GCDurationSeconds = gcDurationSeconds

	// Database operations by entity
	dbOperationsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.db_server.db_operations_total",
		[]string{"entity", "operation", "status"},
		observability.WithDescription("Total database operations by entity"),
	)
	DBOperationsTotal = dbOperationsTotal

	// Database operation duration metrics
	databaseOperationDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.db_server.database_operation_duration_seconds",
		[]string{"operation", "table"},
		observability.WithDescription("Database operation duration in seconds"),
		observability.WithUnit("s"),
	)
	DatabaseOperationDuration = databaseOperationDuration

	// Database operations per second metrics
	dbOperationsPerSecond = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.db_server.db_operations_per_second",
		[]string{"operation"},
		observability.WithDescription("Database operations throughput rate"),
	)
	DBOperationsPerSecond = dbOperationsPerSecond
}
