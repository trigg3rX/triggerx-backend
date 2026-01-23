# Observability Infrastructure

TriggerX uses a Prometheus-based observability stack for metrics, with OpenTelemetry for distributed tracing.

## Stack Components

| Component | Purpose | Port |
|-----------|---------|------|
| [Prometheus](https://prometheus.io/) | Metrics collection & storage | 9090 |
| [Grafana](https://grafana.com/) | Visualization & dashboards | 3000 |
| [Tempo](https://grafana.com/oss/tempo/) | Distributed tracing | 3200 |
| [Loki](https://grafana.com/oss/loki/) | Log aggregation | 3100 |
| [Mimir](https://grafana.com/oss/mimir/) | Long-term metrics storage | 9009 |
| [OTel Collector](https://opentelemetry.io/docs/collector/) | Telemetry pipeline | 4317/4318 |

## Metrics Package

The `pkg/observability/` package provides centralized metrics collection:

```go
import "github.com/trigg3rX/triggerx-backend/pkg/observability"

// Create collector per service
collector := metrics.NewCollector("keeper")
collector.Start()
defer collector.Stop()

// Create custom metrics
builder := metrics.NewMetricBuilder(collector, "keeper")
tasksTotal := builder.Counter("tasks_total", "Total tasks processed")

// Update metrics
tasksTotal.Inc()

// Expose endpoint
router.GET("/metrics", metrics.MetricsHandler(collector))
```

### Common Metrics (Auto-collected)

- `uptime_seconds` - Service uptime
- `memory_usage_bytes` - Memory consumption
- `cpu_usage_percent` - CPU utilization
- `goroutines_active` - Active goroutines
- `gc_duration_seconds` - Garbage collection time

### HTTP Metrics

- `http_requests_total` - Total requests by method, endpoint, status
- `http_request_duration_seconds` - Request latency histogram

## Distributed Tracing

OpenTelemetry tracing configured in `pkg/observability/tracer.go`:

```go
// Initialize tracer
tp, err := observability.InitTracer(ctx, "service-name", otelEndpoint)

// Create spans
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

## Configuration

```yaml
# config/otel-collector.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
  otlp/tempo:
    endpoint: tempo:4317
```

### Environment Variables

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=service-name
```

## Grafana Dashboards

Pre-configured dashboards in `config/grafana/provisioning/dashboards/`:

- Task lifecycle monitoring
- Service health metrics
- Resource utilization

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
