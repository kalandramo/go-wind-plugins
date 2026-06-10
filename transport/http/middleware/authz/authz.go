// Package authz provides an HTTP middleware that enforces authorization
// decisions using an [engine.Engine] (e.g. casbin, opa, acl, rbac).
//
// The middleware expects authentication claims to already be in the request
// context — typically injected by the authn middleware. It extracts the
// subject from the authn claims, then calls the authz engine's
// [engine.IsAuthorized] to decide whether to allow or deny the request.
//
// Usage:
//
//	import (
//	    httpPlugin "github.com/tx7do/go-wind-plugins/transport/http"
//	    "github.com/tx7do/go-wind-plugins/transport/http/middleware/authn"
//	    authzMw "github.com/tx7do/go-wind-plugins/transport/http/middleware/authz"
//	    authnEngine "github.com/tx7do/go-wind-plugins/security/authn"
//	    authzEngine "github.com/tx7do/go-wind-plugins/security/authz"
//	    "github.com/tx7do/go-wind-plugins/security/authz/acl"
//	)
//
//	authzEng, _ := acl.NewEngine(ctx,
//	    acl.WithRule("alice", "read", "doc:*"),
//	)
//	srv.Use(authn.Middleware(authenticator))
//	srv.Use(authzMw.Middleware(authzEng,
//	    authzMw.WithActionResolver(func(r *http.Request) string {
//	        switch r.Method {
//	        case "GET": return "read"
//	        case "POST", "PUT", "PATCH": return "write"
//	        case "DELETE": return "delete"
//	        }
//	        return ""
//	    }),
//	    authzMw.WithResourceResolver(func(r *http.Request) string {
//	        return r.URL.Path  // e.g. "/api/docs/1"
//	    }),
//	))
//
// On failure the middleware writes a 403 Forbidden response and does NOT
// call next.
package authz

import (
	"net/http"

	authnEngine "github.com/tx7do/go-wind-plugins/security/authn"
	authzEngine "github.com/tx7do/go-wind-plugins/security/authz"
	httpPlugin "github.com/tx7do/go-wind-plugins/transport/http"
)

// ResolverFunc extracts the action or resource string from an HTTP request.
// This allows callers to customize how URLs map to authorization resources
// and how HTTP methods map to actions.
type ResolverFunc func(r *http.Request) string

// Option configures the authz middleware.
type Option func(*options)

type options struct {
	// actionResolver maps an HTTP request to an authorization action.
	// If nil, the HTTP method is used directly as the action.
	actionResolver ResolverFunc

	// resourceResolver maps an HTTP request to an authorization resource.
	// If nil, the URL path is used directly as the resource.
	resourceResolver ResolverFunc

	// subjectResolver extracts the subject from the request context.
	// If nil, the authn claims subject is used.
	// This allows custom subject extraction (e.g. from headers or query params).
	subjectResolver func(r *http.Request) string

	// projectResolver extracts the project from the request.
	// If nil, no project is passed (empty string).
	projectResolver ResolverFunc

	// errorHandle lets the caller customise the 403 response body.
	// If nil, a plain-text "forbidden" is written.
	errorHandle func(http.ResponseWriter, *http.Request, error)

	// skipFunc returns true if the middleware should skip authorization
	// for this request (e.g. health checks, public endpoints).
	skipFunc func(r *http.Request) bool
}

// WithActionResolver sets a custom function to derive the authorization
// action from the HTTP request.
func WithActionResolver(fn ResolverFunc) Option {
	return func(o *options) { o.actionResolver = fn }
}

// WithResourceResolver sets a custom function to derive the authorization
// resource from the HTTP request.
func WithResourceResolver(fn ResolverFunc) Option {
	return func(o *options) { o.resourceResolver = fn }
}

// WithSubjectResolver sets a custom function to extract the subject.
func WithSubjectResolver(fn func(r *http.Request) string) Option {
	return func(o *options) { o.subjectResolver = fn }
}

// WithProjectResolver sets a custom function to derive the project.
func WithProjectResolver(fn ResolverFunc) Option {
	return func(o *options) { o.projectResolver = fn }
}

// WithErrorHandle sets a custom function to handle authorization errors.
func WithErrorHandle(fn func(http.ResponseWriter, *http.Request, error)) Option {
	return func(o *options) { o.errorHandle = fn }
}

// WithSkipFunc sets a function that returns true to skip authorization
// for certain requests.
func WithSkipFunc(fn func(r *http.Request) bool) Option {
	return func(o *options) { o.skipFunc = fn }
}

// Middleware returns an [httpPlugin.Middleware] that enforces authorization
// on every incoming request.
//
// It should be placed AFTER the authn middleware in the chain, as it
// relies on authn claims being in the request context.
func Middleware(eng authzEngine.Engine, opts ...Option) httpPlugin.Middleware {
	cfg := &options{
		actionResolver:   defaultActionResolver,
		resourceResolver: defaultResourceResolver,
		errorHandle: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "forbidden", http.StatusForbidden)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if needed (e.g. health checks).
			if cfg.skipFunc != nil && cfg.skipFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract subject from authn claims or custom resolver.
			subject := ""
			if cfg.subjectResolver != nil {
				subject = cfg.subjectResolver(r)
			} else {
				// Default: extract from authn claims in context.
				claims, ok := authnEngine.AuthClaimsFromContext(r.Context())
				if ok {
					sub, _ := claims.GetSubject()
					subject = sub
				}
			}

			action := cfg.actionResolver(r)
			resource := cfg.resourceResolver(r)
			project := ""
			if cfg.projectResolver != nil {
				project = cfg.projectResolver(r)
			}

			// Call the authz engine.
			authorized, err := eng.IsAuthorized(r.Context(),
				authzEngine.Subject(subject),
				authzEngine.Action(action),
				authzEngine.Resource(resource),
				authzEngine.Project(project),
			)
			if err != nil {
				cfg.errorHandle(w, r, err)
				return
			}
			if !authorized {
				cfg.errorHandle(w, r, authzEngine.ErrMissingAuthClaims)
				return
			}

			// Inject authz info into context for downstream handlers.
			ctx := authzEngine.ContextWithAuthClaims(r.Context(), &authzEngine.AuthClaims{
				Subject:  (*authzEngine.Subject)(&subject),
				Action:   (*authzEngine.Action)(&action),
				Resource: (*authzEngine.Resource)(&resource),
				Project:  (*authzEngine.Project)(&project),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// defaultActionResolver maps HTTP methods to default authorization actions.
func defaultActionResolver(r *http.Request) string {
	return r.Method
}

// defaultResourceResolver uses the URL path as the resource.
func defaultResourceResolver(r *http.Request) string {
	return r.URL.Path
}
