package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		handler: promhttp.Handler(),
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
