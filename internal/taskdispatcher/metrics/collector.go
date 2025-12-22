package metrics

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()
)

// prometheusMetrics is an interface for metrics that support Prometheus export
type prometheusMetrics interface {
	PrometheusHandler() http.Handler
	PrometheusRegistry() *prometheus.Registry
}

// Collector manages metrics collection
type Collector struct {
	handler http.Handler
	metrics observability.Metrics
}

// NewCollector creates a new metrics collector
// If metrics is nil, it falls back to a not-implemented handler
func NewCollector(metrics observability.Metrics) *Collector {
	var handler http.Handler

	if metrics != nil {
		// Try to access PrometheusHandler and Registry via type assertion
		if promMetrics, ok := metrics.(prometheusMetrics); ok {
			// Register our Prometheus metrics with the observability registry
			if registry := promMetrics.PrometheusRegistry(); registry != nil {
				// Use Prometheus handler from observability
				handler = promMetrics.PrometheusHandler()
			} else {
				// Prometheus export is not enabled
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
					_, err := w.Write([]byte("Prometheus export is not enabled"))
					if err != nil {
						fmt.Println("Failed to write response:", err)
					}
				})
			}
		} else {
			// Metrics doesn't support Prometheus export
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
				_, err := w.Write([]byte("Metrics does not support Prometheus export"))
				if err != nil {
					fmt.Println("Failed to write response:", err)
				}
			})
		}
	} else {
		// Fallback for backward compatibility
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			_, err := w.Write([]byte("Metrics not initialized"))
			if err != nil {
				fmt.Println("Failed to write response:", err)
			}
		})
	}

	return &Collector{
		handler: handler,
		metrics: metrics,
	}
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return c.handler
}

// Start starts metrics collection
func (c *Collector) Start() {
	// Update uptime every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if uptimeSeconds != nil {
				uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
			UpdateSystemMetrics()
		}
	}()
}

// UpdateSystemMetrics updates system metrics
func UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memoryUsageBytes != nil {
		memoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}
	if cpuUsagePercent != nil {
		cpuUsagePercent.Set(ctx, float64(memStats.Sys)) // Using Sys memory as a proxy for now
	}
	if goroutinesActive != nil {
		goroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
	}
	if gcDurationSeconds != nil {
		gcDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
	}
}

// CreateRedisMonitoringHooks creates monitoring hooks for the Redis client
func CreateRedisMonitoringHooks() *redis.MonitoringHooks {
	return &redis.MonitoringHooks{
		OnOperationStart: func(operation string, key string) {
			// Optional: Could track operation starts if needed
		},

		OnOperationEnd: func(operation string, key string, duration time.Duration, err error) {
			status := "success"
			if err != nil {
				status = "error"
			}

			if redisOperationsTotal != nil {
				redisOperationsTotal.WithLabelValues(operation, status).Inc(ctx)
			}
			if redisOperationDuration != nil {
				redisOperationDuration.WithLabelValues(operation).Record(ctx, duration.Seconds())
			}
		},

		OnConnectionStatus: func(connected bool, latency time.Duration) {
			healthValue := float64(0)
			if connected {
				healthValue = float64(1)
			}
			if redisConnectionHealth != nil {
				redisConnectionHealth.WithLabelValues("main").Set(ctx, healthValue)
			}

			// Update ping duration if connected
			if connected {
				if pingDuration != nil {
					pingDuration.Record(ctx, latency.Seconds())
				}
				if pingOperationsTotal != nil {
					pingOperationsTotal.WithLabelValues("success").Inc(ctx)
				}
			} else {
				if pingOperationsTotal != nil {
					pingOperationsTotal.WithLabelValues("failure").Inc(ctx)
				}
			}
		},

		OnRecoveryStart: func(reason string) {
			// Mark connection as unhealthy during recovery
			if redisConnectionHealth != nil {
				redisConnectionHealth.WithLabelValues("main").Set(ctx, float64(0))
			}
		},

		OnRecoveryEnd: func(success bool, attempts int, duration time.Duration) {
			status := "failure"
			if success {
				status = "success"
				if redisConnectionHealth != nil {
					redisConnectionHealth.WithLabelValues("main").Set(ctx, float64(1))
				}
			}
			if redisConnectionRecoveries != nil {
				redisConnectionRecoveries.WithLabelValues(status).Inc(ctx)
			}
		},

		OnRetryAttempt: func(operation string, attempt int, err error) {
			if redisRetryAttempts != nil {
				redisRetryAttempts.WithLabelValues(operation).Inc(ctx)
			}
		},
	}
}
