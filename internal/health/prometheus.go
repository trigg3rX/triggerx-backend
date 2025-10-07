package health

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
)

// HealthMetrics contains all health service specific metrics
type HealthMetrics struct {
	HTTPRequestsTotal                    *prometheus.CounterVec
	HTTPRequestDuration                  *prometheus.HistogramVec
	CheckinsByVersionTotal               *prometheus.CounterVec
	KeepersTotal                         prometheus.Gauge
	KeepersActiveTotal                   prometheus.Gauge
	KeepersInactiveTotal                 prometheus.Gauge
	KeeperUptimeSeconds                  *prometheus.GaugeVec
	MostActiveKeeperSeconds              *prometheus.CounterVec
	DBHostOperationDuration              *prometheus.HistogramVec
	TelegramKeeperNotificationsSentTotal *prometheus.CounterVec
	NetworkConnectionsTotal              *prometheus.CounterVec
}

// NewHealthMetrics creates and registers health service metrics
func NewHealthMetrics(collector *metrics.Collector) *HealthMetrics {
	builder := metrics.NewMetricBuilder(collector, "health_service")

	hm := &HealthMetrics{
		HTTPRequestsTotal: builder.CounterVec(
			"http_requests_total",
			"Total HTTP requests received",
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: builder.HistogramVec(
			"http_request_duration_seconds",
			"HTTP request processing time",
			[]string{"method", "endpoint"},
			prometheus.DefBuckets,
		),
		CheckinsByVersionTotal: builder.CounterVec(
			"checkins_by_version_total",
			"Check-ins by keeper version",
			[]string{"version"},
		),
		KeepersTotal: builder.Gauge(
			"keepers_total",
			"Total number of registered keepers",
		),
		KeepersActiveTotal: builder.Gauge(
			"keepers_active_total",
			"Currently active keepers",
		),
		KeepersInactiveTotal: builder.Gauge(
			"keepers_inactive_total",
			"Currently inactive keepers",
		),
		KeeperUptimeSeconds: builder.GaugeVec(
			"keeper_uptime_seconds",
			"Keeper uptime since first check-in",
			[]string{"keeper_address"},
		),
		MostActiveKeeperSeconds: builder.CounterVec(
			"most_active_keeper_uptime_seconds",
			"Most active keeper uptime since first check-in",
			[]string{"keeper_address"},
		),
		DBHostOperationDuration: builder.HistogramVec(
			"db_host_operation_duration_seconds",
			"Scylla Database operation execution time",
			[]string{"operation"},
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		),
		TelegramKeeperNotificationsSentTotal: builder.CounterVec(
			"telegram_keeper_notifications_sent_total",
			"Notifications sent per keeper",
			[]string{"keeper_address"},
		),
		NetworkConnectionsTotal: builder.CounterVec(
			"network_connections_total",
			"Network connections (type=incoming/outgoing)",
			[]string{"type"},
		),
	}

	return hm
}
