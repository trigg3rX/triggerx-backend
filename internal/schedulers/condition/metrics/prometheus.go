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
		Subsystem: "condition_scheduler",
		Name:      "uptime_seconds",
		Help:      "The uptime of the condition scheduler service in seconds",
	})

	// JobsScheduled tracks the total number of jobs scheduled
	JobsScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "jobs_scheduled",
		Help:      "Total number of jobs scheduled",
	})

	// JobsRunning tracks the number of jobs currently running
	JobsRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "jobs_running",
		Help:      "Total number of jobs currently running",
	})

	// JobsCompleted tracks the total number of jobs completed
	JobsCompleted = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "jobs_completed",
		Help:      "Total number of jobs completed",
	})

	// JobsFailed tracks the total number of jobs failed
	JobsFailed = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "jobs_failed",
		Help:      "Total number of jobs failed",
	})

	// ConditionsChecked tracks the total number of condition checks performed
	ConditionsChecked = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "conditions_checked_total",
		Help:      "Total number of condition checks performed",
	})

	// ConditionsSatisfied tracks the total number of conditions satisfied
	ConditionsSatisfied = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "conditions_satisfied_total",
		Help:      "Total number of conditions satisfied",
	})

	// ValueSourceRequests tracks the total number of value source API requests
	ValueSourceRequests = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "value_source_requests_total",
		Help:      "Total number of value source API requests made",
	})

	// ValueSourceErrors tracks the total number of value source request errors
	ValueSourceErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "triggerx",
		Subsystem: "condition_scheduler",
		Name:      "value_source_errors_total",
		Help:      "Total number of value source request errors",
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
