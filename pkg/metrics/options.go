package metrics

import "time"

// CollectorOptions configures the metrics collector
type CollectorOptions struct {
	Namespace             string
	EnableCommonMetrics   bool
	UptimeUpdateInterval  time.Duration
	SystemMetricsInterval time.Duration
}

// Option is a functional option for configuring the collector
type Option func(*CollectorOptions)

// defaultOptions returns default collector options
func defaultOptions() CollectorOptions {
	return CollectorOptions{
		Namespace:             "triggerx",
		EnableCommonMetrics:   true,
		UptimeUpdateInterval:  15 * time.Second,
		SystemMetricsInterval: 30 * time.Second,
	}
}

// WithNamespace sets a custom namespace
func WithNamespace(namespace string) Option {
	return func(o *CollectorOptions) {
		o.Namespace = namespace
	}
}

// WithCommonMetrics enables or disables common metrics collection
func WithCommonMetrics(enable bool) Option {
	return func(o *CollectorOptions) {
		o.EnableCommonMetrics = enable
	}
}

// WithUptimeInterval sets the uptime update interval
func WithUptimeInterval(interval time.Duration) Option {
	return func(o *CollectorOptions) {
		o.UptimeUpdateInterval = interval
	}
}

// WithSystemMetricsInterval sets the system metrics update interval
func WithSystemMetricsInterval(interval time.Duration) Option {
	return func(o *CollectorOptions) {
		o.SystemMetricsInterval = interval
	}
}
