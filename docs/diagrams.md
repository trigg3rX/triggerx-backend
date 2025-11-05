# Architecture Diagrams

Visual representations of the metrics package architecture.

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Your Services Layer                          │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ Keeper   │  │  Health  │  │ DBServer │  │Dispatcher│  ...      │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘           │
│       │             │             │             │                    │
│       └─────────────┴─────────────┴─────────────┘                   │
│                            │                                         │
└────────────────────────────┼─────────────────────────────────────────┘
                             │
                    Uses pkg/metrics
                             │
┌────────────────────────────▼─────────────────────────────────────────┐
│                      pkg/metrics Package                             │
│                                                                       │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    Collector (Core)                            │ │
│  │  • Creates isolated Prometheus registry per service           │ │
│  │  • Manages metric lifecycle (Start/Stop)                       │ │
│  │  • Exposes HTTP handler for /metrics endpoint                  │ │
│  │  • Registers custom service metrics                            │ │
│  └───┬─────────────┬───────────────┬──────────────┬───────────────┘ │
│      │             │               │              │                  │
│  ┌───▼────────┐ ┌──▼──────────┐ ┌─▼─────────┐ ┌──▼──────────────┐ │
│  │  Common    │ │ HTTP        │ │  Helpers  │ │  Options        │ │
│  │  Metrics   │ │ Metrics     │ │           │ │                  │ │
│  ├────────────┤ ├─────────────┤ ├───────────┤ ├──────────────────┤ │
│  │• Uptime    │ │• Requests   │ │• Builders │ │• Namespace       │ │
│  │• Memory    │ │• Duration   │ │• Fluent   │ │• Intervals       │ │
│  │• CPU       │ │• Status     │ │  API      │ │• Feature Flags   │ │
│  │• Goroutines│ │• RPS        │ │           │ │                  │ │
│  │• GC        │ │             │ │           │ │                  │ │
│  └────────────┘ └─────────────┘ └───────────┘ └──────────────────┘ │
│                                                                       │
└───────────────────────────────┬───────────────────────────────────────┘
                                │
                    Exposes /metrics endpoint
                                │
┌───────────────────────────────▼───────────────────────────────────────┐
│                      Prometheus Server                                │
│  • Scrapes /metrics endpoints from all services                       │
│  • Stores time-series data                                            │
│  • Provides query interface (PromQL)                                  │
└───────────────────────────────┬───────────────────────────────────────┘
                                │
                       Visualized by
                                │
┌───────────────────────────────▼───────────────────────────────────────┐
│                         Grafana Dashboard                             │
│  • Real-time metrics visualization                                    │
│  • Custom dashboards per service                                      │
│  • Alerting based on thresholds                                       │
└───────────────────────────────────────────────────────────────────────┘
```

## Service Integration Flow

```
Service Startup
      │
      ├─→ 1. Create Collector
      │      collector := metrics.NewCollector("keeper")
      │      • Creates isolated Prometheus registry
      │      • Applies configuration options
      │      • Initializes common metrics (if enabled)
      │
      ├─→ 2. Start Collector
      │      collector.Start()
      │      • Spawns goroutine: Update uptime every 15s
      │      • Spawns goroutine: Update system metrics every 30s
      │      • Metrics begin collecting
      │
      ├─→ 3. Create Custom Metrics
      │      builder := metrics.NewMetricBuilder(collector, "keeper")
      │      tasksTotal := builder.Counter("tasks_total", "help text")
      │      • Metrics are automatically registered with collector
      │      • Available for updates immediately
      │
      ├─→ 4. Setup HTTP Endpoint
      │      router.GET("/metrics", metrics.MetricsHandler(collector))
      │      • Exposes all metrics via HTTP
      │      • Prometheus-compatible format
      │
      ├─→ 5. Use Metrics in Business Logic
      │      tasksTotal.Inc()
      │      taskDuration.Observe(1.5)
      │      • Thread-safe updates
      │      • No locks needed
      │
      └─→ 6. Graceful Shutdown
           defer collector.Stop()
           • Stops background goroutines
           • Waits for cleanup (WaitGroup)
           • Clean shutdown
```

## Data Flow: Metric Update

```
Application Code                 Prometheus Client           Registry
      │                                │                         │
      │  metric.Inc()                  │                         │
      ├───────────────────────────────→│                         │
      │                                │                         │
      │                                │  Atomic update          │
      │                                ├────────────────────────→│
      │                                │                         │
      │  metric.Observe(1.5)           │                         │
      ├───────────────────────────────→│                         │
      │                                │                         │
      │                                │  Update histogram       │
      │                                ├────────────────────────→│
      │                                │                         │
      ↓                                ↓                         ↓
   Continue                        Thread-safe              Metrics ready
   execution                      no blocking              for scraping
