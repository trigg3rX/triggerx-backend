package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// prometheusMetrics is an interface for metrics that support Prometheus export
type prometheusMetrics interface {
	PrometheusHandler() http.Handler
	PrometheusRegistry() *prometheus.Registry
}

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
	metrics observability.Metrics
}

// NewCollector creates a new metrics collector
// If metrics is nil, it falls back to a not-implemented handler
func NewCollector(metrics observability.Metrics) *Collector {
	var handler http.Handler

	if metrics != nil {
		// Try to access PrometheusHandler and Registry via type assertion
		if promMetrics, ok := metrics.(prometheusMetrics); ok {
			// Register our Prometheus metrics with the observability registry
			if registry := promMetrics.PrometheusRegistry(); registry != nil {
				if err := RegisterMetrics(registry); err != nil {
					// If registration fails, return error handler
					handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
						_, err := w.Write([]byte("Failed to register metrics: " + err.Error()))
						if err != nil {
							fmt.Println("Failed to write response:", err)
						}
					})
				} else {
					// Use Prometheus handler from observability
					handler = promMetrics.PrometheusHandler()
				}
			} else {
				// Prometheus export is not enabled
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
					_, err := w.Write([]byte("Prometheus export is not enabled"))
					if err != nil {
						fmt.Println("Failed to write response:", err)
					}
				})
			}
		} else {
			// Metrics doesn't support Prometheus export
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
				_, err := w.Write([]byte("Metrics does not support Prometheus export"))
				if err != nil {
					fmt.Println("Failed to write response:", err)
				}
			})
		}
	} else {
		// Fallback for backward compatibility
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			_, err := w.Write([]byte("Metrics not initialized"))
			if err != nil {
				fmt.Println("Failed to write response:", err)
			}
		})
	}

	return &Collector{
		handler: handler,
		metrics: metrics,
	}
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return c.handler
}

// Start starts metrics collection
func (c *Collector) Start() {
	StartMetricsCollection()
}
