// Package authn provides an HTTP middleware that authenticates incoming
// requests via an [engine.Authenticator] and injects the resulting
// [engine.AuthClaims] into the request context for downstream handlers.
//
// Usage:
//
//	import (
//	    httpPlugin "github.com/tx7do/go-wind-plugins/transport/http"
//	    "github.com/tx7do/go-wind-plugins/transport/http/middleware/authn"
//	    jwtAuthn "github.com/tx7do/go-wind-plugins/security/authn/jwt"
//	)
//
//	authenticator, _ := jwtAuthn.NewAuthenticator(jwtAuthn.WithKey(secret))
//	srv := httpPlugin.NewServer(":8080")
//	srv.Use(authn.Middleware(authenticator))
//	srv.GET("/api/data", myHandler)
//
// Inside the handler, claims are retrieved via:
//
//	claims, ok := engine.AuthClaimsFromContext(r.Context())
package authn

import (
	"net/http"

	"google.golang.org/grpc/metadata"

	engine "github.com/tx7do/go-wind-plugins/security/authn"
	httpPlugin "github.com/tx7do/go-wind-plugins/transport/http"
)

// Option configures the authn middleware.
type Option func(*options)

type options struct {
	// headerName is the HTTP header that carries the token. Defaults to
	// "Authorization" (engine.HeaderAuthorize).
	headerName string

	// errorHandle lets the caller customise the 401 response body. If nil,
	// a plain-text "unauthorized" is written.
	errorHandle func(http.ResponseWriter, *http.Request, error)
}

// WithHeaderName overrides the HTTP header used to extract the token.
func WithHeaderName(name string) Option {
	return func(o *options) { o.headerName = name }
}

// WithErrorHandle sets a custom function to handle authentication errors.
// This allows callers to write JSON error bodies, set specific headers, etc.
func WithErrorHandle(fn func(http.ResponseWriter, *http.Request, error)) Option {
	return func(o *options) { o.errorHandle = fn }
}

// Middleware returns an [httpPlugin.Middleware] that authenticates every
// incoming request using the provided [engine.Authenticator].
//
// On success the resulting [*engine.AuthClaims] are injected into the
// request context via [engine.ContextWithAuthClaims]. Downstream handlers
// retrieve them with [engine.AuthClaimsFromContext].
//
// On failure the middleware writes a 401 response and does NOT call next.
func Middleware(auth engine.Authenticator, opts ...Option) httpPlugin.Middleware {
	cfg := &options{
		headerName: engine.HeaderAuthorize,
		errorHandle: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the HTTP header.
			token := r.Header.Get(cfg.headerName)
			if token == "" {
				cfg.errorHandle(w, r, engine.ErrMissingBearerToken)
				return
			}

			// Build a gRPC-style incoming metadata context so the
			// authenticator's AuthFromMD can find the token.
			md := metadata.Pairs(engine.HeaderAuthorize, token)
			ctx := metadata.NewIncomingContext(r.Context(), md)

			// Authenticate.
			claims, err := auth.Authenticate(ctx)
			if err != nil {
				cfg.errorHandle(w, r, err)
				return
			}

			// Inject claims into the context for downstream handlers.
			ctx = engine.ContextWithAuthClaims(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
