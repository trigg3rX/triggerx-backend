package metrics

import (
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		handler: promhttp.Handler(),
	}
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return c.handler
}

// Start starts metrics collection
func (c *Collector) Start() {
	// Update uptime every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
			UpdateSystemMetrics()
		}
	}()
}

// UpdateSystemMetrics updates system metrics (similar to keeper's middleware)
func UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	MemoryUsageBytes.Set(float64(memStats.Alloc))
	CPUUsagePercent.Set(float64(memStats.Sys))
	GoroutinesActive.Set(float64(runtime.NumGoroutine()))
	GCDurationSeconds.Set(float64(memStats.PauseTotalNs) / 1e9)
}
