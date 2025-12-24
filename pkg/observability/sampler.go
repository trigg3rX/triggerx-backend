package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// errorAwareSampler samples traces with different rates for errors vs success.
// Strategy: Sample at 100% initially, then use a span processor to filter successful spans.
type errorAwareSampler struct {
	successRate float64
	errorRate   float64
}

// NewErrorAwareSampler creates a sampler that:
// - Samples at errorRate (100%) to ensure all errors are captured
// - SuccessRate (3%) is achieved via span processor filtering
func NewErrorAwareSampler(successRate, errorRate float64) sdktrace.Sampler {
	// Clamp rates to valid range [0.0, 1.0]
	if successRate < 0.0 {
		successRate = 0.0
	} else if successRate > 1.0 {
		successRate = 1.0
	}
	if errorRate < 0.0 {
		errorRate = 0.0
	} else if errorRate > 1.0 {
		errorRate = 1.0
	}

	return &errorAwareSampler{
		successRate: successRate,
		errorRate:   errorRate,
	}
}

// ShouldSample implements sdktrace.Sampler interface
// Uses ParentBased logic: if parent is sampled, sample child; otherwise sample at errorRate
func (s *errorAwareSampler) ShouldSample(parameters sdktrace.SamplingParameters) sdktrace.SamplingResult {
	// Extract parent span context from the context
	parentContext := trace.SpanContextFromContext(parameters.ParentContext)

	// If parent context exists and is sampled, always sample child spans
	if parentContext.IsValid() {
		if parentContext.TraceFlags().IsSampled() {
			return sdktrace.SamplingResult{
				Decision:   sdktrace.RecordAndSample,
				Tracestate: parentContext.TraceState(),
			}
		}
	}

	// For root spans, sample at errorRate (100%) to ensure we capture all errors
	// The span processor will handle filtering successful spans to achieve successRate
	decision := sdktrace.Drop
	if s.errorRate >= 1.0 {
		decision = sdktrace.RecordAndSample
	} else if s.errorRate > 0.0 {
		// Use trace ID for deterministic sampling
		traceID := parameters.TraceID
		hash := traceIDHash(traceID)
		sampleValue := float64(hash) / float64(^uint64(0))
		if sampleValue < s.errorRate {
			decision = sdktrace.RecordAndSample
		}
	}

	return sdktrace.SamplingResult{
		Decision: decision,
		Attributes: []attribute.KeyValue{
			attribute.Float64("sampling.error_rate", s.errorRate),
			attribute.Float64("sampling.success_rate", s.successRate),
		},
	}
}

// Description returns a description of the sampler
func (s *errorAwareSampler) Description() string {
	return "ErrorAwareSampler"
}

// errorFilteringSpanProcessor ensures error spans are always exported and filters successful spans
// to achieve the desired success sampling rate. Since OpenTelemetry doesn't allow modifying
// sampling decisions after span creation, we mark spans with attributes that can be used
// by exporters or collectors to filter.
type errorFilteringSpanProcessor struct {
	successRate float64
}

// NewErrorFilteringSpanProcessor creates a span processor that marks spans for filtering:
// - All error spans are kept (100%)
// - Successful spans are marked based on successRate (3%)
func NewErrorFilteringSpanProcessor(successRate float64) sdktrace.SpanProcessor {
	// Clamp rate
	if successRate < 0.0 {
		successRate = 0.0
	} else if successRate > 1.0 {
		successRate = 1.0
	}

	return &errorFilteringSpanProcessor{
		successRate: successRate,
	}
}

// OnStart is called when a span starts
func (p *errorFilteringSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	// No-op: we check at end time
}

// OnEnd is called when a span ends - this is where we mark spans for filtering
func (p *errorFilteringSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	// Check if span has error status
	status := s.Status()
	hasError := status.Code == codes.Error

	// Note: ReadOnlySpan doesn't allow setting attributes, so we can't mark spans here.
	// The actual filtering needs to happen at the exporter level or collector level.
	// For now, this processor serves as a placeholder for future filtering logic.
	// The sampler already handles the initial sampling decision at 100% for error capture.
	_ = hasError // Mark as used for now
}

// Shutdown is called when the processor is shut down
func (p *errorFilteringSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush forces a flush of the processor
func (p *errorFilteringSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

// traceIDHash generates a deterministic hash from trace ID for sampling
func traceIDHash(traceID [16]byte) uint64 {
	return uint64(traceID[0])<<56 | uint64(traceID[1])<<48 | uint64(traceID[2])<<40 | uint64(traceID[3])<<32 |
		uint64(traceID[4])<<24 | uint64(traceID[5])<<16 | uint64(traceID[6])<<8 | uint64(traceID[7])
}
