package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// Metrics instances
	uptimeSeconds                        observability.Gauge
	keepersTotal                         observability.Gauge
	keepersActiveTotal                   observability.Gauge
	httpRequestsTotal                    *observability.CounterVec
	httpRequestDuration                  *observability.HistogramVec
	requestsPerSecond                    *observability.GaugeVec
	checkinsByVersionTotal               *observability.CounterVec
	keeperUptimeSeconds                  *observability.GaugeVec
	mostActiveKeeperSeconds              *observability.CounterVec
	dbHostOperationDuration              *observability.HistogramVec
	telegramKeeperNotificationsSentTotal *observability.CounterVec
	memoryUsageBytes                     observability.Gauge
	cpuUsagePercent                      observability.Gauge
	goroutinesActive                     observability.Gauge
	gcDurationSeconds                    observability.Gauge
	networkConnectionsTotal              *observability.CounterVec
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
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
			memoryUsageBytes.Set(ctx, float64(m.Alloc))
			goroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
			gcDurationSeconds.Set(ctx, float64(m.PauseTotalNs)/1e9) // Convert nanoseconds to seconds

			// CPU usage metrics
			cpuPercentages, err := cpu.Percent(0, false)
			if err == nil && len(cpuPercentages) > 0 {
				cpuUsagePercent.Set(ctx, cpuPercentages[0])
			} else {
				// Fallback to 0.0 if CPU monitoring fails
				cpuUsagePercent.Set(ctx, 0.0)
			}
		}
	}()
}

// InitializeMetrics initializes all metrics using the observability metrics instance
// This must be called before using any metrics
func InitializeMetrics(obsMetrics observability.Metrics) {
	// Simple gauges
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.health_service.uptime_seconds",
		observability.WithDescription("Time passed since Health Service started in seconds"),
		observability.WithUnit("s"),
	)

	keepersTotal = obsMetrics.Gauge(
		"triggerx.health_service.keepers_total",
		observability.WithDescription("Total number of registered keepers"),
	)

	keepersActiveTotal = obsMetrics.Gauge(
		"triggerx.health_service.keepers_active_total",
		observability.WithDescription("Currently active keepers"),
	)

	// Labeled metrics (Vec types)
	httpRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.health_service.http_requests_total",
		[]string{"method", "endpoint", "status_code"},
		observability.WithDescription("Total HTTP requests received"),
	)

	httpRequestDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.health_service.http_request_duration_seconds",
		[]string{"method", "endpoint"},
		observability.WithDescription("HTTP request processing time"),
		observability.WithUnit("s"),
	)

	requestsPerSecond = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.health_service.requests_per_second",
		[]string{"endpoint"},
		observability.WithDescription("Request throughput rate"),
		observability.WithUnit("1/s"),
	)

	checkinsByVersionTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.health_service.checkins_by_version_total",
		[]string{"version"},
		observability.WithDescription("Check-ins by keeper version"),
	)

	keeperUptimeSeconds = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.health_service.keeper_uptime_seconds",
		[]string{"keeper_address"},
		observability.WithDescription("Keeper uptime since first check-in"),
		observability.WithUnit("s"),
	)

	mostActiveKeeperSeconds = observability.NewCounterVec(
		obsMetrics,
		"triggerx.health_service.most_active_keeper_uptime_seconds",
		[]string{"keeper_address"},
		observability.WithDescription("Most active keeper uptime since first check-in"),
		observability.WithUnit("s"),
	)

	dbHostOperationDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.health_service.db_host_operation_duration_seconds",
		[]string{"operation"},
		observability.WithDescription("Scylla Database operation execution time"),
		observability.WithUnit("s"),
	)

	telegramKeeperNotificationsSentTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.health_service.telegram_keeper_notifications_sent_total",
		[]string{"keeper_address"},
		observability.WithDescription("Notifications sent per keeper"),
	)

	// Performance metrics
	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.health_service.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.health_service.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.health_service.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.health_service.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	networkConnectionsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.health_service.network_connections_total",
		[]string{"type"},
		observability.WithDescription("Network connections (type=incoming/outgoing)"),
	)
}
