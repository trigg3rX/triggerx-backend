package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the database server uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "uptime_seconds",
		Help:      "The uptime of the database server in seconds",
	})

	// StartupDuration tracks the time taken for database server to start
	StartupDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "startup_duration_seconds",
		Help:      "Time taken in seconds for the database server to start",
	})

	// Health check metrics with status
	HealthChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "health_checks_total",
		Help:      "Total health check requests",
	}, []string{"status"})

	// Total HTTP Request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "http_requests_total",
		Help:      "Total HTTP requests processed",
	}, []string{"method", "endpoint", "status"})

	// HTTP Request duration metrics
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "endpoint"})

	// Active HTTP requests
	ActiveRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "active_requests",
		Help:      "Currently active HTTP requests",
	}, []string{"endpoint"})

	// Request throughput rate
	RequestsPerSecond = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "requests_per_second",
		Help:      "Request throughput rate",
	}, []string{"endpoint"})

	// Average response time
	AverageResponseTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "average_response_time_seconds",
		Help:      "Mean response time",
	}, []string{"endpoint"})

	// Database operation metrics
	DatabaseOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "database_operations_total",
		Help:      "Total database operations performed",
	}, []string{"operation", "table", "status"})

	// Database queries metrics
	DBQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "db_queries_total",
		Help:      "Total database queries executed",
	}, []string{"query_type"})

	// Database query duration metrics
	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "db_query_duration_seconds",
		Help:      "Database query execution time",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"query_type"})

	// Database slow queries metrics
	DBSlowQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "db_slow_queries_total",
		Help:      "Queries exceeding time threshold",
	}, []string{"threshold"})

	// Connection metrics
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "active_connections",
		Help:      "Current number of active database connections",
	})
	// Retry mechanism metrics
	RetryAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "retry_attempts_total",
		Help:      "Retry mechanism attempts",
	}, []string{"endpoint", "attempt_number"})

	// Retry success metrics
	RetrySuccessesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "retry_successes_total",
		Help:      "Successful retries",
	}, []string{"endpoint"})

	// Retry failure metrics
	RetryFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "retry_failures_total",
		Help:      "Failed retry attempts",
	}, []string{"endpoint"})

	// Error and reliability metrics
	DatabaseErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "database_errors_total",
		Help:      "Database errors (error_type=timeout/connection/query/constraint)",
	}, []string{"error_type"})

	// Panic recovery metrics
	PanicRecoveriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "panic_recoveries_total",
		Help:      "Panic recovery instances",
	}, []string{"endpoint"})

	// Request timeout metrics
	RequestTimeoutsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "request_timeouts_total",
		Help:      "Request timeout occurrences",
	}, []string{"endpoint"})

	// Memory usage metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "memory_usage_bytes",
		Help:      "Memory consumption",
	})

	// CPU usage metrics
	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	// Goroutines active metrics
	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "goroutines_active",
		Help:      "Active Go routines",
	})

	// Garbage collection duration metrics
	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// Database operations by entity
	DBOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "db_operations_total",
		Help:      "Total database operations (entity=job/task/user/keeper/apikey, operation=create/read/update/delete)",
	}, []string{"entity", "operation", "status"})

	// Database operation duration metrics
	DatabaseOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "database_operation_duration_seconds",
		Help:      "Database operation duration in seconds",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"operation", "table"})

	// Database operations per second metrics
	DBOperationsPerSecond = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "db_server",
		Name:      "db_operations_per_second",
		Help:      "Database operations throughput rate",
	}, []string{"operation"})
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
		}
	}()
}
