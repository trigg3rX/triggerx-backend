# Observability Package

Unified observability package providing Logs, Traces, and Metrics (LGTM) using OpenTelemetry.

## Overview

This package provides a unified interface for all three pillars of observability:

- **Logs**: Structured logging with trace correlation
- **Traces**: Distributed tracing with span management
- **Metrics**: Counter, Gauge, and Histogram metrics

All telemetry is exported via OTLP HTTP to an OpenTelemetry collector, which can then forward to your observability backend (Grafana, Prometheus, etc.).

## Quick Start

### 1. Initialize Observability

```go
package main

import (
    "context"
    "time"

    "github.com/trigg3rX/triggerx-backend/pkg/observability"
    "go.opentelemetry.io/otel/attribute"
)

func main() {
    // Create configuration
    cfg := observability.Config{
        ServiceName:    observability.ServerService,
        ServiceVersion: "1.0.0",
        InstanceID:     observability.InstanceID,
        OTELExporterEndpoint: "http://localhost:4318",
        DevMode:        false,
        LogLevel:       "info",
        BatchTimeout:   5 * time.Second,
        ExportTimeout:  30 * time.Second,
        MaxExportBatch: 512,
    }

    // Initialize observability
    obs, err := observability.Initialize(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer obs.Shutdown(context.Background())

    // Use the components
    logger := obs.Logger()
    tracer := obs.Tracer()
    metrics := obs.Metrics()

    // Your service code here...
}
```

### 2. Define Service-Specific Metrics

Each service can define its own metrics:

```go
// In your service initialization
var (
    requestCounter   observability.Counter
    latencyHistogram observability.Histogram
    memoryGauge      observability.Gauge
)

func initMetrics(metrics observability.Metrics) {
    requestCounter = metrics.Counter("http_requests_total",
        observability.WithDescription("Total number of HTTP requests"),
        observability.WithUnit("1"),
    )

    latencyHistogram = metrics.Histogram("http_request_duration_seconds",
        observability.WithDescription("HTTP request latency"),
        observability.WithUnit("s"),
    )

    memoryGauge = metrics.Gauge("memory_usage_bytes",
        observability.WithDescription("Memory usage in bytes"),
        observability.WithUnit("By"),
    )
}
```

### 3. Use Logging

```go
func handleRequest(ctx context.Context, logger observability.Logger) {
    logger.Info(ctx, "Processing request",
        observability.String("user_id", "123"),
        observability.Int("request_id", 456),
    )

    logger.Error(ctx, "Failed to process",
        observability.Error(err),
        observability.String("operation", "process_request"),
    )
}
```

### 4. Use Tracing

```go
func processRequest(ctx context.Context, tracer observability.Tracer) error {
    ctx, span := tracer.Start(ctx, "process-request",
        observability.WithSpanKind(trace.SpanKindServer),
        observability.WithAttributes(
            attribute.String("user.id", "123"),
        ),
    )
    defer span.End()

    // Your business logic
    result, err := doWork(ctx)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetAttributes(attribute.String("result.status", "success"))
    span.SetStatus(codes.Ok, "")
    return nil
}
```

### 5. Use Metrics

```go
func handleHTTPRequest(ctx context.Context, metrics observability.Metrics) {
    start := time.Now()

    // Increment request counter
    requestCounter.Inc(ctx,
        attribute.String("method", "GET"),
        attribute.String("endpoint", "/api/users"),
    )

    // Record latency
    latency := time.Since(start).Seconds()
    latencyHistogram.Record(ctx, latency,
        attribute.String("method", "GET"),
        attribute.String("status", "200"),
    )

    // Update memory gauge
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    memoryGauge.Record(ctx, float64(memStats.Alloc),
        attribute.String("type", "heap"),
    )
}
```

### 6. gRPC Integration

```go
import (
    "google.golang.org/grpc"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Server-side interceptor
server := grpc.NewServer(
    grpc.UnaryInterceptor(
        observability.TraceInterceptor(tracer, "my-service"),
    ),
)

// Client-side interceptor
conn, err := grpc.Dial("localhost:50051",
    grpc.WithUnaryInterceptor(
        observability.TraceClientInterceptor(tracer, "my-service"),
    ),
)
```

## Configuration

### Environment Variables

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP collector endpoint (default: `http://localhost:4318`)
- `DEV_MODE`: Enable console logging (default: `false`)

### Config Structure

```go
type Config struct {
    ServiceName          ServiceName  // Service identifier
    ServiceVersion      string       // Version string
    InstanceID          string       // Instance identifier
    OTELExporterEndpoint string      // OTLP collector endpoint
    DevMode             bool         // Enable console logging
    LogLevel            string       // Log level
    BatchTimeout        time.Duration // Batch export timeout
    ExportTimeout       time.Duration // Export timeout
    MaxExportBatch      int          // Max batch size
    SuccessSamplingRate float64      // Success sampling rate (0.0 to 1.0)
    ErrorSamplingRate   float64      // Error sampling rate (0.0 to 1.0)
}
```

## Best Practices

1. **Initialize once**: Create the observability instance at application startup
2. **Define metrics early**: Create metric instruments during initialization
3. **Use context**: Always pass context.Context for trace correlation
4. **Graceful shutdown**: Always call `obs.Shutdown()` on application exit
5. **Service-specific metrics**: Each service should define its own metrics with unique names
6. **Consistent naming**: Use consistent metric naming conventions (e.g., `service_metric_name_unit`)
