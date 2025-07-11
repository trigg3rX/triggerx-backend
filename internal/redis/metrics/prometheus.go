package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	redisClient "github.com/trigg3rX/triggerx-backend-imua/pkg/client/redis"
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

	// Single flag to indicate which Redis is being used
	IsRedisUpstashAvailable = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "is_upstash_available",
		Help:      "Whether Upstash Redis is available and being used (1=Upstash, 0=Local)",
	})

	// Connection Management
	ClientConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "client_connections_total",
		Help:      "Redis client connections (status=success/failure)",
	}, []string{"status"})

	ClientConnectionErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "client_connection_errors_total",
		Help:      "Redis client connection errors",
	}, []string{"error_type"})

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

	// Updated buckets for longer task execution times (30+ seconds to several minutes)
	TaskLifecycleTransitionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "task_lifecycle_transition_duration_seconds",
		Help:      "Task lifecycle transition time",
		Buckets:   []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120, 300, 600}, // Up to 10 minutes
	}, []string{"from_stream", "to_stream"})

	// Redis Client Operation Metrics
	RedisOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "operations_total",
		Help:      "Total Redis operations performed",
	}, []string{"operation", "status"})

	RedisOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "operation_duration_seconds",
		Help:      "Redis operation duration",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"operation"})

	RedisRetryAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "retry_attempts_total",
		Help:      "Total Redis retry attempts",
	}, []string{"operation"})

	RedisConnectionRecoveries = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "connection_recoveries_total",
		Help:      "Total Redis connection recovery attempts",
	}, []string{"status"})

	RedisConnectionHealth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "redis",
		Name:      "connection_health",
		Help:      "Redis connection health status",
	}, []string{"type"})
)

// CreateRedisMonitoringHooks creates monitoring hooks for the Redis client
func CreateRedisMonitoringHooks() *redisClient.MonitoringHooks {
	return &redisClient.MonitoringHooks{
		OnOperationStart: func(operation string, key string) {
			// Optional: Could track operation starts if needed
		},

		OnOperationEnd: func(operation string, key string, duration time.Duration, err error) {
			status := "success"
			if err != nil {
				status = "error"
			}

			RedisOperationsTotal.WithLabelValues(operation, status).Inc()
			RedisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
		},

		OnConnectionStatus: func(connected bool, latency time.Duration) {
			healthValue := float64(0)
			if connected {
				healthValue = float64(1)
			}
			RedisConnectionHealth.WithLabelValues("main").Set(healthValue)

			// Update ping duration if connected
			if connected {
				PingDuration.Observe(latency.Seconds())
				PingOperationsTotal.WithLabelValues("success").Inc()
			} else {
				PingOperationsTotal.WithLabelValues("failure").Inc()
			}
		},

		OnRecoveryStart: func(reason string) {
			// Mark connection as unhealthy during recovery
			RedisConnectionHealth.WithLabelValues("main").Set(float64(0))
		},

		OnRecoveryEnd: func(success bool, attempts int, duration time.Duration) {
			status := "failure"
			if success {
				status = "success"
				RedisConnectionHealth.WithLabelValues("main").Set(float64(1))
			}
			RedisConnectionRecoveries.WithLabelValues(status).Inc()
		},

		OnRetryAttempt: func(operation string, attempt int, err error) {
			RedisRetryAttempts.WithLabelValues(operation).Inc()
		},
	}
}

// UpdateRedisClientMetrics updates metrics from Redis client operation metrics
func UpdateRedisClientMetrics(operationMetrics map[string]*redisClient.OperationMetrics) {
	for operation, metrics := range operationMetrics {
		// Update operation counters
		RedisOperationsTotal.WithLabelValues(operation, "success").Add(float64(metrics.SuccessCount))
		RedisOperationsTotal.WithLabelValues(operation, "error").Add(float64(metrics.ErrorCount))

		// Update retry attempts
		RedisRetryAttempts.WithLabelValues(operation).Add(float64(metrics.RetryCount))

		// Update average latency (as a gauge for monitoring)
		if metrics.AverageLatency > 0 {
			RedisOperationDuration.WithLabelValues(operation).Observe(metrics.AverageLatency.Seconds())
		}
	}
}
