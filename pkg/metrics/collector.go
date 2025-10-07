package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector manages metrics collection for a service
type Collector struct {
	serviceName   string
	namespace     string
	registry      *prometheus.Registry
	commonMetrics *CommonMetrics
	handler       http.Handler
	stopCh        chan struct{}
	wg            sync.WaitGroup
	options       CollectorOptions
}

// NewCollector creates a new metrics collector for a service
func NewCollector(serviceName string, opts ...Option) *Collector {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	registry := prometheus.NewRegistry()

	collector := &Collector{
		serviceName: serviceName,
		namespace:   options.Namespace,
		registry:    registry,
		stopCh:      make(chan struct{}),
		options:     options,
	}

	// Initialize common metrics if enabled
	if options.EnableCommonMetrics {
		collector.commonMetrics = newCommonMetrics(options.Namespace, serviceName, registry)
	}

	// Create HTTP handler
	collector.handler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	})

	return collector
}

// Start begins collecting metrics
func (c *Collector) Start() {
	if c.options.EnableCommonMetrics && c.commonMetrics != nil {
		c.startCommonMetricsCollection()
	}
}

// Stop stops metrics collection
func (c *Collector) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return c.handler
}

// Registry returns the Prometheus registry for custom metrics
func (c *Collector) Registry() *prometheus.Registry {
	return c.registry
}

// Common returns common metrics for direct access
func (c *Collector) Common() *CommonMetrics {
	return c.commonMetrics
}

// startCommonMetricsCollection starts background collection
func (c *Collector) startCommonMetricsCollection() {
	// Update uptime
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		interval := c.options.UptimeUpdateInterval
		if interval == 0 {
			return
		}
		ticker := newTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.commonMetrics.UpdateUptime()
			case <-c.stopCh:
				return
			}
		}
	}()

	// Update system metrics
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		interval := c.options.SystemMetricsInterval
		if interval == 0 {
			return
		}
		ticker := newTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.commonMetrics.UpdateSystemMetrics()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// MustRegister registers custom metrics with the collector's registry
// Panics if registration fails
func (c *Collector) MustRegister(collectors ...prometheus.Collector) {
	c.registry.MustRegister(collectors...)
}

// Register registers custom metrics with the collector's registry
// Returns an error if registration fails
func (c *Collector) Register(collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		if err := c.registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}

// newTicker creates a new time.Ticker wrapper
func newTicker(d interface{}) *timeTicker {
	switch v := d.(type) {
	case int64:
		return &timeTicker{Ticker: time.NewTicker(time.Duration(v))}
	default:
		return &timeTicker{Ticker: time.NewTicker(d.(time.Duration))}
	}
}

type timeTicker struct {
	*time.Ticker
}

func (t *timeTicker) Stop() {
	if t.Ticker != nil {
		t.Ticker.Stop()
	}
}
