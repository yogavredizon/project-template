package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Config struct {
	Host          string
	ExcludeRoutes map[string]struct{}
	Probability   float64
	ServiceName   string
}

func InitTracing(cfg *Config) (trace.TracerProvider, func(ctx context.Context), error) {
	exporter, err := otlptrace.New(context.Background(), otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Host),
	))

	if err != nil {
		return nil, nil, err
	}

	var traceProvider trace.TracerProvider
	teardown := func(context.Context) {}

	switch cfg.Host {
	case "":
		traceProvider = noop.NewTracerProvider()
	default:
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.ParentBased(newExcluderEndpoints(cfg.ExcludeRoutes, cfg.Probability))),
			sdktrace.WithBatcher(exporter,
				sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
				sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			),
			sdktrace.WithResource(
				resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNameKey.String(cfg.ServiceName),
				),
			),
		)

		teardown = func(ctx context.Context) {
			tp.Shutdown(ctx)
		}

		traceProvider = tp
	}

	otel.SetTracerProvider(traceProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider, teardown, nil
}

func AddSpan(parent context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	v, ok := parent.Value("tracer").(trace.Tracer)
	if !ok {
		return parent, trace.SpanFromContext(parent)
	}

	ctx, span := v.Start(parent, spanName)
	span.SetAttributes(keyValues...)

	return ctx, span
}
