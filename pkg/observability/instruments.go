package observability

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// counter wraps OpenTelemetry counter metric
type counter struct {
	counter otelmetric.Float64Counter
}

// Add adds the given value to the counter
func (c *counter) Add(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	c.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

// Inc increments the counter by 1
func (c *counter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	c.counter.Add(ctx, 1, otelmetric.WithAttributes(attrs...))
}

// gauge wraps OpenTelemetry up-down counter metric (used as gauge)
// OpenTelemetry's ObservableGauge uses callbacks, so we use UpDownCounter
// which allows direct value setting, making it suitable for gauge-like behavior
// Note: This implementation tracks the last value to calculate deltas.
// For concurrent updates, consider using proper synchronization or ObservableGauge with callbacks.
type gauge struct {
	upDownCounter otelmetric.Float64UpDownCounter
	lastValue     float64
}

// Record records a value for the gauge
// This uses an UpDownCounter internally to allow direct value setting
// The value is treated as an absolute value, and the delta from the last value is recorded
func (g *gauge) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	// Calculate the difference from the last value
	// Note: This is not thread-safe. For concurrent access, use proper synchronization
	delta := value - g.lastValue
	g.lastValue = value
	g.upDownCounter.Add(ctx, delta, otelmetric.WithAttributes(attrs...))
}

// Set sets the gauge to a specific value (convenience method for Prometheus compatibility)
func (g *gauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	g.Record(ctx, value, attrs...)
}

// histogram wraps OpenTelemetry histogram metric
type histogram struct {
	histogram otelmetric.Float64Histogram
}

// Record records a value in the histogram
func (h *histogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	h.histogram.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}

// instrumentOption wraps otelmetric.InstrumentOption
type instrumentOption struct {
	opt otelmetric.InstrumentOption
}

func (o instrumentOption) applyInstrumentOption() {}

// WithDescription creates an instrument option for setting the description
func WithDescription(description string) InstrumentOption {
	return instrumentOption{opt: otelmetric.WithDescription(description)}
}

// WithUnit creates an instrument option for setting the unit
func WithUnit(unit string) InstrumentOption {
	return instrumentOption{opt: otelmetric.WithUnit(unit)}
}

// CounterVec is a labeled counter (similar to Prometheus CounterVec)
// It allows creating counters with predefined label names
type CounterVec struct {
	metrics   Metrics
	name      string
	labelKeys []string
	opts      []InstrumentOption
}

// NewCounterVec creates a new CounterVec with the given label names
func NewCounterVec(metrics Metrics, name string, labelKeys []string, opts ...InstrumentOption) *CounterVec {
	return &CounterVec{
		metrics:   metrics,
		name:      name,
		labelKeys: labelKeys,
		opts:      opts,
	}
}

// WithLabelValues returns a Counter with the given label values
func (cv *CounterVec) WithLabelValues(labelValues ...string) Counter {
	if len(labelValues) != len(cv.labelKeys) {
		// Return no-op counter if label count mismatch
		return &noOpCounter{}
	}

	attrs := make([]attribute.KeyValue, len(cv.labelKeys))
	for i, key := range cv.labelKeys {
		attrs[i] = attribute.String(key, labelValues[i])
	}

	// Create a counter with attributes embedded in the name or use a wrapper
	// For OpenTelemetry, we'll create a counter and store attributes separately
	// We'll need to pass attributes on each call
	counter := cv.metrics.Counter(cv.name, cv.opts...)
	return &labeledCounter{
		counter: counter,
		attrs:   attrs,
	}
}

// labeledCounter wraps a Counter with fixed attributes
type labeledCounter struct {
	counter Counter
	attrs   []attribute.KeyValue
}

func (lc *labeledCounter) Add(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	allAttrs := append(lc.attrs, attrs...)
	lc.counter.Add(ctx, value, allAttrs...)
}

func (lc *labeledCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	allAttrs := append(lc.attrs, attrs...)
	lc.counter.Inc(ctx, allAttrs...)
}

// GaugeVec is a labeled gauge (similar to Prometheus GaugeVec)
type GaugeVec struct {
	metrics   Metrics
	name      string
	labelKeys []string
	opts      []InstrumentOption
	cache     map[string]Gauge // Cache gauge instances per label combination
}

// NewGaugeVec creates a new GaugeVec with the given label names
func NewGaugeVec(metrics Metrics, name string, labelKeys []string, opts ...InstrumentOption) *GaugeVec {
	return &GaugeVec{
		metrics:   metrics,
		name:      name,
		labelKeys: labelKeys,
		opts:      opts,
		cache:     make(map[string]Gauge),
	}
}

// WithLabelValues returns a Gauge with the given label values
func (gv *GaugeVec) WithLabelValues(labelValues ...string) Gauge {
	if len(labelValues) != len(gv.labelKeys) {
		return &noOpGauge{}
	}

	// Create cache key from label values
	cacheKey := strings.Join(labelValues, "\x00") // Use null byte as separator

	// Return cached gauge if it exists
	if cached, ok := gv.cache[cacheKey]; ok {
		return cached
	}

	attrs := make([]attribute.KeyValue, len(gv.labelKeys))
	for i, key := range gv.labelKeys {
		attrs[i] = attribute.String(key, labelValues[i])
	}

	gauge := gv.metrics.Gauge(gv.name, gv.opts...)
	labeledGauge := &labeledGauge{
		gauge: gauge,
		attrs: attrs,
	}

	// Cache the gauge instance
	gv.cache[cacheKey] = labeledGauge

	return labeledGauge
}

// labeledGauge wraps a Gauge with fixed attributes
type labeledGauge struct {
	gauge Gauge
	attrs []attribute.KeyValue
}

func (lg *labeledGauge) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	allAttrs := append(lg.attrs, attrs...)
	lg.gauge.Record(ctx, value, allAttrs...)
}

func (lg *labeledGauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	allAttrs := append(lg.attrs, attrs...)
	lg.gauge.Set(ctx, value, allAttrs...)
}

// HistogramVec is a labeled histogram (similar to Prometheus HistogramVec)
type HistogramVec struct {
	metrics   Metrics
	name      string
	labelKeys []string
	opts      []InstrumentOption
}

// NewHistogramVec creates a new HistogramVec with the given label names
func NewHistogramVec(metrics Metrics, name string, labelKeys []string, opts ...InstrumentOption) *HistogramVec {
	return &HistogramVec{
		metrics:   metrics,
		name:      name,
		labelKeys: labelKeys,
		opts:      opts,
	}
}

// WithLabelValues returns a Histogram with the given label values
func (hv *HistogramVec) WithLabelValues(labelValues ...string) Histogram {
	if len(labelValues) != len(hv.labelKeys) {
		return &noOpHistogram{}
	}

	attrs := make([]attribute.KeyValue, len(hv.labelKeys))
	for i, key := range hv.labelKeys {
		attrs[i] = attribute.String(key, labelValues[i])
	}

	histogram := hv.metrics.Histogram(hv.name, hv.opts...)
	return &labeledHistogram{
		histogram: histogram,
		attrs:     attrs,
	}
}

// labeledHistogram wraps a Histogram with fixed attributes
type labeledHistogram struct {
	histogram Histogram
	attrs     []attribute.KeyValue
}

func (lh *labeledHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	allAttrs := append(lh.attrs, attrs...)
	lh.histogram.Record(ctx, value, allAttrs...)
}
