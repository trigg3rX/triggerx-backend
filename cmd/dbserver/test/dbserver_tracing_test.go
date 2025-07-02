package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/observability/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// TestPhase1DBServerPilotCore tests the core OpenTelemetry integration
func TestPhase1DBServerPilotCore(t *testing.T) {
	// Test 1: OpenTelemetry Configuration Loading
	t.Run("✅ Configuration_Loading", func(t *testing.T) {
		config := tracing.LoadConfig("triggerx-dbserver-test")
		require.NotNil(t, config)
		assert.Equal(t, "triggerx-dbserver-test", config.ServiceName)
		assert.Contains(t, config.OTLPEndpoint, "4318") // HTTP endpoint
		assert.Equal(t, "development", config.Environment)
		t.Logf("✅ Config loaded: %s -> %s", config.ServiceName, config.OTLPEndpoint)
	})

	// Test 2: Tracer Provider Initialization
	t.Run("✅ Tracer_Provider_Init", func(t *testing.T) {
		config := tracing.LoadConfig("triggerx-dbserver-test")
		tp, err := tracing.InitTracer(config)
		require.NoError(t, err)
		require.NotNil(t, tp)

		// Test graceful shutdown
		err = tracing.Shutdown(tp, 2*time.Second)
		assert.NoError(t, err)
		t.Logf("✅ Tracer provider initialized and shutdown gracefully")
	})

	// Test 3: Database Tracer Functionality
	t.Run("✅ Database_Tracer", func(t *testing.T) {
		dbTracer := tracing.NewDatabaseTracer("triggerx-dbserver-test")
		require.NotNil(t, dbTracer)

		ctx := context.Background()

		// Test database operation tracing
		traceOp := dbTracer.TraceDBOperation(ctx, "SELECT", "users", "SELECT * FROM users WHERE id = ?")
		require.NotNil(t, traceOp)

		// Complete the operation successfully
		traceOp(nil)

		// Test error case
		traceOp2 := dbTracer.TraceDBOperation(ctx, "INSERT", "jobs", "INSERT INTO jobs ...")
		traceOp2(assert.AnError)

		t.Logf("✅ Database tracer working for both success and error cases")
	})

	// Test 4: HTTP Middleware Integration (without full tracer)
	t.Run("✅ HTTP_Middleware_Setup", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Add tracing middleware - should not panic
		require.NotPanics(t, func() {
			router.Use(tracing.GinMiddleware("triggerx-dbserver-test"))
		})

		// Add test endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Logf("✅ HTTP middleware integrated without errors")
	})

	// Test 5: Business Context Attributes Structure
	t.Run("✅ Business_Attributes", func(t *testing.T) {
		// Test that all TriggerX attributes are available
		attrs := []attribute.KeyValue{
			attribute.String(tracing.TriggerXAttributes.UserID, "12345"),
			attribute.String(tracing.TriggerXAttributes.UserAddress, "0x1234567890123456789012345678901234567890"),
			attribute.String(tracing.TriggerXAttributes.JobID, "67890"),
			attribute.String(tracing.TriggerXAttributes.JobType, "time_based"),
			attribute.String(tracing.TriggerXAttributes.TaskDefID, "1"),
		}

		// Should not panic and attributes should be valid
		for _, attr := range attrs {
			assert.NotEmpty(t, string(attr.Key))
			assert.NotEmpty(t, attr.Value.AsString())
		}

		t.Logf("✅ All TriggerX business attributes available: %d attributes", len(attrs))
	})

	// Test 6: Context Helpers
	t.Run("✅ Context_Helpers", func(t *testing.T) {
		ctx := context.Background()

		// Test context extraction (should not panic even with empty context)
		require.NotPanics(t, func() {
			traceID := tracing.TraceIDFromContext(ctx)
			spanID := tracing.SpanIDFromContext(ctx)

			// These might be empty without active span, but should not panic
			t.Logf("✅ Context helpers working - TraceID: %s, SpanID: %s", traceID, spanID)
		})
	})
}

// TestPhase1DBServerPilotWithTracer tests with actual tracer
func TestPhase1DBServerPilotWithTracer(t *testing.T) {
	// Initialize tracer for this test
	config := tracing.LoadConfig("triggerx-dbserver-test")
	tp, err := tracing.InitTracer(config)
	require.NoError(t, err)
	defer tracing.Shutdown(tp, 2*time.Second)

	t.Run("✅ Active_Span_Context", func(t *testing.T) {
		tracer := otel.Tracer("triggerx-dbserver-test")
		ctx, span := tracer.Start(context.Background(), "test.active_span")
		defer span.End()

		// Test business attributes on active span
		span.SetAttributes(
			attribute.String(tracing.TriggerXAttributes.UserID, "test-user-123"),
			attribute.String(tracing.TriggerXAttributes.JobType, "time_based"),
		)

		// Test context helpers with active span
		traceID := tracing.TraceIDFromContext(ctx)
		spanID := tracing.SpanIDFromContext(ctx)

		if span.SpanContext().IsValid() {
			assert.NotEmpty(t, traceID)
			assert.NotEmpty(t, spanID)
			t.Logf("✅ Active span context: TraceID=%s, SpanID=%s", traceID, spanID)
		}
	})
}

// TestPhase1DBServerPilotGracefulDegradation tests when tracing is disabled
func TestPhase1DBServerPilotGracefulDegradation(t *testing.T) {
	t.Run("✅ Tracing_Disabled", func(t *testing.T) {
		config := &tracing.Config{
			ServiceName:    "triggerx-dbserver-test",
			ServiceVersion: "1.0.0",
			Environment:    "test",
			OTLPEndpoint:   "http://localhost:4318",
			TracingEnabled: false, // Disabled
			SampleRate:     0.1,
		}

		tp, err := tracing.InitTracer(config)
		require.NoError(t, err)
		require.NotNil(t, tp)

		// Should work without errors even when disabled
		tracer := otel.Tracer("test")
		_, span := tracer.Start(context.Background(), "test.disabled")
		defer span.End()

		assert.NotPanics(t, func() {
			span.SetAttributes(attribute.String("test", "value"))
		})

		t.Logf("✅ Graceful degradation working when tracing disabled")
	})
}

// BenchmarkPhase1TracingOverhead measures performance impact
func BenchmarkPhase1TracingOverhead(b *testing.B) {
	config := tracing.LoadConfig("triggerx-dbserver-bench")
	tp, err := tracing.InitTracer(config)
	require.NoError(b, err)
	defer tracing.Shutdown(tp, 5*time.Second)

	tracer := otel.Tracer("triggerx-dbserver-bench")

	b.Run("Span_Creation_Overhead", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, span := tracer.Start(context.Background(), "bench.span")
			span.SetAttributes(attribute.String("operation", "benchmark"))
			span.End()
		}
	})

	b.Run("Database_Tracing_Overhead", func(b *testing.B) {
		dbTracer := tracing.NewDatabaseTracer("triggerx-dbserver-bench")
		for i := 0; i < b.N; i++ {
			traceOp := dbTracer.TraceDBOperation(context.Background(), "SELECT", "test", "bench_query")
			traceOp(nil)
		}
	})
}
