package metrics

import "github.com/prometheus/client_golang/prometheus"

// MetricsCollector is the interface for metrics collection
type MetricsCollector interface {
	// Start begins collecting metrics
	Start()

	// Stop stops metrics collection
	Stop()

	// Handler returns the HTTP handler for metrics endpoint
	Handler() interface{}

	// Registry returns the Prometheus registry
	Registry() *prometheus.Registry

	// MustRegister registers custom metrics
	MustRegister(collectors ...prometheus.Collector)
}

// ServiceMetrics is an interface for service-specific metrics
type ServiceMetrics interface {
	// RegisterWith registers the metrics with a collector
	RegisterWith(collector *Collector)
}
