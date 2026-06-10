package retry

import (
	"context"

	coreRetry "github.com/tx7do/go-wind-plugins/retry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that retries
// idempotent outgoing RPCs when the server returns a transient failure.
//
// This is typically the most important client interceptor — it handles
// network blips, server restarts, and transient failures transparently.
//
// Usage:
//
//	r := retry.New(retry.WithMaxAttempts(3))
//	conn, _ := grpc.NewClient(addr,
//	    grpc.WithUnaryInterceptor(grpcRetry.UnaryClientInterceptor(r)),
//	)
func UnaryClientInterceptor(r *coreRetry.Retrier, opts ...Option) grpc.UnaryClientInterceptor {
	cfg := &options{
		idempotentPrefixes: defaultIdempotentPrefixes,
		retryCodes:         defaultRetryCodes,
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
		if cfg.skipMethods[method] || !isIdempotent(method, cfg.idempotentPrefixes) {
			return invoker(ctx, method, req, reply, cc, callOpts...)
		}

		var lastErr error

		_ = r.Do(ctx, func(attemptCtx context.Context) error {
			err := invoker(attemptCtx, method, req, reply, cc, callOpts...)
			lastErr = err

			if err == nil {
				return nil
			}

			st, _ := status.FromError(err)
			if cfg.retryCodes[st.Code()] {
				return err // retryable
			}
			return nil // not retryable — stop
		})

		return lastErr
	}
}

// StreamClientInterceptor is intentionally NOT provided.
//
// Retrying streaming RPCs safely requires replaying all sent messages, which
// is application-specific and cannot be done generically. For streaming RPCs
// that need resilience, consider wrapping the stream consumer with a
// [retry.Retrier] at the application layer.
