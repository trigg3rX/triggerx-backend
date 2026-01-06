package metrics

import (
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
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
	logger  observability.Logger
}

// NewCollector creates a new metrics collector
// If metrics is nil, it falls back to a not-implemented handler
func NewCollector(metrics observability.Metrics, logger observability.Logger) *Collector {
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
					if err != nil && logger != nil {
						logger.Error(r.Context(), "Failed to write response", observability.Error(err))
					}
				})
			}
		} else {
			// Metrics doesn't support Prometheus export
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
				_, err := w.Write([]byte("Metrics does not support Prometheus export"))
				if err != nil && logger != nil {
					logger.Error(r.Context(), "Failed to write response", observability.Error(err))
				}
			})
		}
	} else {
		// Fallback for backward compatibility
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			_, err := w.Write([]byte("Metrics not initialized"))
			if err != nil && logger != nil {
				logger.Error(r.Context(), "Failed to write response", observability.Error(err))
			}
		})
	}

	return &Collector{
		handler: handler,
		metrics: metrics,
		logger:  logger,
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
			if UptimeSeconds != nil {
				UptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
			UpdateSystemMetrics()
		}
	}()
}

// UpdateSystemMetrics updates system metrics (similar to keeper's middleware)
func UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	if MemoryUsageBytes != nil {
		MemoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}
	if CPUUsagePercent != nil {
		CPUUsagePercent.Set(ctx, float64(memStats.Sys))
	}
	if GoroutinesActive != nil {
		GoroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
	}
	if GCDurationSeconds != nil {
		GCDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
	}
}

// CreateRedisMonitoringHooks creates monitoring hooks for the Redis client
func CreateRedisMonitoringHooks() *redisClient.MonitoringHooks {
	return &redisClient.MonitoringHooks{
		OnOperationStart: func(operation string, key string) {
			// Optional: Could track operation starts if needed
		},

		OnOperationEnd: func(operation string, key string, duration time.Duration, err error) {
			status := "success"
			if err != nil {
				status = "error"
			}

			if RedisOperationsTotal != nil {
				RedisOperationsTotal.WithLabelValues(operation, status).Inc(ctx)
			}
			if RedisOperationDuration != nil {
				RedisOperationDuration.WithLabelValues(operation).Record(ctx, duration.Seconds())
			}
		},

		OnConnectionStatus: func(connected bool, latency time.Duration) {
			healthValue := float64(0)
			if connected {
				healthValue = float64(1)
			}
			if RedisConnectionHealth != nil {
				RedisConnectionHealth.WithLabelValues("main").Set(ctx, healthValue)
			}

			// Update ping duration if connected
			if connected {
				if PingDuration != nil {
					PingDuration.Record(ctx, latency.Seconds())
				}
				if PingOperationsTotal != nil {
					PingOperationsTotal.WithLabelValues("success").Inc(ctx)
				}
			} else {
				if PingOperationsTotal != nil {
					PingOperationsTotal.WithLabelValues("failure").Inc(ctx)
				}
			}
		},

		OnRecoveryStart: func(reason string) {
			// Mark connection as unhealthy during recovery
			if RedisConnectionHealth != nil {
				RedisConnectionHealth.WithLabelValues("main").Set(ctx, float64(0))
			}
		},

		OnRecoveryEnd: func(success bool, attempts int, duration time.Duration) {
			status := "failure"
			if success {
				status = "success"
				if RedisConnectionHealth != nil {
					RedisConnectionHealth.WithLabelValues("main").Set(ctx, float64(1))
				}
			}
			if RedisConnectionRecoveries != nil {
				RedisConnectionRecoveries.WithLabelValues(status).Inc(ctx)
			}
		},

		OnRetryAttempt: func(operation string, attempt int, err error) {
			if RedisRetryAttempts != nil {
				RedisRetryAttempts.WithLabelValues(operation).Inc(ctx)
			}
		},
	}
}

// UpdateRedisClientMetrics updates metrics from Redis client operation metrics
func UpdateRedisClientMetrics(operationMetrics map[string]*redisClient.OperationMetrics) {
	for operation, metrics := range operationMetrics {
		// Update operation counters
		if RedisOperationsTotal != nil {
			RedisOperationsTotal.WithLabelValues(operation, "success").Add(ctx, float64(metrics.SuccessCount))
			RedisOperationsTotal.WithLabelValues(operation, "error").Add(ctx, float64(metrics.ErrorCount))
		}

		// Update retry attempts
		if RedisRetryAttempts != nil {
			RedisRetryAttempts.WithLabelValues(operation).Add(ctx, float64(metrics.RetryCount))
		}

		// Update average latency (as a gauge for monitoring)
		if metrics.AverageLatency > 0 {
			if RedisOperationDuration != nil {
				RedisOperationDuration.WithLabelValues(operation).Record(ctx, metrics.AverageLatency.Seconds())
			}
		}
	}
}
