package ratelimit

import (
	"context"

	"github.com/tx7do/go-wind-plugins/ratelimit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that enforces
// rate-limiting on outgoing unary RPCs.
func UnaryClientInterceptor(limiter ratelimit.Limiter, opts ...Option) grpc.UnaryClientInterceptor {
	cfg := &options{}
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

		if cfg.waitMode {
			if err := limiter.Wait(ctx); err != nil {
				return status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
		} else {
			ok, err := limiter.Allow()
			if !ok || err != nil {
				return status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
		}

		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that enforces
// rate-limiting on outgoing streaming RPCs.
func StreamClientInterceptor(limiter ratelimit.Limiter, opts ...Option) grpc.StreamClientInterceptor {
	cfg := &options{}
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

		if cfg.waitMode {
			if err := limiter.Wait(ctx); err != nil {
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
		} else {
			ok, err := limiter.Allow()
			if !ok || err != nil {
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
		}

		return streamer(ctx, desc, cc, method, callOpts...)
	}
}