```

## HTTP Request Flow

```
Prometheus Server                  /metrics Endpoint         Collector
      │                                   │                       │
      │  GET /metrics                     │                       │
      ├──────────────────────────────────→│                       │
      │                                   │                       │
      │                                   │  Get all metrics      │
      │                                   ├──────────────────────→│
      │                                   │                       │
      │                                   │  Registry snapshot    │
      │                                   │←──────────────────────┤
      │                                   │                       │
      │                                   │  Format as Prometheus │
      │                                   │  text format          │
      │                                   │                       │
      │  200 OK + metrics text            │                       │
      │←──────────────────────────────────┤                       │
      │                                   │                       │
      │  # TYPE triggerx_keeper_tasks... │                       │
      │  triggerx_keeper_tasks_total 42  │                       │
      │  ...                              │                       │
      ↓                                   ↓                       ↓
   Store in                           Handler                Continue
   time-series DB                     complete              collecting
```

## Metric Types & Use Cases

```
┌────────────────────────────────────────────────────────────────────┐
│                         Metric Types                               │
├────────────────┬───────────────────────────────────────────────────┤
│  Counter       │  Monotonically increasing value                   │
│  ============  │  ===============================================  │
│                │  • requests_total                                 │
│    📈          │  • errors_total                                   │
│   ▗▞▀▀▀▀▀      │  • bytes_sent_total                              │
│  ▞▘            │                                                   │
│                │  Operations: Inc(), Add(n)                        │
├────────────────┼───────────────────────────────────────────────────┤
│  Gauge         │  Value that can go up or down                     │
│  ============  │  ===============================================  │
│                │  • active_connections                             │
│    📊          │  • queue_size                                     │
│   ▗▀▖▗▄▖▗▀▖    │  • memory_usage_bytes                            │
│  ▞  ▘  ▘  ▘   │                                                   │
│                │  Operations: Set(), Inc(), Dec()                  │
├────────────────┼───────────────────────────────────────────────────┤
│  Histogram     │  Distribution of observed values                  │
│  ============  │  ===============================================  │
│                │  • request_duration_seconds                       │
│    📊📊📊      │  • response_size_bytes                           │
│    █           │  • query_duration                                 │
│   ██           │                                                   │
│  ███  █        │  Operations: Observe(value)                       │
│  ████████      │  Provides: count, sum, buckets, quantiles         │
└────────────────┴───────────────────────────────────────────────────┘
```

## Collector Lifecycle

```
┌───────────────────────────────────────────────────────────────────┐
│                    Collector Lifecycle                            │
└───────────────────────────────────────────────────────────────────┘

NewCollector()
     │
     ├─→ Create Registry
     ├─→ Apply Options
     └─→ Init Common Metrics (if enabled)
           │
           ↓
        [CREATED]
           │
           │ collector.Start()
           ↓
     ┌───────────┐
     │  RUNNING  │ ←──┐
     └───────────┘    │
           │          │
           ├─→ Goroutine: Update Uptime (15s) ────┐
           ├─→ Goroutine: Update System (30s) ────┤
           │                                       │
           │ collector.MustRegister(...)           │
           ├─→ Register Custom Metrics             │
           │                                       │
           │ HTTP GET /metrics                     │
           ├─→ Serve Metrics                       │
           │                                       │
           │                                       │
           │ collector.Stop()                      │
           ↓                                       │
     ┌───────────┐                                │
     │ STOPPING  │                                │
     └───────────┘                                │
           │                                       │
           ├─→ Close stopCh ──────────────────────┘
           ├─→ Wait for goroutines (WaitGroup)
           └─→ Cleanup
                 │
                 ↓
             [STOPPED]
