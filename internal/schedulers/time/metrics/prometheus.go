package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the service uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "uptime_seconds",
		Help:      "Time passed since Time Scheduler started in seconds",
	})

	// Memory usage metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "memory_usage_bytes",
		Help:      "Memory consumption",
	})

	// CPU usage metrics
	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	// Goroutines active metrics
	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "goroutines_active",
		Help:      "Number of active goroutines",
	})

	// Garbage collection duration metrics
	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// API server status
	APIServerStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "api_server_status",
		Help:      "API server health status",
	})

	// Tasks per minute
	TasksPerMinute = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_per_minute",
		Help:      "Task throughput rate",
	})

	// Average task completion time
	AverageTaskCompletionTimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "average_task_completion_time_seconds",
		Help:      "Mean task completion time",
	})

	// Task success rate
	TaskSuccessRatePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "task_success_rate_percent",
		Help:      "Overall task success percentage",
	})

	// Tasks scheduled
	TasksScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "scheduler_tasks_scheduled",
		Help:      "Total number of jobs currently scheduled",
	})

	// Tasks completed
	TasksCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "scheduler_tasks_completed",
		Help:      "Total number of jobs completed (success/fail)",
	}, []string{"status"})

	// Task execution time
	TaskExecutionTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "scheduler_task_execution_time",
		Help:      "Time taken to execute a task in seconds",
		Buckets:   []float64{1, 5, 10, 50, 100, 500},
	})

	// Tasks expired
	TasksExpiredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_expired_total",
		Help:      "Tasks that expired before execution",
	})

	// Job batch size
	JobBatchSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "job_batch_size",
		Help:      "Number of jobs processed per batch",
	})

	// DB requests
	DBRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "db_requests_total",
		Help:      "Database client HTTP requests",
	}, []string{"method", "endpoint", "status"})

	// HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "http_requests_total",
		Help:      "HTTP API requests received",
	}, []string{"method", "endpoint", "status_code"})

	// DB connection errors
	DBConnectionErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "db_connection_errors_total",
		Help:      "Database connection failures",
	})

	// DB retries
	DBRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "db_retries_total",
		Help:      "Database request retry attempts",
	}, []string{"endpoint"})

	// Task broadcasts
	TaskBroadcastsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "task_broadcasts_total",
		Help:      "Task broadcasts to performers",
	}, []string{"status"})

	// Tasks by schedule type
	TasksByScheduleTypeTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_by_schedule_type_total",
		Help:      "Tasks processed by schedule type",
	}, []string{"type"})

	// Duplicate task window
	DuplicateTaskWindowSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "duplicate_task_window_seconds",
		Help:      "Duplicate task detection window",
	})

	// Internal tracking variables for performance calculations
	taskStatsLock       sync.RWMutex
	taskCompletionTimes []float64
	successfulTasks     int64
	totalTasks          int64
	tasksLastMinute     int64
	lastMinuteReset     time.Time
)

// Starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
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
