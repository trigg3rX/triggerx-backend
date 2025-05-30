package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	startTime = time.Now()

	// UptimeSeconds tracks the service uptime in seconds
	UptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "uptime_seconds",
		Help:      "The uptime of the time scheduler service in seconds",
	})

	// JobsScheduled tracks the total number of jobs scheduled
	JobsScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "jobs_scheduled",
		Help:      "Total number of jobs scheduled",
	})

	// JobsRunning tracks the number of jobs currently running
	JobsRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "jobs_running",
		Help:      "Total number of jobs currently running",
	})

	// JobsCompleted tracks the total number of jobs completed
	JobsCompleted = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "jobs_completed",
		Help:      "Total number of jobs completed",
	})

	// JobsFailed tracks the total number of jobs failed
	JobsFailed = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "jobs_failed",
		Help:      "Total number of jobs failed",
	})
)

// StartMetricsCollection starts collecting metrics
func StartMetricsCollection() {
	// Update uptime every 60 seconds
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UptimeSeconds.Set(time.Since(startTime).Seconds())
		}
	}()
}
