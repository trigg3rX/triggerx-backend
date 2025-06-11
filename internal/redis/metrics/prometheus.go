package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// System metrics
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "uptime_seconds",
		Help:      "Time passed since Redis Service started in seconds",
	})

	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "memory_usage_bytes",
		Help:      "Service memory consumption",
	})

	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "goroutines_active",
		Help:      "Active Go routines",
	})

	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// Service Health & Availability
	ServiceStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "service_status",
		Help:      "Service component health status (component=client/job_stream_manager/task_stream_manager)",
	}, []string{"component"})

	RedisAvailable = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "redis_available",
		Help:      "Redis availability flag",
	})

	RedisTypeActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "redis_type_active",
		Help:      "Active Redis type (type=upstash/local)",
	}, []string{"type"})

	UpstashEnabled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "upstash_enabled",
		Help:      "Upstash Redis enabled flag",
	})

	LocalRedisEnabled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "local_redis_enabled",
		Help:      "Local Redis enabled flag",
	})

	// Connection Management
	ClientConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "client_connections_total",
		Help:      "Redis client connections (redis_type=upstash/local, status=success/failure)",
	}, []string{"redis_type", "status"})

	ClientConnectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "client_connection_duration_seconds",
		Help:      "Redis client connection establishment time",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"redis_type"})

	ClientConnectionErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "client_connection_errors_total",
		Help:      "Redis client connection errors",
	}, []string{"redis_type", "error_type"})

	PingOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "ping_operations_total",
		Help:      "Redis ping operations",
	}, []string{"status"})

	PingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "ping_duration_seconds",
		Help:      "Redis ping response time",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
	})

	ConnectionChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "connection_checks_total",
		Help:      "Connection health checks",
	}, []string{"status"})

	// Core Stream Operations
	TaskStreamLengths = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_stream_lengths",
		Help:      "Current task stream lengths (stream=ready/retry/processing/completed/failed)",
	}, []string{"stream"})

	JobStreamLengths = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "job_stream_lengths",
		Help:      "Current job stream lengths (stream=running/completed)",
	}, []string{"stream"})

	TasksAddedToStreamTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "tasks_added_to_stream_total",
		Help:      "Tasks added to streams",
	}, []string{"stream", "status"})

	TasksReadFromStreamTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "tasks_read_from_stream_total",
		Help:      "Tasks read from streams",
	}, []string{"stream", "status"})

	JobsAddedToStreamTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "jobs_added_to_stream_total",
		Help:      "Jobs added to streams",
	}, []string{"stream", "status"})

	JobsReadFromStreamTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "jobs_read_from_stream_total",
		Help:      "Jobs read from streams",
	}, []string{"stream", "status"})

	// Task Lifecycle & Performance
	TaskRetryOperationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_retry_operations_total",
		Help:      "Task retry operations",
	})

	TaskMaxRetriesExceededTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_max_retries_exceeded_total",
		Help:      "Tasks exceeding max retry attempts",
	})

	TasksMovedToFailedStreamTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "tasks_moved_to_failed_stream_total",
		Help:      "Tasks permanently failed and moved to failed stream",
	})

	TaskReadyToProcessingTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_ready_to_processing_total",
		Help:      "Tasks moved from ready to processing stream",
	})

	TaskProcessingToCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_processing_to_completed_total",
		Help:      "Tasks moved from processing to completed stream",
	})

	TaskLifecycleTransitionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_lifecycle_transition_duration_seconds",
		Help:      "Task lifecycle transition time",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"from_stream", "to_stream"})
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
