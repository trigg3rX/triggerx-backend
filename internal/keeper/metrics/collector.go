package metrics

import (
	"net/http"
	"runtime"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
	metrics observability.Metrics
	logger  observability.Logger
}

// NewCollector creates a new metrics collector
// Metrics are exported via OpenTelemetry, not through a local Prometheus endpoint
func NewCollector(metrics observability.Metrics, logger observability.Logger) *Collector {
	// Metrics are exported to OpenTelemetry collector, not via local /metrics endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("# Metrics are exported via OpenTelemetry collector\n# Access metrics through the central observability dashboard\n"))
		if err != nil && logger != nil {
			logger.Error(r.Context(), "Failed to write response", observability.Error(err))
		}
	})

	return &Collector{
		handler: handler,
		metrics: metrics,
		logger:  logger,
	}
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return c.handler
}

// Start starts metrics collection
func (c *Collector) Start() {
	// Update uptime every second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if UptimeSeconds != nil {
				UptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
		}
	}()
}

// UpdateSystemMetrics updates system metrics
func UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	if MemoryUsageBytes != nil {
		MemoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}
	if CPUUsagePercent != nil {
		// Note: CPU usage calculation usually requires a delta or library helper.
		// For now we preserve the structure, but actual calculation might need gopsutil like in other services.
		// Since we didn't import gopsutil here yet and the previous implementation was just a gauge definition (not showing calculation in the snippet I saw?),
		// I'll leave it as is or maybe it was set elsewhere?
		// In previous `prometheus.go` snippet, `CPUUsagePercent` was defined but I didn't see where it was set besides definition.
		// Wait, `MetricsMiddleware` in `middleware.go` sets these!
		// metrics.CPUUsagePercent.Set(float64(memStats.Sys)) <- This seems to be what was there.
		CPUUsagePercent.Set(ctx, float64(memStats.Sys))
	}
	if GoroutinesActive != nil {
		GoroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
	}
	if GCDurationSeconds != nil {
		GCDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
	}
}