```

## Multi-Service Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Platform                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Service A (Keeper)          Service B (Health)     Service C (...)  │
│  ┌──────────────────┐        ┌──────────────────┐  ┌──────────────┐│
│  │ Collector        │        │ Collector        │  │ Collector    ││
│  │ namespace:keeper │        │ namespace:health │  │ ...          ││
│  │                  │        │                  │  │              ││
│  │ Registry A       │        │ Registry B       │  │ Registry C   ││
│  │ ┌──────────────┐ │        │ ┌──────────────┐ │  │ ...          ││
│  │ │Common Metrics│ │        │ │Common Metrics│ │  │              ││
│  │ │Keeper Metrics│ │        │ │Health Metrics│ │  │              ││
│  │ └──────────────┘ │        │ └──────────────┘ │  │              ││
│  │                  │        │                  │  │              ││
│  │ :8080/metrics    │        │ :8081/metrics    │  │ ...          ││
│  └──────────────────┘        └──────────────────┘  └──────────────┘│
│           │                           │                      │      │
└───────────┼───────────────────────────┼──────────────────────┼──────┘
            │                           │                      │
            └───────────────┬───────────┴──────────────────────┘
                            │
                   Scraped by Prometheus
                            │
                            ↓
                 ┌────────────────────────┐
                 │   Prometheus Server    │
                 │                        │
                 │  All metrics combined  │
                 │  with service labels   │
                 └────────────────────────┘
```

## Component Interaction

```
┌──────────────────────────────────────────────────────────────────┐
│  Application Layer                                               │
│                                                                   │
│    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│    │  HTTP Router │    │ Worker Pool  │    │   DB Client  │   │
│    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘   │
│           │                   │                    │            │
└───────────┼───────────────────┼────────────────────┼────────────┘
            │                   │                    │
            │ Update metrics    │                    │
            ↓                   ↓                    ↓
┌──────────────────────────────────────────────────────────────────┐
│  Metrics Layer (pkg/metrics)                                     │
│                                                                   │
│    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│    │HTTPMetrics   │    │CustomMetrics │    │CommonMetrics │   │
│    │.Inc()        │    │.Observe()    │    │ Auto-update  │   │
│    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘   │
│           │                   │                    │            │
│           └───────────────────┼────────────────────┘            │
│                               │                                  │
│                         ┌─────▼─────┐                           │
│                         │ Collector │                           │
│                         │           │                           │
│                         │ Registry  │                           │
│                         └─────┬─────┘                           │
└───────────────────────────────┼──────────────────────────────────┘
                                │
                                │ Expose via HTTP
                                ↓
                        ┌───────────────┐
                        │   /metrics    │
                        │   Endpoint    │
                        └───────────────┘
```

## Summary

This architecture provides:

✅ **Isolation**: Each service has its own registry  
✅ **Consistency**: All services use same patterns  
✅ **Flexibility**: Easy to add custom metrics  
✅ **Performance**: Minimal overhead, thread-safe  
✅ **Maintainability**: Common code in one place  
✅ **Scalability**: Works with any number of services  
✅ **Observability**: Full Prometheus integration  

# Metrics Package Architecture

## Overview

The `/pkg/metrics` package provides a centralized, reusable metrics collection system for all TriggerX services using Prometheus.

## Design Goals

