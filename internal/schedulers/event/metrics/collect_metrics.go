package metrics

import (
	"runtime"
)

// Collects system resource metrics
func collectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage (current allocated bytes)
	MemoryUsageBytes.Set(float64(memStats.Alloc))

	// Update CPU usage (using system memory as a proxy)
	CPUUsagePercent.Set(float64(memStats.Sys))

	// Update active goroutines count
	GoroutinesActive.Set(float64(runtime.NumGoroutine()))

	// Update garbage collection duration (total pause time in seconds)
	GCDurationSeconds.Set(float64(memStats.PauseTotalNs) / 1e9)
}
