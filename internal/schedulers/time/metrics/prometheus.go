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

	// TasksScheduled tracks the total number of tasks scheduled
	TasksScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_scheduled",
		Help:      "Total number of jobs scheduled",
	})

	// TasksRunning tracks the number of tasks currently running
	TasksRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_running",
		Help:      "Total number of jobs currently running",
	})

	// TasksCompleted tracks the total number of tasks completed
	TasksCompleted = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_completed",
		Help:      "Total number of jobs completed",
	})

	// TasksFailed tracks the total number of tasks failed
	TasksFailed = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "tasks_failed",
		Help:      "Total number of jobs failed",
	})

	TaskExecutionTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "triggerx",
		Subsystem: "time_scheduler",
		Name:      "task_execution_time",
		Help:      "Time taken to execute a task",
		Buckets:   []float64{1, 5, 10, 50, 100, 500},
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