1. **DRY (Don't Repeat Yourself)**: Common metrics defined once
2. **Flexibility**: Services can define custom metrics easily
3. **Consistency**: All services follow the same patterns
4. **Type Safety**: Strong typing with Go interfaces
5. **Performance**: Minimal overhead with efficient collection
6. **Testability**: Easy to mock and test

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Service Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Keeper     │  │   Health     │  │  DBServer    │  ...    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼──────────────────┼──────────────────┼────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                   pkg/metrics Package                           │
│                            │                                     │
│  ┌─────────────────────────▼──────────────────────────┐        │
│  │               Collector (Core)                      │        │
│  │  - Registry Management                              │        │
│  │  - Lifecycle (Start/Stop)                          │        │
│  │  - HTTP Handler                                     │        │
│  └──────┬───────────────┬──────────────┬──────────────┘        │
│         │               │              │                        │
│  ┌──────▼───────┐ ┌────▼──────┐ ┌────▼──────────┐            │
│  │CommonMetrics│ │HTTPMetrics│ │MetricBuilder  │            │
│  │- Uptime     │ │- Requests │ │- Helpers      │            │
│  │- CPU        │ │- Duration │ │- Fluent API   │            │
│  │- Memory     │ │- RPS      │ │               │            │
│  │- GC         │ └───────────┘ └───────────────┘            │
│  │- Goroutines │                                              │
│  └─────────────┘                                              │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │         Prometheus Registry                              │ │
│  │  - Metric Registration                                   │ │
│  │  - Metric Storage                                        │ │
│  │  - HTTP Exposition                                       │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                             │
                             │ HTTP /metrics
                             ▼
                      ┌──────────────┐
                      │  Prometheus  │
                      │   Server     │
                      └──────────────┘
```

## Component Details

### 1. Collector (Core)

The central component that manages metrics collection for a service.

**Responsibilities:**
- Creates and manages a Prometheus registry
- Initializes common metrics
- Starts/stops background collection goroutines
- Provides HTTP handler for metrics exposition
- Registers custom service metrics

**Key Methods:**
```go
NewCollector(serviceName string, opts ...Option) *Collector
Start()
Stop()
Handler() http.Handler
Registry() *prometheus.Registry
MustRegister(collectors ...prometheus.Collector)
Common() *CommonMetrics
```

### 2. CommonMetrics

Provides system-level metrics common to all services.

**Metrics Provided:**
- `uptime_seconds` - Service uptime
- `memory_usage_bytes` - Memory consumption
- `cpu_usage_percent` - CPU utilization
- `goroutines_active` - Active goroutines
- `gc_duration_seconds` - Garbage collection time

**Collection:**
- Automatic background collection
- Configurable update intervals
- Thread-safe updates

### 3. HTTPMetrics

HTTP-specific metrics for services with HTTP APIs.

**Metrics Provided:**
- `http_requests_total` - Total requests (by method, endpoint, status)
- `http_request_duration_seconds` - Request latency histogram
- `requests_per_second` - Request throughput gauge

**Integration:**
- Gin middleware
- Standard http.Handler support
- Automatic request tracking

### 4. MetricBuilder

Fluent API for creating and registering metrics.

**Benefits:**
- Reduces boilerplate
- Automatic registration
- Consistent metric naming
- Type-safe construction

**Example:**
```go
builder := metrics.NewMetricBuilder(collector, "keeper")
counter := builder.Counter("tasks_total", "Total tasks processed")
```

### 5. Options & Configuration

Functional options pattern for flexible configuration.

**Available Options:**
- `WithNamespace(string)` - Custom namespace
- `WithCommonMetrics(bool)` - Enable/disable common metrics
- `WithUptimeInterval(duration)` - Uptime update frequency
- `WithSystemMetricsInterval(duration)` - System metrics update frequency

## Data Flow

### Initialization Flow

```
Service Startup
      ↓
Create Collector ──→ Create Registry
      ↓                    ↓
Apply Options        Initialize CommonMetrics
      ↓                    ↓
Start() ─────────→ Start Background Goroutines
      ↓                    ↓
Create Custom      Register with Registry
  Metrics               ↓
      ↓              Setup HTTP Handler
Register Custom          ↓
  Metrics          Ready to Serve /metrics
      ↓
Service Running
```

### Metric Update Flow

```
Application Code
      ↓
Call Metric Method (Inc(), Set(), Observe())
      ↓
Prometheus Client Library
      ↓
Thread-Safe Update in Registry
      ↓
Metric Ready for Scraping
```

### HTTP Request Flow

```
Prometheus Scraper
      ↓
GET /metrics
      ↓
HTTP Handler (from Collector)
      ↓
Prometheus HTTP Exporter
      ↓
Read all metrics from Registry
      ↓
Format as Prometheus Text
      ↓
Return HTTP Response
```

## Service Integration Pattern

### Pattern 1: Simple Service (No HTTP)

```go
func main() {
    // Create collector
    collector := metrics.NewCollector("worker")
    collector.Start()
    defer collector.Stop()
    
    // Create custom metrics
    tasksProcessed := prometheus.NewCounter(...)
    collector.MustRegister(tasksProcessed)
    
    // Expose metrics on separate port
    go func() {
        http.Handle("/metrics", collector.Handler())
        http.ListenAndServe(":9090", nil)
    }()
    
    // Run service
    runWorker(tasksProcessed)
}
```

### Pattern 2: HTTP Service (with Gin)

```go
func main() {
    // Create collector
    collector := metrics.NewCollector("api")
    collector.Start()
    defer collector.Stop()
    
    // Create HTTP metrics
    httpMetrics := metrics.NewHTTPMetrics(collector, "api")
    
    // Create custom metrics
    apiMetrics := NewAPIMetrics(collector)
    
    // Setup router
    router := gin.Default()
    router.Use(httpMetrics.GinMiddleware())
    router.GET("/metrics", metrics.MetricsHandler(collector))
    
    // Run service
    router.Run(":8080")
}
```

### Pattern 3: Complex Service (Multiple Components)

```go
type Service struct {
    collector     *metrics.Collector
    apiMetrics    *APIMetrics
    workerMetrics *WorkerMetrics
    dbMetrics     *DBMetrics
}

func NewService() *Service {
    collector := metrics.NewCollector("complex_service")
    collector.Start()
    
    return &Service{
        collector:     collector,
        apiMetrics:    NewAPIMetrics(collector),
        workerMetrics: NewWorkerMetrics(collector),
        dbMetrics:     NewDBMetrics(collector),
    }
}

func (s *Service) Start() {
    // Components use their respective metrics
    go s.runAPI(s.apiMetrics)
    go s.runWorker(s.workerMetrics)
    go s.runDBMaintenance(s.dbMetrics)
}

func (s *Service) Stop() {
    s.collector.Stop()
}
```

## Thread Safety

### Concurrent Access

All metrics are thread-safe by default:
- Prometheus client library provides internal locking
- Multiple goroutines can update metrics concurrently
- No additional synchronization needed in application code

### Background Goroutines

The collector manages background goroutines:
- Uptime updater (default: 15s interval)
- System metrics updater (default: 30s interval)
- Proper cleanup on Stop()
- WaitGroup ensures graceful shutdown

## Performance Considerations

### Memory

- Each metric consumes minimal memory (~100-200 bytes base)
- Label combinations create new metric instances
- Registry maintains references to all metrics
- Common metrics shared across services

### CPU

- Metric updates: ~10-50 nanoseconds (extremely fast)
- Background collection: ~1-5ms every 15-30 seconds
- HTTP scraping: ~1-10ms depending on metric count
- Negligible impact on service performance

### Cardinality

**Important:** Avoid high cardinality labels!

❌ **Bad** (unbounded):
```go
counter.WithLabelValues(userID)  // Millions of users = millions of metrics
```

✅ **Good** (bounded):
```go
counter.WithLabelValues(userType)  // 3-5 user types = 3-5 metrics
```

## Extensibility

### Adding New Common Metrics

To add a new common metric for all services:

1. Add to `CommonMetrics` struct in `common.go`
2. Initialize in `newCommonMetrics()`
3. Register with registry
4. Add update logic if needed

### Creating New Metric Types

To create new helper types (e.g., RedisMetrics):

1. Create new file `redis.go`
2. Define struct with metrics
3. Add constructor: `NewRedisMetrics(collector *Collector)`
4. Register metrics with collector
5. Return initialized struct

### Custom Collectors

For advanced use cases, implement `prometheus.Collector` interface:

```go
type CustomCollector struct {}

func (c *CustomCollector) Describe(ch chan<- *prometheus.Desc) {}
func (c *CustomCollector) Collect(ch chan<- prometheus.Metric) {}

collector.MustRegister(customCollector)
```

## Testing Strategy

### Unit Tests

```go
func TestMetrics(t *testing.T) {
    collector := metrics.NewCollector("test_service")
    defer collector.Stop()
    
    counter := prometheus.NewCounter(...)
    collector.MustRegister(counter)
    
    counter.Inc()
    
    // Verify metric value
    // Note: In real tests, use testutil.CollectAndCompare
}
```

### Integration Tests

```go
func TestMetricsEndpoint(t *testing.T) {
    collector := metrics.NewCollector("test")
    collector.Start()
    defer collector.Stop()
    
    server := httptest.NewServer(collector.Handler())
    defer server.Close()
    
    resp, _ := http.Get(server.URL)
    body, _ := io.ReadAll(resp.Body)
    
    assert.Contains(t, string(body), "triggerx_test_uptime_seconds")
}
```

## Future Enhancements

### Potential Additions

1. **Tracing Integration**: OpenTelemetry support
2. **Dynamic Labels**: Runtime label configuration
3. **Metric Aggregation**: Cross-service aggregation
4. **Alert Rules**: Built-in alerting configuration
5. **Dashboard Templates**: Grafana dashboard generator
6. **Metric Validation**: Schema validation for metrics
7. **Cost Metrics**: Cloud cost tracking
8. **Business Metrics**: Custom business KPI tracking

### Backwards Compatibility

When adding new features:
- Use new optional parameters
- Maintain existing function signatures
- Provide migration guides
- Version the package if needed

## Best Practices

1. **One Collector Per Service**: Don't create multiple collectors
2. **Start Early**: Initialize collector during service startup
3. **Clean Shutdown**: Always defer collector.Stop()
4. **Meaningful Names**: Use clear, descriptive metric names
5. **Consistent Labels**: Use same label names across metrics
6. **Document Metrics**: Add help text to all metrics
7. **Test Metrics**: Include metrics in integration tests
8. **Monitor Cardinality**: Watch for label explosion
9. **Use Helpers**: Leverage MetricBuilder for consistency
10. **Version Metrics**: Plan for metric evolution

## Conclusion

The centralized metrics package provides a robust, efficient, and extensible foundation for observability across all TriggerX services. By standardizing metric collection and exposition, it reduces code duplication while improving maintainability and consistency.

# X-Trace-ID Implementation

## Overview
All API requests now include an `X-Trace-ID` header for distributed tracing and debugging.

## Trace ID Format
- **Frontend**: `tgrx-frnt-<uuid>` (e.g., `tgrx-frnt-550e8400-e29b-41d4-a716-446655440000`)
- **SDK**: `tgrx-sdk-<uuid>` (to be implemented in SDK)

## How It Works

### 1. Trace ID Generation
The utility function `generateTraceId()` in `/lib/traceId.ts` generates UUIDs with the appropriate prefix:

```typescript
export function generateTraceId(): string {
  const uuid = crypto.randomUUID();
  return `tgrx-frnt-${uuid}`;
}
```

### 2. API Proxy Routes
All Next.js API proxy routes automatically:
- Check if the client provided an `X-Trace-ID` header
- Generate a new trace ID if not provided
- Forward the trace ID to the backend
- Include the trace ID in logs for easy debugging

Example from `/api/jobs/route.ts`:
```typescript
const traceId = getOrGenerateTraceId(req.headers.get("X-Trace-ID"));
devLog(`[API Route] [${traceId}] Proxying POST request to /api/jobs`);

const response = await fetch(`${API_BASE_URL}/api/jobs`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-Api-Key": API_KEY,
    "X-Trace-ID": traceId,  // <-- Forwarded to backend
  },
  body: JSON.stringify(body),
});
```

### 3. Backend Validation
The backend's `TraceMiddleware` enforces that all requests (except health/metrics endpoints) include an `X-Trace-ID` header. If missing, it returns:

```json
{
  "error": "Missing trace ID",
  "code": "MISSING_TRACE_ID",
  "message": "All requests must include X-Trace-ID header"
}
```

## Debugging with Trace IDs

### In Frontend Logs
Look for log entries with trace IDs:
```
[API Route] [tgrx-frnt-550e8400-e29b-41d4-a716-446655440000] Proxying POST request to /api/jobs
```

### In Backend Logs
The backend's traced logger will include the trace ID in all related log entries, making it easy to track a request through the entire system.

### In Grafana Tempo
The OpenTelemetry integration allows you to search for traces by trace ID in Grafana Tempo, providing end-to-end visibility of requests across all microservices.

## Benefits

1. **Easy Debugging**: Track a single request across frontend → Next.js API → backend → microservices
2. **Error Tracing**: When an error occurs, the trace ID helps identify all related logs
3. **Performance Monitoring**: Measure request latency at each layer
4. **Distributed Tracing**: Full OpenTelemetry integration with Grafana Tempo
5. **Log Correlation**: All logs for a request share the same trace ID

## Future: SDK Implementation

When implementing the SDK, use the same pattern with `tgrx-sdk-` prefix:

```typescript
// SDK example
const traceId = `tgrx-sdk-${crypto.randomUUID()}`;
const response = await fetch(url, {
  headers: {
    'X-Trace-ID': traceId,
    // ... other headers
  }
});
```

# Wei-Based Calculations Implementation

## Overview
All monetary calculations in the TriggerX platform now use Wei to avoid floating-point precision errors. This document describes the implementation and conversion utilities.

## Why Wei?

Using Wei (the smallest unit) for all calculations provides:
- **No Precision Loss**: Avoids floating-point arithmetic errors
- **Consistency**: Use it everywhere; for calculations and storage
- **Ethereum Compatibility**: Natural fit with smart contracts
- **Deterministic Results**: Calculations are exact and reproducible

## Units and Conversions

### TriggerGas (TG) Unit System
```
1 ETH  = 1e18 Wei
1 ETH  = 1000 TG
1 TG   = 1e15 Wei
0.001 TG = 1e12 Wei (static job fee per execution)
```

## Where is it used?

- Resource utilized by a task is stored in Wei, named TaskOpxCost
- This includes the tx cost too
- Each job would have a JobCostActual field, which is the total cost of the job (all tasks in it) in Wei
- Keeper points are the TaskOpxCost keeper has performed, or resources consumed to verify a task
- User points are the sum of JobCostActual for all jobs the user has created
