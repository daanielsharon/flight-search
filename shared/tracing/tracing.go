package tracing

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tp *sdktrace.TracerProvider

func MustInit(serviceName string) {
	ctx := context.Background()
	if err := Init(ctx, serviceName); err != nil {
		log.Fatalf("failed to init tracing for %s: %v", serviceName, err)
	}
}

func Init(ctx context.Context, serviceName string) error {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return err
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return nil
}

func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown tracer provider: %v", err)
	}
}

func InjectTracingToMap(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ExtractTracingFromMap(ctx context.Context, carrierMap any) context.Context {
	m, ok := carrierMap.(map[string]any)
	if !ok {
		return ctx
	}
	strMap := map[string]string{}
	for k, v := range m {
		if s, ok := v.(string); ok {
			strMap[k] = s
		}
	}
	carrier := propagation.MapCarrier(strMap)
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
