package metrics

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// StartSystemMetricsCollection starts collecting system metrics
func StartSystemMetricsCollection() {
	// Update uptime every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
		}
	}()

	// Update memory and CPU usage every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Memory usage
			if vmStat, err := mem.VirtualMemory(); err == nil {
				MemoryUsageBytes.Set(float64(vmStat.Used))
			}

			// CPU usage
			if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
				CPUUsagePercent.Set(cpuPercent[0])
			}

			// Goroutines count
			GoroutinesActive.Set(float64(runtime.NumGoroutine()))
		}
	}()

	// Update GC stats every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		var lastGCStats runtime.MemStats
		runtime.ReadMemStats(&lastGCStats)

		for range ticker.C {
			var currentGCStats runtime.MemStats
			runtime.ReadMemStats(&currentGCStats)

			// Calculate GC duration
			gcDuration := float64(currentGCStats.PauseTotalNs-lastGCStats.PauseTotalNs) / float64(time.Second)
			GCDurationSeconds.Set(gcDuration)

			lastGCStats = currentGCStats
		}
	}()
}
