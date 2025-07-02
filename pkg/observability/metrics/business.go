package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Business metrics for OpenTelemetry integration
var (
	// Trace collection metrics
	TracesCollected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "triggerx_traces_collected_total",
			Help: "Total number of traces collected by service",
		},
		[]string{"service", "endpoint"},
	)

	// Tracing overhead metrics
	TracingOverheadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "triggerx_tracing_overhead_seconds",
			Help: "Time overhead added by tracing instrumentation",
		},
		[]string{"service", "operation"},
	)

	// Business context metrics
	BusinessSpansCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "triggerx_business_spans_total",
			Help: "Total number of business context spans created",
		},
		[]string{"service", "span_type"},
	)
)

// IncrementTracesCollected increments the traces collected counter
func IncrementTracesCollected(service, endpoint string) {
	TracesCollected.WithLabelValues(service, endpoint).Inc()
}

// ObserveTracingOverhead records the overhead duration for tracing
func ObserveTracingOverhead(service, operation string, duration float64) {
	TracingOverheadDuration.WithLabelValues(service, operation).Observe(duration)
}

// IncrementBusinessSpans increments the business spans counter
func IncrementBusinessSpans(service, spanType string) {
	BusinessSpansCreated.WithLabelValues(service, spanType).Inc()
}
