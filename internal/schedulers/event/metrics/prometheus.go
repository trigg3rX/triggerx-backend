package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the service uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "uptime_seconds",
		Help:      "Time passed since Event Scheduler started in seconds",
	})

	// Memory usage metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "memory_usage_bytes",
		Help:      "Total memory consumption",
	})

	// CPU usage metrics
	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	// Goroutines active metrics
	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "goroutines_active",
		Help:      "Number of active goroutines",
	})

	// Garbage collection duration metrics
	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// Events per minute by chain
	EventsPerMinute = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "events_per_minute",
		Help:      "Event detection rate per chain",
	}, []string{"chain_id"})

	// Jobs scheduled
	JobsScheduled = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "scheduler_jobs_scheduled",
		Help:      "Total number of jobs scheduled",
	})

	// Jobs completed
	JobsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "scheduler_jobs_completed",
		Help:      "Total number of jobs completed successfully or failed",
	}, []string{"status"})

	// Active workers
	ActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "active_workers",
		Help:      "Number of active job workers currently running",
	})

	// Chain connections
	ChainConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "chain_connections_total",
		Help:      "Blockchain connection attempts",
	}, []string{"chain_id", "status"})

	// RPC requests
	RPCRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "rpc_requests_total",
		Help:      "RPC requests to blockchain nodes",
	}, []string{"chain_id", "method", "status"})

	// DB requests
	DBRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "db_requests_total",
		Help:      "Database client HTTP requests",
	}, []string{"method", "endpoint", "status"})

	// DB connection errors
	DBConnectionErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "db_connection_errors_total",
		Help:      "Database connection failures",
	})

	// DB retries
	DBRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "db_retries_total",
		Help:      "Database request retry attempts",
	}, []string{"endpoint"})

	// Worker uptime
	WorkerUptimeSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "worker_uptime_seconds",
		Help:      "Individual worker uptime",
	}, []string{"job_id"})

	// Worker errors
	WorkerErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "worker_errors_total",
		Help:      "Worker errors by type",
	}, []string{"job_id", "error_type"})

	// Worker memory usage
	WorkerMemoryUsageBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "worker_memory_usage_bytes",
		Help:      "Memory usage per worker",
	}, []string{"job_id"})

	// Action executions
	ActionExecutionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "action_executions_total",
		Help:      "Action executions triggered by events",
	}, []string{"job_id", "status"})

	// HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "http_requests_total",
		Help:      "HTTP API requests received",
	}, []string{"method", "endpoint", "status_code"})

	// Duplicate event window
	DuplicateEventWindowSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "duplicate_event_window_seconds",
		Help:      "Duplicate event detection window",
	})

	// Average event processing time
	AverageEventProcessingTimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "average_event_processing_time_seconds",
		Help:      "Mean event processing time",
	})

	// Connection failures
	ConnectionFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "connection_failures_total",
		Help:      "Blockchain connection failures",
	}, []string{"chain_id"})

	// Critical errors
	CriticalErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "event_scheduler",
		Name:      "critical_errors_total",
		Help:      "Critical system errors",
	}, []string{"error_type"})
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime and collect metrics every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
			collectSystemMetrics()
			collectConfigurationMetrics()
			collectPerformanceMetrics()
			collectWorkerMetrics()
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
