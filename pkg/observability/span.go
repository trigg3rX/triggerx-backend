package observability

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// span wraps OpenTelemetry trace.Span with convenience methods
type span struct {
	span trace.Span
}

// End ends the span
func (s *span) End() {
	s.span.End()
}

// SetAttributes sets attributes on the span
func (s *span) SetAttributes(attrs ...attribute.KeyValue) {
	s.span.SetAttributes(attrs...)
}

// AddEvent adds an event to the span
func (s *span) AddEvent(name string, opts ...EventOption) {
	eventOpts := make([]trace.EventOption, 0, len(opts))
	for _, opt := range opts {
		if eventOpt, ok := opt.(eventOption); ok {
			eventOpts = append(eventOpts, eventOpt.opt)
		}
	}
	s.span.AddEvent(name, eventOpts...)
}

// RecordError records an error on the span
func (s *span) RecordError(err error, opts ...RecordErrorOption) {
	errorOpts := make([]trace.EventOption, 0, len(opts))
	for _, opt := range opts {
		if errorOpt, ok := opt.(recordErrorOption); ok {
			errorOpts = append(errorOpts, errorOpt.opt)
		}
	}
	s.span.RecordError(err, errorOpts...)
}

// SetStatus sets the status of the span
func (s *span) SetStatus(code codes.Code, description string) {
	s.span.SetStatus(code, description)
}

// SpanContext returns the span context
func (s *span) SpanContext() trace.SpanContext {
	return s.span.SpanContext()
}

// spanOption wraps trace.SpanStartOption
type spanOption struct {
	opt trace.SpanStartOption
}

func (o spanOption) applySpanOption() {}

// WithSpanKind creates a span option for setting the span kind
func WithSpanKind(kind trace.SpanKind) SpanOption {
	return spanOption{opt: trace.WithSpanKind(kind)}
}

// WithAttributes creates a span option for setting attributes
func WithAttributes(attrs ...attribute.KeyValue) SpanOption {
	return spanOption{opt: trace.WithAttributes(attrs...)}
}

// eventOption wraps trace.EventOption
type eventOption struct {
	opt trace.EventOption
}

func (o eventOption) applyEventOption() {}

// WithEventAttributes creates an event option for setting attributes
func WithEventAttributes(attrs ...attribute.KeyValue) EventOption {
	return eventOption{opt: trace.WithAttributes(attrs...)}
}

// recordErrorOption wraps trace.EventOption for error recording
type recordErrorOption struct {
	opt trace.EventOption
}

func (o recordErrorOption) applyRecordErrorOption() {}

// WithErrorAttributes creates a record error option for setting attributes
func WithErrorAttributes(attrs ...attribute.KeyValue) RecordErrorOption {
	return recordErrorOption{opt: trace.WithAttributes(attrs...)}
}
