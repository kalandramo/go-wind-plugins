package cerbos

import (
	"net/http"
	"time"
)

// Options holds configuration for the Cerbos engine.
type Options struct {
	// endpoint is the Cerbos PDP HTTP endpoint (e.g. "http://localhost:3592").
	endpoint string

	// httpClient allows the caller to inject a custom HTTP client.
	httpClient *http.Client

	// playgroundInstance is the Cerbos Playground instance ID for
	// testing against the hosted playground.
	playgroundInstance string

	// principalRoles maps a subject to Cerbos roles for the request.
	// This is a static mapping; for dynamic roles, use a callback.
	principalRoles map[string][]string

	// authorizer is an optional callback for local evaluation (testing).
	authorizer AuthorizerFunc
}

// AuthorizerFunc is the signature for a local authorization callback.
type AuthorizerFunc func(principal, action, resource string) (bool, error)

type OptFunc func(*Options)

// WithEndpoint sets the Cerbos PDP endpoint URL.
func WithEndpoint(url string) OptFunc {
	return func(o *Options) { o.endpoint = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) OptFunc {
	return func(o *Options) { o.httpClient = c }
}

// WithPlaygroundInstance sets the Cerbos Playground instance ID.
func WithPlaygroundInstance(id string) OptFunc {
	return func(o *Options) { o.playgroundInstance = id }
}

// WithPrincipalRoles sets a static principal→roles mapping.
func WithPrincipalRoles(roles map[string][]string) OptFunc {
	return func(o *Options) { o.principalRoles = roles }
}

// WithAuthorizer sets a local authorization callback for testing.
func WithAuthorizer(fn AuthorizerFunc) OptFunc {
	return func(o *Options) { o.authorizer = fn }
}

func (o *Options) getHTTPClient() *http.Client {
	if o.httpClient != nil {
		return o.httpClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (o *Options) getEndpoint() string {
	if o.endpoint != "" {
		return o.endpoint
	}
	return "http://localhost:3592"
}
