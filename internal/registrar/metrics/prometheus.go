package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the registrar service uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "uptime_seconds",
		Help:      "Time passed since Registrar started in seconds",
	})

	// RPC request metrics
	RPCRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "rpc_requests_total",
		Help:      "RPC requests to blockchain nodes",
	}, []string{"chain", "method", "status"})

	// RPC error metrics
	RPCErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "rpc_errors_total",
		Help:      "RPC request errors",
	}, []string{"chain", "error_type"})

	// Current block number metrics
	CurrentBlockNumber = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "current_block_number",
		Help:      "Current processed block number",
	}, []string{"chain"})

	// Event detection metrics
	EventsDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "events_detected_total",
		Help:      "Events detected (event_type=task_submitted/task_rejected)",
	}, []string{"event_type"})

	// Points distribution metrics
	PointsDistributedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "points_distributed_total",
		Help:      "Points distributed (recipient_type=performer/attester/user)",
	}, []string{"recipient_type"})

	// Daily rewards processing metrics
	NoOfDaysRewardsProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "no_of_days_rewards_processed_total",
		Help:      "Daily rewards processing executions",
	})

	// Pinata file cleanup metrics
	PinataFilesCleanedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "pinata_files_cleaned_total",
		Help:      "Files cleaned from Pinata",
	})

	// Memory usage metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "memory_usage_bytes",
		Help:      "Memory consumption",
	})

	// CPU usage metrics
	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	// Goroutines active metrics
	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "goroutines_active",
		Help:      "Active Go routines",
	})

	// Garbage collection duration metrics
	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "registrar",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})
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
