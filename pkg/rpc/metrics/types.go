package metrics

import "time"

// Collector defines the interface for recording RPC metrics.
// Implementations can be backed by Prometheus, OpenTelemetry, or in-memory.
type Collector interface {
	// IncRequestsTotal increments the total number of RPC requests for a method.
	IncRequestsTotal(service string, method string)

	// IncErrorsTotal increments the total number of RPC errors for a method.
	IncErrorsTotal(service string, method string)

	// ObserveRequestDuration records the duration of an RPC request.
	ObserveRequestDuration(service string, method string, duration time.Duration)
}
