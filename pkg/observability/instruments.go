package observability

import (
	"context"

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
