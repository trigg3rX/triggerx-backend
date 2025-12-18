# Observability Architecture Plan

This document outlines the architectural decisions for observability (logging, tracing, metrics) initialization and usage across the triggerx-backend codebase.

## Service-Level Initialization

```bash
┌────────────────────────────────────────────────────────────┐
│                    Service (cmd/*/main.go)                 │
│                                                            │
│  1. Initialize observability (logger, tracer, metrics)     │
    (in cmd/*/main.go, while reading env in internal/config) │
│  2. Pass to internal/ services                             │
│  3. Pass to pkg/ packages via constructors                 │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│              pkg/ Packages (websocket, redis, etc.)        │
│                                                            │
│  - Accept observability.Logger as constructor parameter    │
│  - Accept observability.Tracer (if needed)                 │
│  - Accept observability.Metrics (if needed)                │
│  - Use interfaces, not concrete types                      │
└────────────────────────────────────────────────────────────┘
```

### Benefits

1. **Single Source of Truth**

   - One logger/tracer/metrics instance per service
   - Consistent configuration across all components
   - Easier to manage and debug

2. **Centralized Configuration**

   - Configure observability once per service (dev/prod, log levels, exporters)
   - Environment-specific settings in one place
   - No scattered configuration across packages

3. **Better Correlation**

   - Shared trace context across all `pkg/` packages
   - End-to-end tracing from service entry to package calls
   - Unified request/operation tracking

4. **Easier Testing**

   - Inject mocks/no-ops at service boundaries
   - Test packages in isolation with mock observability
   - No need to mock internal observability initialization

5. **Cleaner Dependencies**

   - `pkg/` packages don't depend on observability internals
   - Packages accept interfaces, not concrete implementations
   - Better separation of concerns

6. **Resource Efficiency**

   - Single exporter/processor per service
   - Reduced memory and CPU usage
   - Better resource management

7. **Consistent Patterns**
   - Aligns with dependency injection best practices
   - Follows Go interface-based design
   - Matches existing codebase patterns

## Implementation Pattern

### Service Initialization (cmd/\*/main.go)

```go
package main

import (
    "context"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
    "github.com/trigg3rX/triggerx-backend/pkg/websocket"
    "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
)

func main() {
    // 1. Initialize configuration
    cfg := config.Init()

    // 2. Create observability configuration (using helper function)
    obsConfig := observability.NewConfig(
        observability.KeeperService,
        config.GetVersion(),
        config.GetOTELExporterEndpoint(),
        config.IsDevMode(),
    // Add custom options:
    //     observability.WithBatchTimeout(10*time.Second),
    //     observability.WithLogLevel("debug"),
    )

    // 3. Initialize observability (all three pillars)
    obs, err := observability.Initialize(obsConfig)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize observability: %v", err))
    }
    defer obs.Shutdown(context.Background())

    // 4. Extract individual components
    logger := obs.Logger()
    tracer := obs.Tracer()
    metrics := obs.Metrics()

    // 5. Pass to pkg/ packages via constructors
    wsClient, err := websocket.NewWebSocketClient(
        wsURL,
        websocketConfig,
        logger, // Pass logger
        tracer, // Pass tracer (if needed)
        metrics, // Pass metrics (if needed)
    )

    redisClient, err := redis.NewRedisClient(
        logger, // Pass logger
        tracer, // Pass tracer (if needed)
        metrics, // Pass metrics (if needed)
        redisConfig,
    )

    // 6. Use in services
    service := NewService(logger, tracer, metrics, wsClient, redisClient)

    // ... rest of service initialization
}
```

### Package Implementation (pkg/websocket/websocket.go)

