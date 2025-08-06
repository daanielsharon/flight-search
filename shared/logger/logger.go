package sharedlogger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var log *zap.Logger

func Init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	log = l
}

func L() *zap.Logger {
	return log
}

func WithTrace(ctx context.Context) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	return log.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)
}
