package otlp

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// Option configures the OTLP tracer provider.
type Option func(*options)

type options struct {
	endpoint       string
	serviceName    string
	serviceVersion string
	insecure       bool
	useHTTP        bool
	sampleRatio    float64
	batchTimeout   time.Duration
	exportTimeout  time.Duration
	headers        map[string]string
}

func defaultOptions() *options {
	return &options{
		endpoint:       "localhost:4317",
		serviceName:    "go-wind-service",
		serviceVersion: "v0.0.1",
		insecure:       false,
		useHTTP:        false, // 默认使用 gRPC
		sampleRatio:    1.0,   // 全量采样
		batchTimeout:   5 * time.Second,
		exportTimeout:  30 * time.Second,
		headers:        make(map[string]string),
	}
}

// WithEndpoint sets the OTLP collector endpoint.
func WithEndpoint(endpoint string) Option {
	return func(o *options) {
		o.endpoint = endpoint
	}
}

// WithServiceName sets the service name.
func WithServiceName(name string) Option {
	return func(o *options) {
		o.serviceName = name
	}
}

// WithServiceVersion sets the service version.
func WithServiceVersion(version string) Option {
	return func(o *options) {
		o.serviceVersion = version
	}
}

// WithInsecure disables TLS for gRPC connection.
func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
	}
}

// WithHTTP uses HTTP protocol instead of gRPC.
func WithHTTP(useHTTP bool) Option {
	return func(o *options) {
		o.useHTTP = useHTTP
	}
}

// WithSampleRatio sets the sampling ratio (0.0 to 1.0).
func WithSampleRatio(ratio float64) Option {
	return func(o *options) {
		o.sampleRatio = ratio
	}
}

// WithBatchTimeout sets the batch export timeout.
func WithBatchTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.batchTimeout = timeout
	}
}

// WithExportTimeout sets the export request timeout.
func WithExportTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.exportTimeout = timeout
	}
}

// WithHeaders sets additional headers for OTLP requests.
func WithHeaders(headers map[string]string) Option {
	return func(o *options) {
		o.headers = headers
	}
}

// New creates and registers an OTLP-backed OpenTelemetry TracerProvider.
//
// It configures the OTLP exporter (gRPC or HTTP), sampler, resource attributes,
// and sets the global TracerProvider and W3C trace-context propagator.
//
// The returned *sdktrace.TracerProvider should be shut down on exit:
//
//	tp, _ := otlp.New(otlp.WithEndpoint("localhost:4317"))
//	defer tp.Shutdown(context.Background())
//
// After initialization, use the standard OpenTelemetry API:
//
//	tracer := tp.Tracer("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
func New(opts ...Option) (*sdktrace.TracerProvider, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	ctx := context.Background()

	// Create OTLP exporter
	var exporter *otlptrace.Exporter
	var err error

	if cfg.useHTTP {
		httpOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.endpoint),
			otlptracehttp.WithHeaders(cfg.headers),
			otlptracehttp.WithTimeout(cfg.exportTimeout),
		}
		if cfg.insecure {
			httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptrace.New(ctx, otlptracehttp.NewClient(httpOpts...))
	} else {
		grpcOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.endpoint),
			otlptracegrpc.WithHeaders(cfg.headers),
			otlptracegrpc.WithTimeout(cfg.exportTimeout),
		}
		if cfg.insecure {
			grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient(grpcOpts...))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.serviceName),
			semconv.ServiceVersion(cfg.serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.sampleRatio)),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.batchTimeout),
		),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
