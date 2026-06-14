package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "status-monitor"

// Tracer is the global tracer for the application.
var Tracer trace.Tracer

// InitTracing sets up the OTel TracerProvider.
// If OTEL_EXPORTER_OTLP_ENDPOINT is set, uses OTLP HTTP; otherwise stdout (dev).
// Returns a shutdown function that must be called on exit.
func InitTracing(ctx context.Context, isProd bool) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithOS(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, err
	}

	var exporter sdktrace.SpanExporter
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		)
	} else if !isProd {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	} else {
		// Production without OTLP configured — use no-op (traces go nowhere)
		tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		otel.SetTracerProvider(tp)
		Tracer = otel.Tracer(serviceName)
		return tp.Shutdown, nil
	}
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	Tracer = otel.Tracer(serviceName)
	return tp.Shutdown, nil
}