```go
package websocket

import (
    "context"
    "time"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// WebSocketClient accepts observability interfaces
type WebSocketClient struct {
    url     string
    config  *WebSocketRetryConfig
    logger  observability.Logger // Interface, not concrete type
    tracer  observability.Tracer // Optional: for tracing
    metrics observability.Metrics // Optional: for metrics
    // Metric instruments (created once, used multiple times)
    connectionCounter observability.Counter
    connectionGauge   observability.Gauge
    connectionLatency observability.Histogram
    // Track active connections for gauge
    activeConnections int64
}

// Constructor accepts observability interfaces as parameters
func NewWebSocketClient(
    url string,
    config *WebSocketRetryConfig,
    logger observability.Logger,
    tracer observability.Tracer, // Optional
    metrics observability.Metrics, // Optional
) (*WebSocketClient, error) {
    if logger == nil {
        return nil, fmt.Errorf("logger is required")
    }

    client := &WebSocketClient{
        url:    url,
        config: config,
        logger: logger,
        tracer: tracer,
        metrics: metrics,
    }

    // Initialize metric instruments if metrics is provided
    if metrics != nil {
        client.connectionCounter = metrics.Counter(
            "websocket_connections_total",
            observability.WithDescription("Total number of WebSocket connections"),
            observability.WithUnit("1"),
        )
        client.connectionGauge = metrics.Gauge(
            "websocket_connections_active",
            observability.WithDescription("Number of active WebSocket connections"),
            observability.WithUnit("1"),
        )
        client.connectionLatency = metrics.Histogram(
            "websocket_connection_duration_ms",
            observability.WithDescription("WebSocket connection duration in milliseconds"),
            observability.WithUnit("ms"),
        )
    }

    return client, nil
}

// Use logger, tracer, and metrics in methods
func (c *WebSocketClient) Connect(ctx context.Context) error {
    // Start a span for this operation
    var span observability.Span
    if c.tracer != nil {
        ctx, span = c.tracer.Start(ctx, "websocket.connect",
            observability.WithSpanKind(trace.SpanKindClient),
            observability.WithAttributes(
                attribute.String("websocket.url", c.url),
            ),
        )
        defer span.End()
    }

    // Log the operation
    c.logger.Info(ctx, "Connecting to WebSocket",
        observability.String("url", c.url),
    )

    // Record metrics
    startTime := time.Now()
    if c.metrics != nil {
        c.connectionCounter.Inc(ctx, attribute.String("status", "attempting"))
        c.activeConnections++
        c.connectionGauge.Set(ctx, float64(c.activeConnections))
        defer func() {
            duration := time.Since(startTime)
            c.connectionLatency.Record(ctx, float64(duration.Milliseconds()))
        }()
    }

    // ... connection logic
    err := c.doConnect(ctx)
    
    // Handle errors with span and metrics
    if err != nil {
        if span != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            span.SetAttributes(attribute.String("error.type", err.Error()))
        }
        if c.metrics != nil {
            c.connectionCounter.Inc(ctx, attribute.String("status", "error"))
            c.activeConnections--
            c.connectionGauge.Set(ctx, float64(c.activeConnections))
        }
        c.logger.Error(ctx, "Failed to connect to WebSocket",
            observability.String("url", c.url),
            observability.Error(err),
        )
        return err
    }

    // Success case
    if span != nil {
        span.SetStatus(codes.Ok, "")
        span.SetAttributes(attribute.Bool("websocket.connected", true))
    }
    if c.metrics != nil {
        c.connectionCounter.Inc(ctx, attribute.String("status", "success"))
    }
    c.logger.Info(ctx, "Successfully connected to WebSocket",
        observability.String("url", c.url),
    )

    return nil
}
```

### Alternative: Individual Initialization

If you only need specific components, you can initialize them individually:

```go
// Initialize only logger
logger, loggerShutdown, err := observability.NewLogger(obsConfig, resource)
defer loggerShutdown(ctx)

// Initialize only tracer
tracer, tracerShutdown, err := observability.NewTracer(obsConfig, resource)
defer tracerShutdown(ctx)

// Initialize only metrics
metrics, metricsShutdown, err := observability.NewMetrics(obsConfig, resource)
defer metricsShutdown(ctx)
```

## Best Practices

### 1. Accept Interfaces

Always accept interfaces in `pkg/` packages:

```go
// ✅ Good: Accept interface
func NewClient(logger observability.Logger) *Client

// ❌ Bad: Accept concrete type
func NewClient(logger *observability.OtelLogger) *Client
```

### 2. Optional Parameters Pattern

For optional observability, use functional options:

```go
type ClientOption func(*Client)

func WithLogger(l observability.Logger) ClientOption {
    return func(c *Client) {
        c.logger = l
    }
}

func NewClient(opts ...ClientOption) *Client {
    c := &Client{}
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// Usage
client := NewClient(WithLogger(logger))
```

### 3. Context Propagation

Always pass context through methods to maintain trace correlation:

```go
func (c *Client) DoSomething(ctx context.Context) error {
    // Context contains trace information
    c.logger.Info(ctx, "Doing something")

    // Start child span
    ctx, span := c.tracer.Start(ctx, "operation")
    defer span.End()

    // ... operation
}
```

### 4. Using Tracer

#### Creating Spans

Use spans to track operations and maintain distributed trace context:

```go
func (c *Client) ProcessRequest(ctx context.Context, req *Request) error {
    // Start a new span
    ctx, span := c.tracer.Start(ctx, "client.process_request",
        observability.WithSpanKind(trace.SpanKindClient),
        observability.WithAttributes(
            attribute.String("request.id", req.ID),
            attribute.String("request.type", req.Type),
        ),
    )
    defer span.End()

    // Add events to span
    span.AddEvent("request.received",
        observability.WithEventAttributes(
            attribute.Int("request.size", len(req.Data)),
        ),
    )

    // Record metrics with trace context
    if c.metrics != nil {
        c.requestCounter.Inc(ctx, attribute.String("type", req.Type))
    }

    // Process the request
    result, err := c.doProcess(ctx, req)
    
    if err != nil {
        // Record error on span
        span.RecordError(err,
            observability.WithErrorAttributes(
                attribute.String("error.code", "PROCESSING_FAILED"),
            ),
        )
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    // Set success status
    span.SetStatus(codes.Ok, "")
    span.SetAttributes(attribute.Bool("request.success", true))
    return nil
}
```

