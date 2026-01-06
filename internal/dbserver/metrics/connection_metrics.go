package metrics

import (
	"time"
)

var (
	// dbHealthy tracks if the database connection is healthy
	dbHealthy = true
)

// SetDBConnectionHealth sets the database connection health status
func SetDBConnectionHealth(healthy bool) {
	dbHealthy = healthy
}

// TrackDBConnections tracks database connection health
func TrackDBConnections() {
	// Update connection health every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Track connection health status (1 = healthy, 0 = unhealthy)
			if dbHealthy {
				ActiveConnections.Set(ctx, 1.0)
			} else {
				ActiveConnections.Set(ctx, 0.0)
			}
		}
	}()
}
