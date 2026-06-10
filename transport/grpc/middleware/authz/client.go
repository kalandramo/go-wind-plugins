package authz

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authzEngine "github.com/tx7do/go-wind-plugins/security/authz"
)

// ClientResolverFunc maps a gRPC method string to an action or resource.
type ClientResolverFunc func(method string) string

// WithClientActionResolver sets a custom function to derive the authorization
// action from the outgoing method name.
func WithClientActionResolver(fn ClientResolverFunc) Option {
	return func(o *options) { o.clientActionResolver = fn }
}

// WithClientResourceResolver sets a custom function to derive the authorization
// resource from the outgoing method name.
func WithClientResourceResolver(fn ClientResolverFunc) Option {
	return func(o *options) { o.clientResourceResolver = fn }
}

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that performs
// a pre-flight authorization check before sending the RPC to the server.
//
// This is useful when the client wants to avoid unnecessary network calls for
// requests it already knows will be denied. The authorization decision uses
// the same [authz.Engine] as the server-side interceptor.
//
// Note: the server should still perform its own authorization check, as the
// client's decision can be stale.
func UnaryClientInterceptor(eng authzEngine.Engine, opts ...Option) grpc.UnaryClientInterceptor {
	cfg := &options{
		errorFunc: func(_ context.Context, _ error) error {
			return status.Error(codes.PermissionDenied, "permission denied")
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	actionResolver := cfg.clientActionResolver
	if actionResolver == nil {
		actionResolver = defaultClientActionResolver
	}
	resourceResolver := cfg.clientResourceResolver
	if resourceResolver == nil {
		resourceResolver = defaultClientResourceResolver
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

		subject := extractSubject(ctx, cfg)
		action := actionResolver(method)
		resource := resourceResolver(method)
		project := extractProject(ctx, cfg)

		authorized, err := eng.IsAuthorized(ctx,
			authzEngine.Subject(subject),
			authzEngine.Action(action),
			authzEngine.Resource(resource),
			authzEngine.Project(project),
		)
		if err != nil {
			return cfg.errorFunc(ctx, err)
		}
		if !authorized {
			return cfg.errorFunc(ctx, authzEngine.ErrMissingAuthClaims)
		}

		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that performs
// a pre-flight authorization check before creating the stream.
func StreamClientInterceptor(eng authzEngine.Engine, opts ...Option) grpc.StreamClientInterceptor {
	cfg := &options{
		errorFunc: func(_ context.Context, _ error) error {
			return status.Error(codes.PermissionDenied, "permission denied")
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	actionResolver := cfg.clientActionResolver
	if actionResolver == nil {
		actionResolver = defaultClientActionResolver
	}
	resourceResolver := cfg.clientResourceResolver
	if resourceResolver == nil {
		resourceResolver = defaultClientResourceResolver
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

		subject := extractSubject(ctx, cfg)
		action := actionResolver(method)
		resource := resourceResolver(method)
		project := extractProject(ctx, cfg)

		authorized, err := eng.IsAuthorized(ctx,
			authzEngine.Subject(subject),
			authzEngine.Action(action),
			authzEngine.Resource(resource),
			authzEngine.Project(project),
		)
		if err != nil {
			return nil, cfg.errorFunc(ctx, err)
		}
		if !authorized {
			return nil, cfg.errorFunc(ctx, authzEngine.ErrMissingAuthClaims)
		}

		return streamer(ctx, desc, cc, method, callOpts...)
	}
}

// defaultClientActionResolver extracts the method name from FullMethod.
func defaultClientActionResolver(method string) string {
	return defaultActionFromMethod(method)
}

// defaultClientResourceResolver extracts the service name from FullMethod.
func defaultClientResourceResolver(method string) string {
	return defaultResourceFromMethod(method)
}
