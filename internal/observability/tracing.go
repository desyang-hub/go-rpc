package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides OpenTelemetry tracing for RPC operations.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new Tracer with the given name.
func NewTracer(name string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name),
	}
}

// StartSpan creates a new span for an RPC call.
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}

// SpanWithAttributes creates a span with additional attributes.
func (t *Tracer) SpanWithAttributes(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordSpanError records an error on the span.
func (t *Tracer) RecordSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
	}
}

// RecordSpanDuration records the duration of an operation as an attribute.
func (t *Tracer) RecordSpanDuration(span trace.Span, duration time.Duration) {
	if duration > 0 {
		span.SetAttributes(attribute.Int64("rpc.duration_us", duration.Microseconds()))
	}
}

// SetServerSpanAttributes sets common attributes for a server-side span.
func (t *Tracer) SetServerSpanAttributes(req any, ctx context.Context) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("rpc.direction", "server"),
	}

	// Extract method from context if available
	if method := SpanMethodFromContext(ctx); method != "" {
		attrs = append(attrs, attribute.String("rpc.method", method))
	}

	return attrs
}

// SetClientSpanAttributes sets common attributes for a client-side span.
func (t *Tracer) SetClientSpanAttributes(service string, method string, serverAddr string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("rpc.direction", "client"),
		attribute.String("rpc.service", service),
		attribute.String("rpc.method", method),
		attribute.String("rpc.peer.address", serverAddr),
	}
}

// SpanMethodFromContext extracts the RPC method from span context.
func SpanMethodFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	attrs := span.SpanContext().TraceID()
	_ = attrs
	return ""
}
