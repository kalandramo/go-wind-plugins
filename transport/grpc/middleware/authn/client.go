package authn

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	engine "github.com/tx7do/go-wind-plugins/security/authn"
)

// TokenProvider returns a token string to inject into outgoing gRPC metadata.
// This is used by the client interceptors to authenticate outgoing RPCs.
type TokenProvider func(ctx context.Context) (string, error)

// WithTokenProvider sets a function that provides the auth token for outgoing
// RPCs. The token is injected into the "Authorization" header as
// "<scheme> <token>".
func WithTokenProvider(fn TokenProvider) Option {
	return func(o *options) { o.tokenProvider = fn }
}

// WithScheme sets the authorization scheme for the token.
// Default: "Bearer".
func WithScheme(s string) Option {
	return func(o *options) { o.scheme = s }
}

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that injects
// an authentication token into outgoing unary RPCs.
//
// The token is obtained from the configured [TokenProvider] (set via
// [WithTokenProvider]). If no provider is set, the interceptor is a no-op.
//
// Usage:
//
//	conn, _ := grpc.NewClient(addr,
//	    grpc.WithUnaryInterceptor(
//	        grpcAuthn.UnaryClientInterceptor(
//	            grpcAuthn.WithTokenProvider(func(ctx context.Context) (string, error) {
//	                return getTokenFromVault()
//	            }),
//	        ),
//	    ),
//	)
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	cfg := newClientConfig(opts)

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		ctx = injectToken(ctx, cfg)
		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that injects
// an authentication token into outgoing streaming RPCs.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	cfg := newClientConfig(opts)

	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = injectToken(ctx, cfg)
		return streamer(ctx, desc, cc, method, callOpts...)
	}
}

// newClientConfig builds an options struct from the given options, applying
// defaults suitable for client-side use.
func newClientConfig(opts []Option) *options {
	cfg := &options{
		scheme: engine.BearerWord,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// injectToken calls the token provider (if configured) and injects the
// resulting token into the outgoing gRPC metadata.
func injectToken(ctx context.Context, cfg *options) context.Context {
	if cfg.tokenProvider == nil {
		return ctx
	}
	token, err := cfg.tokenProvider(ctx)
	if err != nil || token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, engine.HeaderAuthorize, cfg.scheme+" "+token)
}
