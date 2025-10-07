package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// NewCounter creates a new Counter metric with the given options
func NewCounter(namespace, subsystem, name, help string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

// NewCounterVec creates a new CounterVec metric with the given options
func NewCounterVec(namespace, subsystem, name, help string, labels []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

// NewGauge creates a new Gauge metric with the given options
func NewGauge(namespace, subsystem, name, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

// NewGaugeVec creates a new GaugeVec metric with the given options
func NewGaugeVec(namespace, subsystem, name, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

// NewHistogram creates a new Histogram metric with the given options
func NewHistogram(namespace, subsystem, name, help string, buckets []float64) prometheus.Histogram {
	opts := prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
	if buckets != nil {
		opts.Buckets = buckets
	} else {
		opts.Buckets = prometheus.DefBuckets
	}
	return prometheus.NewHistogram(opts)
}

// NewHistogramVec creates a new HistogramVec metric with the given options
func NewHistogramVec(namespace, subsystem, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	opts := prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
	if buckets != nil {
		opts.Buckets = buckets
	} else {
		opts.Buckets = prometheus.DefBuckets
	}
	return prometheus.NewHistogramVec(opts, labels)
}

// NewSummary creates a new Summary metric with the given options
func NewSummary(namespace, subsystem, name, help string) prometheus.Summary {
	return prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

// NewSummaryVec creates a new SummaryVec metric with the given options
func NewSummaryVec(namespace, subsystem, name, help string, labels []string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

// MetricBuilder provides a fluent interface for building metrics
type MetricBuilder struct {
	namespace string
	subsystem string
	collector *Collector
}

// NewMetricBuilder creates a new metric builder
func NewMetricBuilder(collector *Collector, subsystem string) *MetricBuilder {
	return &MetricBuilder{
		namespace: collector.namespace,
		subsystem: subsystem,
		collector: collector,
	}
}

// Counter creates and registers a new counter
func (mb *MetricBuilder) Counter(name, help string) prometheus.Counter {
	counter := NewCounter(mb.namespace, mb.subsystem, name, help)
	mb.collector.MustRegister(counter)
	return counter
}

// CounterVec creates and registers a new counter vector
func (mb *MetricBuilder) CounterVec(name, help string, labels []string) *prometheus.CounterVec {
	counter := NewCounterVec(mb.namespace, mb.subsystem, name, help, labels)
	mb.collector.MustRegister(counter)
	return counter
}

// Gauge creates and registers a new gauge
func (mb *MetricBuilder) Gauge(name, help string) prometheus.Gauge {
	gauge := NewGauge(mb.namespace, mb.subsystem, name, help)
	mb.collector.MustRegister(gauge)
	return gauge
}

// GaugeVec creates and registers a new gauge vector
func (mb *MetricBuilder) GaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := NewGaugeVec(mb.namespace, mb.subsystem, name, help, labels)
	mb.collector.MustRegister(gauge)
	return gauge
}

// Histogram creates and registers a new histogram
func (mb *MetricBuilder) Histogram(name, help string, buckets []float64) prometheus.Histogram {
	hist := NewHistogram(mb.namespace, mb.subsystem, name, help, buckets)
	mb.collector.MustRegister(hist)
	return hist
}

// HistogramVec creates and registers a new histogram vector
func (mb *MetricBuilder) HistogramVec(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	hist := NewHistogramVec(mb.namespace, mb.subsystem, name, help, labels, buckets)
	mb.collector.MustRegister(hist)
	return hist
}
