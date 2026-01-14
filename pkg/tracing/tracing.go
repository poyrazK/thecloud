// Package tracing provides OpenTelemetry instrumentation and initialization for the application.
package tracing

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes the OpenTelemetry tracer provider with Jaeger exporter.
// It returns a TracerProvider and an error if initialization fails.
//
// serviceName: The name of the service (e.g., "thecloud-api")
// jaegerEndpoint: The Jaeger OTLP HTTP endpoint (e.g., "http://localhost:4318")
func InitTracer(ctx context.Context, serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
	// 1. Create OTLP HTTP Exporter
	exporter, err := otlptracehttp.New(ctx,
		// Insecure by default for local development
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpointURL(jaegerEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// 2. Create Resource with Service Name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(envOr("ENV", "development")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 3. Create Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// Sample 100% of requests in development, can be configured for prod
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 4. Set Global Tracer Provider and Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// InitNoopTracer initializes a no-op tracer provider for when tracing is disabled.
func InitNoopTracer() *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider()
}

// InitConsoleTracer initializes a tracer that prints to stdout (for debugging without Jaeger).
func InitConsoleTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
