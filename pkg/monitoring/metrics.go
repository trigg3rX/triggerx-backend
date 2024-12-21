package manager

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


// This is temp file please add your metrics which you want

var (
	uptime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "uptime_seconds",
			Help: "Time for which the node is active in seconds",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)

	successfulExecutions = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "successful_executions_total",
			Help: "Total number of successfully executed tasks",
		},
	)

	totalTasks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_tasks_total",
			Help: "Total number of tasks executed",
		},
		[]string{"status"},
	)

	successRate = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "success_rate",
			Help: "Percentage of successful executions",
		},
		func() float64 {
			total := totalTasks.WithLabelValues("total").Get()
			if total == 0 {
				return 0
			}
			return (successfulExecutions.Get() / total) * 100
		},
	)
)

var startTime time.Time
var totalTaskCount float64

func init() {
	startTime = time.Now()
	prometheus.MustRegister(uptime)
	prometheus.MustRegister(successfulExecutions)
	prometheus.MustRegister(totalTasks)
	prometheus.MustRegister(successRate)
}

// UpdateMetrics updates the metrics based on job execution results
func UpdateMetrics(success bool) {
	status := "failed"
	if success {
		status = "success"
		successfulExecutions.Inc()
	}
	totalTasks.WithLabelValues(status).Inc()
	totalTaskCount++
}

// ExposeMetricsHandler handles the metrics endpoint
func ExposeMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the metrics for Prometheus
	promhttp.Handler().ServeHTTP(w, r)
}
