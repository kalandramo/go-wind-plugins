// Package tracer provides distributed tracing initialization helpers
// for the go-wind framework.
//
// Unlike other plugin domains (config, registry, log) that require custom
// abstraction interfaces to bridge fundamentally different engines, tracing
// has a de facto standard: OpenTelemetry. There is no need for a custom
// Provider interface — users should use the standard OTel types directly.
//
// The otlp sub-package provides a convenience constructor that wires up the
// OTLP exporter, sampler, resource, and global propagator in one call:
//
//	tp, err := otlp.New(
//	    otlp.WithEndpoint("localhost:4317"),
//	    otlp.WithServiceName("my-service"),
//	)
//	defer tp.Shutdown(context.Background())
//
// After initialization, use the standard OpenTelemetry API:
//
//	tracer := tp.Tracer("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
package tracer
