package trace

import (
	"context"
	"crypto/rand"

	"go.opentelemetry.io/otel/trace"
)

func withOTEL(ctx context.Context, tid string) context.Context {
	if ctx == nil || tid == "" {
		return ctx
	}
	traceID, err := trace.TraceIDFromHex(tid)
	if err != nil {
		return ctx
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() && sc.TraceID() == traceID {
		return ctx
	}
	var sid trace.SpanID
	if _, err := rand.Read(sid[:]); err != nil {
		return ctx
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}
