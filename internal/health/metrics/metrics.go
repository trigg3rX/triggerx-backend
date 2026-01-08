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

	// Version tracking for keepers online by version metric
	lastKnownVersions map[string]bool // Track versions we've seen to reset them to 0 when needed
	versionsLock      sync.Mutex      // Protect lastKnownVersions map

	// System Metrics
	uptimeSeconds     observability.Gauge
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// Keeper Metrics
	keepersTotal           observability.Gauge
	keepersActiveTotal     observability.Gauge
	keeperUptimeSeconds    *observability.GaugeVec
	keepersOnlineByVersion *observability.GaugeVec

	// HTTP Metrics
	httpRequestsTotal *observability.CounterVec
	requestsPerSecond *observability.GaugeVec
	// telegramKeeperNotificationsSentTotal *observability.CounterVec // Not used yet

	// Database Metrics
	dbOperationDuration *observability.HistogramVec
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
	// Initialize the lastKnownVersions map
	lastKnownVersions = make(map[string]bool)

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

	requestsPerSecond = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.health_service.requests_per_second",
		[]string{"endpoint"},
		observability.WithDescription("Request throughput rate"),
		observability.WithUnit("1/s"),
	)

	keeperUptimeSeconds = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.health_service.keeper_uptime_seconds",
		[]string{"keeper_address"},
		observability.WithDescription("Keeper uptime since first check-in"),
		observability.WithUnit("s"),
	)

	keepersOnlineByVersion = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.health_service.keepers_online_by_version",
		[]string{"version"},
		observability.WithDescription("Number of keepers online by version"),
	)

	// telegramKeeperNotificationsSentTotal = observability.NewCounterVec(
	// 	obsMetrics,
	// 	"triggerx.health_service.telegram_keeper_notifications_sent_total",
	// 	[]string{"keeper_address"},
	// 	observability.WithDescription("Notifications sent per keeper"),
	// )

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

	// Database operation duration metrics
	dbOperationDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.health_service.db_operation_duration_seconds",
		[]string{"operation", "table"},
		observability.WithDescription("Database operation duration in seconds"),
		observability.WithUnit("s"),
	)
}

// RecordHTTPRequest records HTTP request metrics
// Note: Request duration is tracked via tracers/spans instead of metrics
func RecordHTTPRequest(ctx context.Context, method, endpoint, statusCode string, duration time.Duration) {
	if httpRequestsTotal != nil {
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc(ctx)
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
// uptimeSeconds should be the cumulative uptime from the database (in seconds)
func UpdateKeeperUptime(ctx context.Context, keeperAddress string, uptimeSeconds float64) {
	if keeperUptimeSeconds != nil {
		keeperUptimeSeconds.WithLabelValues(keeperAddress).Set(ctx, uptimeSeconds)
	}
}

// RecordDBOperationDuration records database operation duration
func RecordDBOperationDuration(ctx context.Context, operation string, duration time.Duration) {
	if dbOperationDuration != nil {
		// All operations in health service are on keeper_data table
		dbOperationDuration.WithLabelValues(operation, "keeper_data").Record(ctx, duration.Seconds())
	}
}

// RecordTelegramNotification records a telegram notification sent for a keeper
// This is a stub function - telegram notification tracking can be added later if needed
func RecordTelegramNotification(ctx context.Context, keeperAddress string) {
	// TODO: Implement telegram notification metrics if needed
	// This is called from notification code but metrics are not currently defined
	// if telegramKeeperNotificationsSentTotal != nil {
	// 	telegramKeeperNotificationsSentTotal.WithLabelValues(keeperAddress).Inc(ctx)
	// }
}

// UpdateKeepersOnlineByVersion updates the count of keepers online by version
// This should be called whenever keeper status changes (active/inactive)
// It increases/decreases counts and removes versions that go to 0
func UpdateKeepersOnlineByVersion(ctx context.Context, keepersByVersion map[string]int) {
	if keepersOnlineByVersion == nil {
		return
	}

	versionsLock.Lock()
	defer versionsLock.Unlock()

	// Track current versions
	currentVersions := make(map[string]bool)

	// Set the count for each version (increase or decrease as needed)
	for version, count := range keepersByVersion {
		currentVersions[version] = true
		keepersOnlineByVersion.WithLabelValues(version).Set(ctx, float64(count))
		lastKnownVersions[version] = true
	}

	// Reset versions that were previously tracked but now have 0 keepers
	for version := range lastKnownVersions {
		if !currentVersions[version] {
			// This version had keepers before but now has 0, set it to 0
			keepersOnlineByVersion.WithLabelValues(version).Set(ctx, 0)
			delete(lastKnownVersions, version)
		}
	}
}
