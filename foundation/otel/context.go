package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey int

const (
	tracerKey ctxKey = iota + 1
	traceIDKey
)

func setTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

// AddSpan adds an otel span to the existing trace.
func AddSpan(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	v, ok := ctx.Value(tracerKey).(trace.Tracer)
	if !ok || v == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := v.Start(ctx, spanName)
	for _, kv := range keyValues {
		span.SetAttributes(kv)
	}

	return ctx, span
}

func setTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}

	return v
}
