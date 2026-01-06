package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// Request tracking for RPS calculation
	requestCounts     = make(map[string]int64) // endpoint -> count
	requestCountsLock sync.Mutex
	lastRPSUpdate     = time.Now()

	// System Metrics
	uptimeSeconds     observability.Gauge
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// HTTP Metrics
	httpRequestsTotal   *observability.CounterVec
	httpRequestDuration *observability.HistogramVec
	requestsPerSecond   *observability.GaugeVec

	// Event Monitoring Core Metrics
	activeWorkersTotal     observability.Gauge
	blocksPolledTotal      *observability.CounterVec
	eventsDetectedTotal    *observability.CounterVec
	eventsProcessedTotal   *observability.CounterVec
	blockLag               *observability.GaugeVec
	pollDurationSeconds    *observability.HistogramVec
	pollErrorsTotal        *observability.CounterVec
	subscribersTotal       observability.Gauge
	notificationsSentTotal *observability.CounterVec
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if uptimeSeconds != nil {
				uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
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
			if memoryUsageBytes != nil {
				memoryUsageBytes.Set(ctx, float64(m.Alloc))
			}
			if goroutinesActive != nil {
				goroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
			}
			if gcDurationSeconds != nil {
				gcDurationSeconds.Set(ctx, float64(m.PauseTotalNs)/1e9) // Convert nanoseconds to seconds
			}

			// CPU usage metrics
			cpuPercentages, err := cpu.Percent(0, false)
			if err == nil && len(cpuPercentages) > 0 {
				if cpuUsagePercent != nil {
					cpuUsagePercent.Set(ctx, cpuPercentages[0])
				}
			} else {
				// Fallback to 0.0 if CPU monitoring fails
				if cpuUsagePercent != nil {
					cpuUsagePercent.Set(ctx, 0.0)
				}
			}
		}
	}()

	// Calculate and update requests per second every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
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
		"triggerx.eventmonitor_service.uptime_seconds",
		observability.WithDescription("Time passed since Event Monitor Service started in seconds"),
		observability.WithUnit("s"),
	)

	// Labeled metrics (Vec types)
	httpRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.http_requests_total",
		[]string{"method", "endpoint", "status_code"},
		observability.WithDescription("Total HTTP requests received"),
	)

	httpRequestDuration = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.eventmonitor_service.http_request_duration_seconds",
		[]string{"method", "endpoint"},
		observability.WithDescription("HTTP request processing time"),
		observability.WithUnit("s"),
	)

	requestsPerSecond = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.eventmonitor_service.requests_per_second",
		[]string{"endpoint"},
		observability.WithDescription("Request throughput rate"),
		observability.WithUnit("1/s"),
	)

	// Performance metrics
	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	// Event Monitoring Core Metrics
	activeWorkersTotal = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.active_workers_total",
		observability.WithDescription("Number of active event polling workers"),
	)

	blocksPolledTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.blocks_polled_total",
		[]string{"chain_id"},
		observability.WithDescription("Total number of blocks polled for events"),
	)

	eventsDetectedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.events_detected_total",
		[]string{"chain_id", "event_type"},
		observability.WithDescription("Total number of blockchain events detected"),
	)

	eventsProcessedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.events_processed_total",
		[]string{"chain_id", "status"},
		observability.WithDescription("Total number of events processed (success/error)"),
	)

	blockLag = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.eventmonitor_service.block_lag",
		[]string{"chain_id"},
		observability.WithDescription("Number of blocks behind the chain head"),
	)

	pollDurationSeconds = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.eventmonitor_service.poll_duration_seconds",
		[]string{"chain_id"},
		observability.WithDescription("Duration of event polling operations"),
		observability.WithUnit("s"),
	)

	pollErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.poll_errors_total",
		[]string{"chain_id", "error_type"},
		observability.WithDescription("Total number of polling errors"),
	)

	subscribersTotal = obsMetrics.Gauge(
		"triggerx.eventmonitor_service.subscribers_total",
		observability.WithDescription("Total number of event subscribers"),
	)

	notificationsSentTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.eventmonitor_service.notifications_sent_total",
		[]string{"chain_id", "status"},
		observability.WithDescription("Total number of event notifications sent"),
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

// UpdateActiveWorkers updates the count of active event polling workers
func UpdateActiveWorkers(count int) {
	if activeWorkersTotal != nil {
		activeWorkersTotal.Set(ctx, float64(count))
	}
}

// TrackBlocksPolled tracks blocks polled for a chain
func TrackBlocksPolled(chainID string, count int) {
	if blocksPolledTotal != nil {
		blocksPolledTotal.WithLabelValues(chainID).Add(ctx, float64(count))
	}
}

// TrackEventDetected tracks a detected blockchain event
func TrackEventDetected(chainID, eventType string) {
	if eventsDetectedTotal != nil {
		eventsDetectedTotal.WithLabelValues(chainID, eventType).Inc(ctx)
	}
}

// TrackEventProcessed tracks a processed event with status
func TrackEventProcessed(chainID string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	if eventsProcessedTotal != nil {
		eventsProcessedTotal.WithLabelValues(chainID, status).Inc(ctx)
	}
}

// UpdateBlockLag updates the block lag for a chain
func UpdateBlockLag(chainID string, lag uint64) {
	if blockLag != nil {
		blockLag.WithLabelValues(chainID).Set(ctx, float64(lag))
	}
}

// TrackPollDuration tracks the duration of a poll operation
func TrackPollDuration(chainID string, duration time.Duration) {
	if pollDurationSeconds != nil {
		pollDurationSeconds.WithLabelValues(chainID).Record(ctx, duration.Seconds())
	}
}

// TrackPollError tracks a polling error
func TrackPollError(chainID, errorType string) {
	if pollErrorsTotal != nil {
		pollErrorsTotal.WithLabelValues(chainID, errorType).Inc(ctx)
	}
}

// UpdateSubscribersTotal updates the total number of subscribers
func UpdateSubscribersTotal(count int) {
	if subscribersTotal != nil {
		subscribersTotal.Set(ctx, float64(count))
	}
}

// TrackNotificationSent tracks a sent notification
func TrackNotificationSent(chainID string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	if notificationsSentTotal != nil {
		notificationsSentTotal.WithLabelValues(chainID, status).Inc(ctx)
	}
}
