package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// otelMetrics wraps OpenTelemetry metrics implementation
type otelMetrics struct {
	meterProvider *sdkmetric.MeterProvider
	meter         otelmetric.Meter
}

// NewMetrics creates a new metrics instance with the provided configuration and resource
// Each service can call this to create their own metrics instance and define custom metrics
func NewMetrics(cfg Config, res *resource.Resource) (Metrics, func(context.Context) error, error) {
	// Create OTLP HTTP metric exporter
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(cfg.OTELExporterEndpoint),
		otlpmetrichttp.WithInsecure(), // TODO: make this configurable
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create meter provider with periodic reader
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(cfg.BatchTimeout),
				sdkmetric.WithTimeout(cfg.ExportTimeout),
			),
		),
	)

	// Create meter from provider - each service can create their own meter
	// with a unique name to namespace their metrics
	meter := meterProvider.Meter(
		string(cfg.ServiceName),
		otelmetric.WithInstrumentationVersion(cfg.ServiceVersion),
		otelmetric.WithSchemaURL("https://opentelemetry.io/schemas/1.27.0"),
	)

	shutdown := func(ctx context.Context) error {
		return meterProvider.Shutdown(ctx)
	}

	return &otelMetrics{
		meterProvider: meterProvider,
		meter:         meter,
	}, shutdown, nil
}

// Counter creates a new counter metric instrument
// Each service can define their own counters with custom names and options
func (m *otelMetrics) Counter(name string, opts ...InstrumentOption) Counter {
	metricOpts := make([]otelmetric.Float64CounterOption, 0, len(opts))
	for _, opt := range opts {
		if metricOpt, ok := opt.(instrumentOption); ok {
			metricOpts = append(metricOpts, metricOpt.opt.(otelmetric.Float64CounterOption))
		}
	}

	otelCounter, err := m.meter.Float64Counter(name, metricOpts...)
	if err != nil {
		// Return a no-op counter if creation fails
		// In production, you might want to log this error
		return &noOpCounter{}
	}

	return &counter{counter: otelCounter}
}

// Gauge creates a new gauge metric instrument
// Uses UpDownCounter internally to allow direct value setting
func (m *otelMetrics) Gauge(name string, opts ...InstrumentOption) Gauge {
	metricOpts := make([]otelmetric.Float64UpDownCounterOption, 0, len(opts))
	for _, opt := range opts {
		if metricOpt, ok := opt.(instrumentOption); ok {
			metricOpts = append(metricOpts, metricOpt.opt.(otelmetric.Float64UpDownCounterOption))
		}
	}

	upDownCounter, err := m.meter.Float64UpDownCounter(name, metricOpts...)
	if err != nil {
		return &noOpGauge{}
	}

	return &gauge{upDownCounter: upDownCounter, lastValue: 0}
}

// Histogram creates a new histogram metric instrument
func (m *otelMetrics) Histogram(name string, opts ...InstrumentOption) Histogram {
	metricOpts := make([]otelmetric.Float64HistogramOption, 0, len(opts))
	for _, opt := range opts {
		if metricOpt, ok := opt.(instrumentOption); ok {
			metricOpts = append(metricOpts, metricOpt.opt.(otelmetric.Float64HistogramOption))
		}
	}

	otelHistogram, err := m.meter.Float64Histogram(name, metricOpts...)
	if err != nil {
		return &noOpHistogram{}
	}

	return &histogram{histogram: otelHistogram}
}

// GetMeterProvider returns the underlying meter provider (for advanced use cases)
func (m *otelMetrics) GetMeterProvider() *sdkmetric.MeterProvider {
	return m.meterProvider
}

// GetMeter returns the underlying meter (for creating custom instruments)
func (m *otelMetrics) GetMeter() otelmetric.Meter {
	return m.meter
}

// noOpCounter is a no-op counter implementation for error cases
type noOpCounter struct{}

func (n *noOpCounter) Add(ctx context.Context, value float64, attrs ...attribute.KeyValue) {}
func (n *noOpCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue)                {}

// noOpGauge is a no-op gauge implementation for error cases
type noOpGauge struct{}

func (n *noOpGauge) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {}
func (n *noOpGauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue)    {}

// noOpHistogram is a no-op histogram implementation for error cases
type noOpHistogram struct{}

func (n *noOpHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {}
