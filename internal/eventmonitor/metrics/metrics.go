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

	// Metrics instances
	uptimeSeconds    observability.Gauge
	memoryUsageBytes observability.Gauge
	cpuUsagePercent  observability.Gauge
	goroutinesActive observability.Gauge
	gcDurationSeconds observability.Gauge

	httpRequestsTotal   *observability.CounterVec
	httpRequestDuration *observability.HistogramVec
	requestsPerSecond   *observability.GaugeVec
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

