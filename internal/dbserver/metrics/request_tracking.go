package metrics

import (
	"sync"
	"time"
)

var (
	// Request tracking for RPS calculation
	requestCounts     = make(map[string]int64)
	requestCountsLock sync.Mutex
	lastRPSUpdate     = time.Now()

	// Response time tracking for average calculation
	responseTimeSums   = make(map[string]float64)
	responseTimeCounts = make(map[string]int64)
	responseTimeLock   sync.Mutex

	// Active request tracking
	activeRequestCounts     = make(map[string]int64)
	activeRequestCountsLock sync.Mutex
)

// TrackRequest tracks a request for RPS calculation
func TrackRequest(endpoint string) {
	requestCountsLock.Lock()
	requestCounts[endpoint]++
	requestCountsLock.Unlock()
}

// TrackResponseTime tracks response time for average calculation
func TrackResponseTime(endpoint string, duration float64) {
	responseTimeLock.Lock()
	responseTimeSums[endpoint] += duration
	responseTimeCounts[endpoint]++
	responseTimeLock.Unlock()
}

// IncrementActiveRequests increments the active request count for an endpoint
func IncrementActiveRequests(endpoint string) {
	activeRequestCountsLock.Lock()
	activeRequestCounts[endpoint]++
	count := activeRequestCounts[endpoint]
	activeRequestCountsLock.Unlock()

	if ActiveRequests != nil {
		ActiveRequests.WithLabelValues(endpoint).Set(ctx, float64(count))
	}
}

// DecrementActiveRequests decrements the active request count for an endpoint
func DecrementActiveRequests(endpoint string) {
	activeRequestCountsLock.Lock()
	activeRequestCounts[endpoint]--
	if activeRequestCounts[endpoint] < 0 {
		activeRequestCounts[endpoint] = 0
	}
	count := activeRequestCounts[endpoint]
	activeRequestCountsLock.Unlock()

	if ActiveRequests != nil {
		ActiveRequests.WithLabelValues(endpoint).Set(ctx, float64(count))
	}
}

// StartRequestMetricsCollection starts the background goroutine for RPS and average response time calculation
func StartRequestMetricsCollection() {
	// Calculate and update RPS every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			calculateAndUpdateRPS()
			calculateAndUpdateAverageResponseTime()
		}
	}()
}

// calculateAndUpdateRPS calculates requests per second from request counts
func calculateAndUpdateRPS() {
	requestCountsLock.Lock()
	defer requestCountsLock.Unlock()

	now := time.Now()
	timeWindow := now.Sub(lastRPSUpdate).Seconds()
	if timeWindow <= 0 {
		timeWindow = 1.0 // Avoid division by zero
	}

	// Calculate RPS for each endpoint and update metric
	for endpoint, count := range requestCounts {
		rps := float64(count) / timeWindow
		if RequestsPerSecond != nil {
			RequestsPerSecond.WithLabelValues(endpoint).Set(ctx, rps)
		}
	}

	// Reset counts for next window
	for k := range requestCounts {
		requestCounts[k] = 0
	}
	lastRPSUpdate = now
}

// calculateAndUpdateAverageResponseTime calculates average response time
func calculateAndUpdateAverageResponseTime() {
	responseTimeLock.Lock()
	defer responseTimeLock.Unlock()

	for endpoint, sum := range responseTimeSums {
		count := responseTimeCounts[endpoint]
		if count > 0 {
			avg := sum / float64(count)
			if AverageResponseTime != nil {
				AverageResponseTime.WithLabelValues(endpoint).Set(ctx, avg)
			}
		}
	}

	// Reset for next window
	for k := range responseTimeSums {
		responseTimeSums[k] = 0
		responseTimeCounts[k] = 0
	}
}
