package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/dubonzi/mantis"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	span := trace.SpanFromContext(ctx)
	return span.TracerProvider().Tracer(tracerName).Start(ctx, name, opts...)
}
