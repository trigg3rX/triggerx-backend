package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// Request tracking for RPS calculation
	requestCounts     = make(map[string]int64) // endpoint -> count
	requestCountsLock sync.Mutex
	lastRPSUpdate     = time.Now()

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
	// Update uptime
	go func() {
		ticker := time.NewTicker(config.GetMetricsUpdateInterval())
		defer ticker.Stop()

		for range ticker.C {
			uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
		}
	}()

	// Update system metrics
	go func() {
		ticker := time.NewTicker(config.GetMetricsUpdateInterval())
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

	// Calculate and update requests per second
	go func() {
		ticker := time.NewTicker(config.GetMetricsUpdateInterval())
		defer ticker.Stop()

		for range ticker.C {
			calculateAndUpdateRPS(ctx)
		}
	}()
}

// calculateAndUpdateRPS calculates requests per second from request counts
func calculateAndUpdateRPS(ctx context.Context) {
	requestCountsLock.Lock()
	defer requestCountsLock.Unlock()

	now := time.Now()
	timeWindow := now.Sub(lastRPSUpdate).Seconds()
	if timeWindow <= 0 {
		timeWindow = 1.0 // Avoid division by zero
	}

	// Calculate RPS for each endpoint and update metric
	for endpoint, count := range requestCounts {
		rps := float64(count) / timeWindow
		RecordRequestsPerSecond(ctx, endpoint, rps)
	}

	// Reset counts for next window
	for k := range requestCounts {
		requestCounts[k] = 0
	}
	lastRPSUpdate = now
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

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(ctx context.Context, method, endpoint, statusCode string, duration time.Duration) {
	if httpRequestsTotal != nil {
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc(ctx)
	}
	if httpRequestDuration != nil {
		httpRequestDuration.WithLabelValues(method, endpoint).Record(ctx, duration.Seconds())
	}

	// Track request count for RPS calculation
	requestCountsLock.Lock()
	requestCounts[endpoint]++
	requestCountsLock.Unlock()
}

// RecordRequestsPerSecond records requests per second for an endpoint
func RecordRequestsPerSecond(ctx context.Context, endpoint string, rps float64) {
	if requestsPerSecond != nil {
		requestsPerSecond.WithLabelValues(endpoint).Set(ctx, rps)
	}
}

// RecordKeeperCheckIn records a keeper check-in by version
func RecordKeeperCheckIn(ctx context.Context, version string) {
	if checkinsByVersionTotal != nil {
		checkinsByVersionTotal.WithLabelValues(version).Inc(ctx)
	}
}

// UpdateKeeperCounts updates the total and active keeper counts
func UpdateKeeperCounts(ctx context.Context, total, active int) {
	if keepersTotal != nil {
		keepersTotal.Set(ctx, float64(total))
	}
	if keepersActiveTotal != nil {
		keepersActiveTotal.Set(ctx, float64(active))
	}
}

// UpdateKeeperUptime updates the uptime for a specific keeper
func UpdateKeeperUptime(ctx context.Context, keeperAddress string, uptimeSeconds float64) {
	if keeperUptimeSeconds != nil {
		keeperUptimeSeconds.WithLabelValues(keeperAddress).Set(ctx, uptimeSeconds)
	}
}

// RecordMostActiveKeeperUptime records the uptime for the most active keeper
func RecordMostActiveKeeperUptime(ctx context.Context, keeperAddress string, uptimeSeconds float64) {
	if mostActiveKeeperSeconds != nil {
		mostActiveKeeperSeconds.WithLabelValues(keeperAddress).Add(ctx, uptimeSeconds)
	}
}

// RecordDBOperationDuration records the duration of a database operation
func RecordDBOperationDuration(ctx context.Context, operation string, duration time.Duration) {
	if dbHostOperationDuration != nil {
		dbHostOperationDuration.WithLabelValues(operation).Record(ctx, duration.Seconds())
	}
}

// RecordTelegramNotification records a telegram notification sent for a keeper
func RecordTelegramNotification(ctx context.Context, keeperAddress string) {
	if telegramKeeperNotificationsSentTotal != nil {
		telegramKeeperNotificationsSentTotal.WithLabelValues(keeperAddress).Inc(ctx)
	}
}

// RecordNetworkConnection records a network connection event
func RecordNetworkConnection(ctx context.Context, connType string) {
	if networkConnectionsTotal != nil {
		networkConnectionsTotal.WithLabelValues(connType).Inc(ctx)
	}
}
