package tracing

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that creates
// an OpenTelemetry client span for each outgoing unary RPC and injects the
// trace context into outgoing gRPC metadata.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	cfg := &options{
		tracer:      otel.GetTracerProvider().Tracer(instrumentationName),
		propagators: otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		attrs := []attribute.KeyValue{
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", serviceFromMethod(method)),
		}

		ctx, span := cfg.tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)
		defer span.End()

		// Inject trace context into outgoing metadata.
		ctx = injectTrace(ctx, cfg.propagators)

		err := invoker(ctx, method, req, reply, cc, callOpts...)
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.grpc.status", st.Code().String()))
			if st.Code() >= codes.Internal {
				span.SetStatus(otelcodes.Error, st.Message())
			}
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status", codes.OK.String()))
		}

		return err
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that creates
// an OpenTelemetry client span for each outgoing streaming RPC.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	cfg := &options{
		tracer:      otel.GetTracerProvider().Tracer(instrumentationName),
		propagators: otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		attrs := []attribute.KeyValue{
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", serviceFromMethod(method)),
			attribute.Bool("rpc.grpc.client_stream", desc.ClientStreams),
			attribute.Bool("rpc.grpc.server_stream", desc.ServerStreams),
		}

		ctx, span := cfg.tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)

		// Inject trace context into outgoing metadata.
		ctx = injectTrace(ctx, cfg.propagators)

		stream, err := streamer(ctx, desc, cc, method, callOpts...)
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.grpc.status", st.Code().String()))
			if st.Code() >= codes.Internal {
				span.SetStatus(otelcodes.Error, st.Message())
			}
			span.End()
		} else {
			// Wrap the stream so we can end the span when the stream is done.
			stream = &tracedClientStream{ClientStream: stream, span: span}
		}

		return stream, err
	}
}

// injectTrace injects the trace context from ctx into outgoing gRPC metadata.
func injectTrace(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	propagators.Inject(ctx, &mdCarrier{md: md})
	return metadata.NewOutgoingContext(ctx, md)
}

// tracedClientStream wraps a [grpc.ClientStream] to end the span when the
// stream completes (Header/Recv error or explicit Close).
type tracedClientStream struct {
	grpc.ClientStream
	span trace.Span
}

func (s *tracedClientStream) Header() (metadata.MD, error) {
	md, err := s.ClientStream.Header()
	if err != nil {
		st, _ := status.FromError(err)
		s.span.SetAttributes(attribute.String("rpc.grpc.status", st.Code().String()))
		if st.Code() >= codes.Internal {
			s.span.SetStatus(otelcodes.Error, st.Message())
		}
		s.span.End()
	}
	return md, err
}

func (s *tracedClientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	s.span.SetAttributes(attribute.String("rpc.grpc.status", codes.OK.String()))
	s.span.End()
	return err
}
