package logging

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that logs each
// outgoing unary RPC after it completes, including method, status code, and
// latency.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	cfg := &options{
		logger: slog.Default(),
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
		if cfg.skipMethods[method] {
			return invoker(ctx, method, req, reply, cc, callOpts...)
		}

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, callOpts...)
		latency := time.Since(start)

		level := slog.LevelInfo
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() >= codes.Internal {
				level = slog.LevelError
			} else if st.Code() >= codes.NotFound {
				level = slog.LevelWarn
			}
		}

		args := []any{
			slog.String("method", method),
			slog.String("target", cc.Target()),
			slog.Int64("latency_ms", latency.Milliseconds()),
		}
		if err != nil {
			st, _ := status.FromError(err)
			args = append(args,
				slog.String("code", st.Code().String()),
				slog.String("error", err.Error()),
			)
		} else {
			args = append(args, slog.String("code", codes.OK.String()))
		}

		cfg.logger.Log(ctx, level, "grpc unary client rpc", args...)
		return err
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that logs
// each outgoing streaming RPC after the stream is created.
//
// Note: only the stream-creation result is logged. Individual SendMsg/RecvMsg
// errors are not captured.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	cfg := &options{
		logger: slog.Default(),
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
		if cfg.skipMethods[method] {
			return streamer(ctx, desc, cc, method, callOpts...)
		}

		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, callOpts...)
		latency := time.Since(start)

		level := slog.LevelInfo
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() >= codes.Internal {
				level = slog.LevelError
			} else if st.Code() >= codes.NotFound {
				level = slog.LevelWarn
			}
		}

		args := []any{
			slog.String("method", method),
			slog.String("target", cc.Target()),
			slog.Bool("client_stream", desc.ClientStreams),
			slog.Bool("server_stream", desc.ServerStreams),
			slog.Int64("latency_ms", latency.Milliseconds()),
		}
		if err != nil {
			st, _ := status.FromError(err)
			args = append(args,
				slog.String("code", st.Code().String()),
				slog.String("error", err.Error()),
			)
		} else {
			args = append(args, slog.String("code", codes.OK.String()))
		}

		cfg.logger.Log(ctx, level, "grpc stream client rpc", args...)
		return stream, err
	}
}
