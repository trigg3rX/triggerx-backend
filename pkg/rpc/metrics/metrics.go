package metrics

import (
	"sync"
	"time"
)

// InMemoryCollector is a simple thread-safe in-memory implementation of the Collector.
// Useful for development/testing or as a default no-op-ish collector.
type InMemoryCollector struct {
	mu sync.RWMutex

	requests map[string]int64
	errors   map[string]int64
	// We store basic aggregation for durations: count and total ns
	durationsCount map[string]int64
	durationsSumNs map[string]int64
}

// NewInMemoryCollector creates a new in-memory metrics collector.
func NewInMemoryCollector() *InMemoryCollector {
	return &InMemoryCollector{
		requests:       make(map[string]int64),
		errors:         make(map[string]int64),
		durationsCount: make(map[string]int64),
		durationsSumNs: make(map[string]int64),
	}
}

func (c *InMemoryCollector) IncRequestsTotal(service string, method string) {
	key := service + "." + method
	c.mu.Lock()
	c.requests[key]++
	c.mu.Unlock()
}

func (c *InMemoryCollector) IncErrorsTotal(service string, method string) {
	key := service + "." + method
	c.mu.Lock()
	c.errors[key]++
	c.mu.Unlock()
}

func (c *InMemoryCollector) ObserveRequestDuration(service string, method string, duration time.Duration) {
	key := service + "." + method
	c.mu.Lock()
	c.durationsCount[key]++
	c.durationsSumNs[key] += duration.Nanoseconds()
	c.mu.Unlock()
}

// Snapshot returns a copy of current counters for external inspection.
func (c *InMemoryCollector) Snapshot() (requests map[string]int64, errors map[string]int64, durationsCount map[string]int64, durationsSumNs map[string]int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	req := make(map[string]int64, len(c.requests))
	err := make(map[string]int64, len(c.errors))
	dcnt := make(map[string]int64, len(c.durationsCount))
	dsum := make(map[string]int64, len(c.durationsSumNs))
	for k, v := range c.requests {
		req[k] = v
	}
	for k, v := range c.errors {
		err[k] = v
	}
	for k, v := range c.durationsCount {
		dcnt[k] = v
	}
	for k, v := range c.durationsSumNs {
		dsum[k] = v
	}
	return req, err, dcnt, dsum
}
