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
		Subsystem: "keeper",
		Name:      "uptime_seconds",
		Help:      "The uptime of the keeper service in seconds",
	})

	// Total task processing requests on keeper API server
	TasksReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "tasks_received_total",
		Help:      "Total tasks received",
	})

	// Totla task completed successfully, type: executed, validated
	TasksCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "tasks_completed_total",
		Help:      "Total tasks completed",
	}, []string{"type"})

	// Time taken for task completion, type: executed, validated
	TaskDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "task_duration_seconds",
		Help:      "Time taken for task completion",
		Buckets:   prometheus.DefBuckets,
	}, []string{"type"})

	// Total tasks validated by type/id: 1-6
	TasksByDefinitionIDTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "tasks_by_definition_id_total",
		Help:      "Tasks validated by type/id",
	}, []string{"id"})

	// Transaction metrics
	TransactionsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "transactions_sent_total",
		Help:      "Total transactions done for task executions",
	}, []string{"chain_id", "status"})
	GasUsedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "gas_used_total",
		Help:      "Total gas used in transactions",
	}, []string{"chain_id"})
	TransactionFeesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "transaction_fees_total",
		Help:      "Total transaction fee incurred in transactions",
	}, []string{"chain_id"})

	// IPFS metrics
	IPFSDownloadSizeBytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "ipfs_download_size_bytes",
		Help:      "Total IPFS content downloaded",
	})
	IPFSUploadSizeBytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "ipfs_upload_size_bytes",
		Help:      "Total IPFS content uploaded",
	})

	// Health metrics
	SuccessfulHealthCheckinsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "successful_health_checkins_total",
		Help:      "Total successful health checkins",
	})

	// System metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "memory_usage_bytes",
		Help:      "Memory consumption",
	})

	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization",
	})

	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "goroutines_active",
		Help:      "Active Go routines",
	})

	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// Docker metrics
	DockerContainersCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "docker_containers_created_total",
		Help:      "Docker container creation count",
	}, []string{"language"})
	
	DockerContainerDurationSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "docker_container_duration_seconds",
		Help:      "Container execution time",
	}, []string{"language"})

	// Aggregate metrics
	TaskSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "task_success_rate",
		Help:      "Overall task success percentage",
	}, []string{"type"})

	// Average task completion time in seconds, type: executed, validated
	// AverageTaskCompletionTimeSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: "triggerx",
	// 	Subsystem: "keeper",
	// 	Name:      "average_task_completion_time_seconds",
	// 	Help:      "Mean completion time",
	// }, []string{"type"})

	// Tasks per minute, type: executed, validated
	TasksPerDay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "tasks_per_day",
		Help:      "Task throughput rate",
	}, []string{"type"})

	// Total service restarts
	RestartsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "keeper",
		Name:      "restarts_total",
		Help:      "Service restart count",
	})
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
		}
	}()

	// Reset metrics every day
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			TasksPerDay.Reset()
		}
	}()
}
