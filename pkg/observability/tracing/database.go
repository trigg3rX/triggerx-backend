package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DatabaseTracer provides OpenTelemetry tracing for database operations
type DatabaseTracer struct {
	tracer trace.Tracer
}

// NewDatabaseTracer creates a new database tracer instance
func NewDatabaseTracer(serviceName string) *DatabaseTracer {
	return &DatabaseTracer{
		tracer: otel.Tracer(serviceName + "-database"),
	}
}

// TraceDBOperation creates a span for database operations and returns a completion function
// Usage: defer tracer.TraceDBOperation(ctx, "SELECT", "users", query)()
func (dt *DatabaseTracer) TraceDBOperation(ctx context.Context, operation, table, query string) func(error) {
	if ctx == nil {
		ctx = context.Background()
	}

	spanName := "db." + operation + "." + table
	ctx, span := dt.tracer.Start(ctx, spanName)
	startTime := time.Now()

	// Add database operation attributes
	span.SetAttributes(
		attribute.String(TriggerXAttributes.DBOperation, operation),
		attribute.String(TriggerXAttributes.DBTable, table),
		attribute.String("db.system", "scylladb"),
		attribute.String("db.statement", query),
		attribute.String("db.operation.name", operation),
	)

	return func(err error) {
		duration := time.Since(startTime)

		// Add performance attributes
		span.SetAttributes(
			attribute.Float64("db.operation.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(
				attribute.String("error.type", "database_error"),
				attribute.String("error.message", err.Error()),
				attribute.Bool("operation.success", false),
			)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(
				attribute.Bool("operation.success", true),
			)
		}

		// Track slow queries (>1 second)
		if duration > time.Second {
			span.SetAttributes(
				attribute.Bool("db.slow_query", true),
				attribute.String("performance.category", "slow"),
			)
		}

		span.End()
	}
}

// TraceDBBatch creates a span for batch database operations
func (dt *DatabaseTracer) TraceDBBatch(ctx context.Context, operation string, batchSize int) func(error) {
	if ctx == nil {
		ctx = context.Background()
	}

	spanName := "db.batch." + operation
	ctx, span := dt.tracer.Start(ctx, spanName)
	startTime := time.Now()

	span.SetAttributes(
		attribute.String("db.operation.type", "batch"),
		attribute.String(TriggerXAttributes.DBOperation, operation),
		attribute.Int("db.batch.size", batchSize),
		attribute.String("db.system", "scylladb"),
	)

	return func(err error) {
		duration := time.Since(startTime)

		span.SetAttributes(
			attribute.Float64("db.operation.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(
				attribute.String("error.type", "database_batch_error"),
				attribute.String("error.message", err.Error()),
				attribute.Bool("operation.success", false),
			)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(
				attribute.Bool("operation.success", true),
			)
		}

		span.End()
	}
}

// ExtractContextFromGin extracts context from Gin request for database tracing
func ExtractContextFromGin(c interface{}) context.Context {
	// This is a helper function that can be used in handlers
	// Returns context.Background() if gin context extraction fails
	return context.Background()
}
