package trace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// getting trace id based on opentelemetry standard
func GetTraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasTraceID() {
		return ""
	}
	return spanCtx.TraceID().String()
}
