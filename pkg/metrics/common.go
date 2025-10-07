package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
)

// CommonMetrics contains metrics common to all services
type CommonMetrics struct {
	startTime         time.Time
	UptimeSeconds     prometheus.Gauge
	MemoryUsageBytes  prometheus.Gauge
	CPUUsagePercent   prometheus.Gauge
	GoroutinesActive  prometheus.Gauge
	GCDurationSeconds prometheus.Gauge
}

// newCommonMetrics creates common metrics and registers them
func newCommonMetrics(namespace, subsystem string, registry *prometheus.Registry) *CommonMetrics {
	cm := &CommonMetrics{
		startTime: time.Now(),

		UptimeSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "uptime_seconds",
			Help:      "Time passed since service started in seconds",
		}),

		MemoryUsageBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_usage_bytes",
			Help:      "Service memory consumption in bytes",
		}),

		CPUUsagePercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cpu_usage_percent",
			Help:      "CPU utilization percentage",
		}),

		GoroutinesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "goroutines_active",
			Help:      "Number of active goroutines",
		}),

		GCDurationSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "gc_duration_seconds",
			Help:      "Total garbage collection pause duration in seconds",
		}),
	}

	// Register all metrics
	registry.MustRegister(
		cm.UptimeSeconds,
		cm.MemoryUsageBytes,
		cm.CPUUsagePercent,
		cm.GoroutinesActive,
		cm.GCDurationSeconds,
	)

	return cm
}

// UpdateUptime updates the uptime metric
func (cm *CommonMetrics) UpdateUptime() {
	cm.UptimeSeconds.Set(time.Since(cm.startTime).Seconds())
}

// UpdateSystemMetrics updates all system metrics
func (cm *CommonMetrics) UpdateSystemMetrics() {
	// Memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	cm.MemoryUsageBytes.Set(float64(m.Alloc))
	cm.GoroutinesActive.Set(float64(runtime.NumGoroutine()))
	cm.GCDurationSeconds.Set(float64(m.PauseTotalNs) / 1e9)

	// CPU usage
	cpuPercentages, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercentages) > 0 {
		cm.CPUUsagePercent.Set(cpuPercentages[0])
	}
}
