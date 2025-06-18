package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/v3/cpu"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the health service uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "uptime_seconds",
		Help:      "Time passed since Health Service started in seconds",
	})

	// HTTP request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "http_requests_total",
		Help:      "Total HTTP requests received",
	}, []string{"method", "endpoint", "status_code"})

	// HTTP Request duration metrics
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request processing time",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "endpoint"})

	// Request throughput rate
	RequestsPerSecond = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "requests_per_second",
		Help:      "Request throughput rate",
	}, []string{"endpoint"})

	// Keeper check-in metrics
	CheckinsByVersionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "checkins_by_version_total",
		Help:      "Check-ins by keeper version",
	}, []string{"version"})

	// Keeper status metrics
	KeepersTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "keepers_total",
		Help:      "Total number of registered keepers",
	})

	// Keepers active total metrics
	KeepersActiveTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "keepers_active_total",
		Help:      "Currently active keepers",
	})

	// Keepers inactive total metrics
	KeepersInactiveTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "keepers_inactive_total",
		Help:      "Currently inactive keepers",
	})

	// Keeper uptime tracking
	KeeperUptimeSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "keeper_uptime_seconds",
		Help:      "Keeper uptime since first check-in",
	}, []string{"keeper_address"})

	// Keeper uptime tracking
	MostActiveKeeperSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "most_active_keeper_uptime_seconds",
		Help:      "Most active keeper uptime since first check-in",
	}, []string{"keeper_address"})

	// Database operation metrics
	DBHostOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "db_host_operation_duration_seconds",
		Help:      "Scylla Database operation execution time",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"operation"})

	// Telegram notification metrics
	TelegramKeeperNotificationsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "telegram_keeper_notifications_sent_total",
		Help:      "Notifications sent per keeper",
	}, []string{"keeper_address"})

	// Memory usage metrics
	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "memory_usage_bytes",
		Help:      "Memory consumption",
	})

	// CPU usage metrics
	CPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "cpu_usage_percent",
		Help:      "CPU utilization percentage",
	})

	// Goroutines active metrics
	GoroutinesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "goroutines_active",
		Help:      "Active Go routines",
	})

	// Garbage collection duration metrics
	GCDurationSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "gc_duration_seconds",
		Help:      "Garbage collection time",
	})

	// Network connection metrics
	NetworkConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "health_service",
		Name:      "network_connections_total",
		Help:      "Network connections (type=incoming/outgoing)",
	}, []string{"type"})
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

	// Update system metrics every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Memory metrics
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			MemoryUsageBytes.Set(float64(m.Alloc))
			GoroutinesActive.Set(float64(runtime.NumGoroutine()))
			GCDurationSeconds.Set(float64(m.PauseTotalNs) / 1e9) // Convert nanoseconds to seconds

			// CPU usage metrics
			cpuPercentages, err := cpu.Percent(0, false)
			if err == nil && len(cpuPercentages) > 0 {
				CPUUsagePercent.Set(cpuPercentages[0])
			} else {
				// Fallback to 0.0 if CPU monitoring fails
				CPUUsagePercent.Set(0.0)
			}
		}
	}()
}
