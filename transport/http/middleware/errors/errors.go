// Package errors provides an HTTP middleware and helper functions that encode
// [errors.Error] values into well-formed HTTP responses.
//
// Since standard [http.HandlerFunc] does not return errors, handlers should
// call [Respond] to write error responses. The [Middleware] stores encoding
// configuration (custom response body builder, etc.) in the request context so
// that [Respond] can access it without requiring explicit options on every call.
//
// Usage:
//
//	srv.Use(errmw.Middleware())                      // register once
//
//	// In a handler:
//	func(w http.ResponseWriter, r *http.Request) {
//	    user, err := svc.GetUser(r.Context(), id)
//	    if err != nil {
//	        errmw.Respond(w, err)  // writes 404 + JSON body
//	        return
//	    }
//	    json.NewEncoder(w).Encode(user)
//	}
package errors

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/tx7do/go-wind-plugins/errors"
	httpPlugin "github.com/tx7do/go-wind-plugins/transport/http"
)

// ctxKey is an unexported type for context keys in this package.
type ctxKey struct{}

// Option configures the error encoding middleware.
type Option func(*options)

type options struct {
	// bodyBuilder customises the JSON response body for an error.
	// If nil, [defaultBodyBuilder] is used.
	bodyBuilder func(w http.ResponseWriter, r *http.Request, e *errs.Error) any
}

// WithBodyBuilder sets a custom response body builder.
//
// The returned value is JSON-encoded and written as the response body.
// This allows applications to standardise their error response format.
func WithBodyBuilder(fn func(w http.ResponseWriter, r *http.Request, e *errs.Error) any) Option {
	return func(o *options) { o.bodyBuilder = fn }
}

// Middleware returns a [httpPlugin.Middleware] that stores error-encoding
// configuration in the request context. Handlers can then call [Respond] to
// write error responses using the stored configuration.
func Middleware(opts ...Option) httpPlugin.Middleware {
	cfg := &options{
		bodyBuilder: defaultBodyBuilder,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKey{}, cfg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Respond writes an error as an HTTP response.
//
// If err wraps an [*errors.Error], its Code is used as the HTTP status code and
// a JSON body is written containing the reason and message. Otherwise a generic
// 500 Internal Server Error is returned without leaking the error details.
//
// If the [Middleware] has been registered, the encoding configuration from the
// request context is used. Otherwise package defaults are used.
func Respond(w http.ResponseWriter, r *http.Request, err error) {
	cfg := configFromRequest(r)

	e := errs.FromError(err)
	if e == nil {
		// Unknown error — do not leak details.
		e = errs.New(errs.StatusInternalServerError, "INTERNAL", "internal server error")
	}

	body := cfg.bodyBuilder(w, r, e)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Code)
	_ = json.NewEncoder(w).Encode(body)
}

// configFromRequest extracts the encoding configuration from the request
// context. If the middleware was not registered, default options are used.
func configFromRequest(r *http.Request) *options {
	if r == nil || r.Context() == nil {
		return &options{bodyBuilder: defaultBodyBuilder}
	}
	if v, ok := r.Context().Value(ctxKey{}).(*options); ok {
		return v
	}
	return &options{bodyBuilder: defaultBodyBuilder}
}

// ErrorBody is the default JSON response body for errors.
type ErrorBody struct {
	Code    int               `json:"code"`
	Reason  string            `json:"reason"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// defaultBodyBuilder produces a standard JSON error response body.
func defaultBodyBuilder(_ http.ResponseWriter, _ *http.Request, e *errs.Error) any {
	return ErrorBody{
		Code:    e.Code,
		Reason:  e.Reason,
		Message: e.Message,
		Meta:    e.Metadata,
	}
}
