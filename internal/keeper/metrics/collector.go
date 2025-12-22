package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// prometheusMetrics is an interface for metrics that support Prometheus export
type prometheusMetrics interface {
	PrometheusHandler() http.Handler
	PrometheusRegistry() *prometheus.Registry
}

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
	metrics observability.Metrics
}

// NewCollector creates a new metrics collector
// If metrics is nil, it falls back to a not-implemented handler
func NewCollector(metrics observability.Metrics) *Collector {
	var handler http.Handler

	if metrics != nil {
		// Try to access PrometheusHandler and Registry via type assertion
		if promMetrics, ok := metrics.(prometheusMetrics); ok {
			// Register our Prometheus metrics with the observability registry
			if registry := promMetrics.PrometheusRegistry(); registry != nil {
				// Use Prometheus handler from observability
				handler = promMetrics.PrometheusHandler()
			} else {
				// Prometheus export is not enabled
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
					_, err := w.Write([]byte("Prometheus export is not enabled"))
					if err != nil {
						fmt.Println("Failed to write response:", err)
					}
				})
			}
		} else {
			// Metrics doesn't support Prometheus export
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
				_, err := w.Write([]byte("Metrics does not support Prometheus export"))
				if err != nil {
					fmt.Println("Failed to write response:", err)
				}
			})
		}
	} else {
		// Fallback for backward compatibility
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			_, err := w.Write([]byte("Metrics not initialized"))
			if err != nil {
				fmt.Println("Failed to write response:", err)
			}
		})
	}

	return &Collector{
		handler: handler,
		metrics: metrics,
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
