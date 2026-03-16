/*
Open telemetry initialization and setup
*
* Every microservice needs the same OTel setup, so instead of copy-pasting,
* we centralize it here. Please follow these patterns and use telemetry.Init() at startup.
*
* what it does
* 1. Creates a "resource" - metadata about this service (name, version, env)
* 2. Creates a "tracer provider" - the thing that creates traces
* 3. Creates a "meter provider" - the thing that creates metrics
* 4. Sets up "propagation" - how trace context travels between services
*/
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds the configuration for telemetry initialization.
// You'll typically populate this from environment variables.
type Config struct {
	ServiceName       string // e.g., "stats-service", "auth-service"
	ServiceVersion    string // e.g., "1.0.0" - useful for tracking deployments
	Environment       string // e.g., "development", "staging", "production"
	CollectorEndpoint string // e.g., "localhost:4317" - where the OTel Collector lives
}

// Init initializes OpenTelemetry and returns a shutdown function.
//
// IMPORTANT: Call the shutdown function when your service stops!
// This ensures all pending traces are flushed before exit.
//
// Usage:
//
//	shutdown, err := telemetry.Init(ctx, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(ctx)
func Init(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	// STEP 1: Create a Resource
	// A Resource describes WHO is producing the telemetry.
	// Think of it as metadata that gets attached to every trace and metric.
	//
	// When you look at traces in Grafana, you'll filter by service.name
	// to find "show me all traces from stats-service"
	//
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironment(cfg.Environment),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// STEP 2: Create the Trace Exporter
	// The exporter is HOW traces leave your application.
	// We're using OTLP (OpenTelemetry Protocol) over gRPC.
	//
	// OTLP is the standard protocol - it works with:
	// - OpenTelemetry Collector (what we'll use)
	// - Jaeger (directly)
	// - Grafana Tempo (directly)
	// - Most commercial APM tools
	//
	// The Collector then forwards to your actual storage (Tempo, Jaeger, etc.)
	//
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlptracegrpc.WithInsecure(), // TODO: Use TLS in production!
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("creating trace exporter: %w", err)
	}

	// STEP 3: Create the Tracer Provider
	// The TracerProvider is the FACTORY that creates Tracers.
	// You don't use it directly much - you just set it globally with otel.SetTracerProvider()
	// Then anywhere in your code, otel.Tracer("name") returns a tracer.
	//
	// BatchSpanProcessor: Batches spans before sending (better performance)
	// vs SimpleSpanProcessor: Sends immediately (good for debugging, bad for prod)
	//
	tracerProvider := trace.NewTracerProvider(
		// WithBatcher: Collects spans and sends them in batches
		// This is MUCH more efficient than sending each span immediately
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second), // Send batch every 5s or when full
		),
		trace.WithResource(res),

		// WithSampler: Controls WHICH traces to record
		// AlwaysSample() = record everything (good for dev, expensive in prod)
		// TraceIDRatioBased(0.1) = record 10% of traces (common in prod)
		// ParentBased() = if parent was sampled, sample this too (preserves full traces)
		trace.WithSampler(trace.AlwaysSample()), // We'll change this for prod later
	)

	// Set as the global tracer provider
	// Now otel.Tracer("any-name") anywhere in your code will use this
	otel.SetTracerProvider(tracerProvider)

	// STEP 4: Set up Context Propagation
	// THIS IS CRITICAL FOR DISTRIBUTED TRACING!
	//
	// When Service A calls Service B, how does B know it's part of A's trace?
	// Answer: A injects trace context into the request (headers for HTTP, metadata for gRPC)
	//         B extracts it and continues the same trace.
	//
	// TraceContext = W3C standard format (traceparent header)
	// Baggage = key-value pairs that travel with the trace (like user_id)
	//
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // The trace_id, span_id, flags
		propagation.Baggage{},      // Optional: carry business data across services
	))

	// Meter provider
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("creating metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(10*time.Second),
			),
		),
	)
	otel.SetMeterProvider(meterProvider)

	// STEP 5: Return Shutdown Function
	// we always call this on application shutdown
	// it flushes any pending traces so you don't lose data.
	//
	shutdown = func(ctx context.Context) error {
		err1 := tracerProvider.Shutdown(ctx)
		err2 := meterProvider.Shutdown(ctx)

		return errors.Join(err1, err2)
	}

	return shutdown, nil
}