#### Nested Spans

Create child spans for sub-operations:

```go
func (c *Client) ComplexOperation(ctx context.Context) error {
    // Parent span
    ctx, parentSpan := c.tracer.Start(ctx, "complex_operation")
    defer parentSpan.End()

    // Child span 1
    ctx, childSpan1 := c.tracer.Start(ctx, "complex_operation.step1")
    // ... do step 1
    childSpan1.End()

    // Child span 2
    ctx, childSpan2 := c.tracer.Start(ctx, "complex_operation.step2")
    // ... do step 2
    childSpan2.End()

    return nil
}
```

#### Span Attributes

Add meaningful attributes to spans for better observability:

```go
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("request.size", len(data)),
    attribute.Bool("cache.hit", cacheHit),
    attribute.Int64("duration_ms", duration.Milliseconds()),
)
```

### 5. Using Metrics

#### Counter Metrics

Use counters for counting events (e.g., requests, errors):

```go
type Service struct {
    metrics observability.Metrics
    requestCounter observability.Counter
    errorCounter   observability.Counter
}

func NewService(metrics observability.Metrics) *Service {
    s := &Service{metrics: metrics}
    
    // Create counters once during initialization
    s.requestCounter = metrics.Counter(
        "service_requests_total",
        observability.WithDescription("Total number of service requests"),
        observability.WithUnit("1"),
    )
    s.errorCounter = metrics.Counter(
        "service_errors_total",
        observability.WithDescription("Total number of service errors"),
        observability.WithUnit("1"),
    )
    
    return s
}

func (s *Service) HandleRequest(ctx context.Context, req *Request) error {
    // Increment counter with attributes
    s.requestCounter.Inc(ctx,
        attribute.String("method", req.Method),
        attribute.String("endpoint", req.Endpoint),
    )
    
    err := s.process(ctx, req)
    if err != nil {
        s.errorCounter.Inc(ctx,
            attribute.String("error.type", err.Error()),
            attribute.String("method", req.Method),
        )
        return err
    }
    
    return nil
}
```

#### Gauge Metrics

Use gauges for values that go up and down (e.g., active connections, queue size):

```go
import (
    "context"
    "sync"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type ConnectionPool struct {
    metrics          observability.Metrics
    activeGauge      observability.Gauge
    activeCount      int64 // Track count separately for gauge
    mu               sync.Mutex // Protect activeCount
}

func NewConnectionPool(metrics observability.Metrics) *ConnectionPool {
    return &ConnectionPool{
        metrics: metrics,
        activeGauge: metrics.Gauge(
            "connection_pool_active",
            observability.WithDescription("Number of active connections"),
            observability.WithUnit("1"),
        ),
        activeCount: 0,
    }
}

func (p *ConnectionPool) AddConnection(ctx context.Context) {
    p.mu.Lock()
    p.activeCount++
    count := p.activeCount
    p.mu.Unlock()
    
    // Update gauge with absolute value
    p.activeGauge.Set(ctx, float64(count))
}

func (p *ConnectionPool) RemoveConnection(ctx context.Context) {
    p.mu.Lock()
    p.activeCount--
    count := p.activeCount
    p.mu.Unlock()
    
    // Update gauge with absolute value
    p.activeGauge.Set(ctx, float64(count))
}

func (p *ConnectionPool) SetActiveConnections(ctx context.Context, count int) {
    p.mu.Lock()
    p.activeCount = int64(count)
    p.mu.Unlock()
    
    // Set absolute value
    p.activeGauge.Set(ctx, float64(count))
}
```

#### Histogram Metrics

Use histograms for measuring distributions (e.g., request duration, response size):

```go
type APIClient struct {
    metrics         observability.Metrics
    requestDuration observability.Histogram
    responseSize    observability.Histogram
}

func NewAPIClient(metrics observability.Metrics) *APIClient {
    return &APIClient{
        metrics: metrics,
        requestDuration: metrics.Histogram(
            "api_request_duration_ms",
            observability.WithDescription("API request duration in milliseconds"),
            observability.WithUnit("ms"),
        ),
        responseSize: metrics.Histogram(
            "api_response_size_bytes",
            observability.WithDescription("API response size in bytes"),
            observability.WithUnit("By"),
        ),
    }
}

func (c *APIClient) CallAPI(ctx context.Context, endpoint string) error {
    startTime := time.Now()
    
    resp, err := c.httpClient.Do(ctx, endpoint)
    duration := time.Since(startTime)
    
    // Record duration
    c.requestDuration.Record(ctx, float64(duration.Milliseconds()),
        attribute.String("endpoint", endpoint),
        attribute.String("status", getStatus(err)),
    )
    
    if err == nil {
        // Record response size
        c.responseSize.Record(ctx, float64(len(resp.Body)),
            attribute.String("endpoint", endpoint),
        )
    }
    
    return err
}
```

#### Labeled Metrics (Vec Pattern)

For metrics with multiple labels, use the Vec pattern:

```go
type HTTPHandler struct {
    metrics observability.Metrics
    requestCounterVec *observability.CounterVec
}

func NewHTTPHandler(metrics observability.Metrics) *HTTPHandler {
    return &HTTPHandler{
        metrics: metrics,
        requestCounterVec: observability.NewCounterVec(
            metrics,
            "http_requests_total",
            []string{"method", "endpoint", "status"}, // Label names
            observability.WithDescription("Total HTTP requests"),
        ),
    }
}

func (h *HTTPHandler) HandleRequest(ctx context.Context, method, endpoint string) {
    // Get counter with specific label values
    counter := h.requestCounterVec.WithLabelValues(method, endpoint, "200")
    counter.Inc(ctx)
}
```

### 6. Service-Level Shutdown

Handle observability shutdown in service main:

```go
func main() {
    obs, err := observability.Initialize(config)
    if err != nil {
        panic(err)
    }

    // Register shutdown handler
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        obs.Shutdown(ctx)
        os.Exit(0)
    }()

    // ... service logic
}
```

### 7. Testing

Use mocks and no-op implementations in tests:

```go
func TestWebSocketClient(t *testing.T) {
    // Use mock logger
    mockLogger := observability.NewMockLogger()
    mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

    client, err := websocket.NewWebSocketClient(url, config, mockLogger)
    // ... test
}

// Or use no-op logger
func TestWebSocketClient_NoOp(t *testing.T) {
    noOpLogger := observability.NewNoOpLogger()
    noOpTracer := observability.NewNoOpTracer()
    noOpMetrics := observability.NewNoOpMetrics()
    client, err := websocket.NewWebSocketClient(url, config, noOpLogger, noOpTracer, noOpMetrics)
    // ... test
}
```

## What NOT to Do

### ❌ Don't Initialize in pkg/ Packages

```go
// ❌ Bad: Don't do this
package websocket

func NewWebSocketClient(url string) (*WebSocketClient, error) {
    // Don't create logger here
    logger, _ := observability.NewLogger(config, resource)
    return &WebSocketClient{logger: logger}, nil
}
```

### ❌ Don't Use Concrete Types

```go
// ❌ Bad: Don't accept concrete types
func NewClient(logger *observability.OtelLogger) *Client

// ✅ Good: Accept interfaces
func NewClient(logger observability.Logger) *Client
```

### ❌ Don't Create Multiple Instances

```go
// ❌ Bad: Multiple logger/tracer/metrics instances
func main() {
    wsLogger, _ := observability.NewLogger(config1, resource)
    redisLogger, _ := observability.NewLogger(config2, resource)
    wsTracer, _ := observability.NewTracer(config1, resource)
    redisTracer, _ := observability.NewTracer(config2, resource)
    // ... creates multiple exporters, processors, etc.
}

// ✅ Good: Single instance
func main() {
    obs, _ := observability.Initialize(config)
    logger := obs.Logger()
    tracer := obs.Tracer()
    metrics := obs.Metrics()
    // Pass same instances to all packages
}
```

### ❌ Don't Create Metric Instruments in Hot Paths

```go
// ❌ Bad: Creating instruments on every request
func (s *Service) HandleRequest(ctx context.Context) {
    counter := s.metrics.Counter("requests_total") // Created every time!
    counter.Inc(ctx)
}

// ✅ Good: Create instruments once during initialization
type Service struct {
    requestCounter observability.Counter
}

func NewService(metrics observability.Metrics) *Service {
    return &Service{
        requestCounter: metrics.Counter("requests_total"),
    }
}

func (s *Service) HandleRequest(ctx context.Context) {
    s.requestCounter.Inc(ctx) // Reuse existing instrument
}
```

### ❌ Don't Forget to End Spans

```go
// ❌ Bad: Span never ends, leaks memory
func (c *Client) DoSomething(ctx context.Context) error {
    ctx, span := c.tracer.Start(ctx, "operation")
    // Forgot defer span.End()
    // ... operation
    return nil
}

// ✅ Good: Always defer span.End()
func (c *Client) DoSomething(ctx context.Context) error {
    ctx, span := c.tracer.Start(ctx, "operation")
    defer span.End() // Always defer
    // ... operation
    return nil
}
```
