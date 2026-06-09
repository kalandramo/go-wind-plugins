package http

import (
	"context"
	"net/http"
	"time"
)

// Option is http config option.
type Option func(o *options)

type options struct {
	ctx          context.Context
	url          string
	method       string
	headers      map[string]string
	httpClient   *http.Client
	pollInterval time.Duration
}

// WithContext with http config context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithURL sets the endpoint URL to fetch configuration from.
func WithURL(u string) Option {
	return func(o *options) {
		o.url = u
	}
}

// WithMethod sets the HTTP method (default GET).
func WithMethod(m string) Option {
	return func(o *options) {
		o.method = m
	}
}

// WithHeader adds a custom HTTP header to every request.
func WithHeader(key, value string) Option {
	return func(o *options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

// WithHTTPClient sets a custom *http.Client (e.g. with TLS config, proxy, etc.).
// If not set, a default client with a 10-second timeout is used.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) {
		o.httpClient = c
	}
}

// WithPollInterval sets the interval for polling-based change detection in
// WatchValue. Uses HTTP conditional GET (ETag / If-None-Match) to minimise
// bandwidth — the server responds with 304 when the content is unchanged.
// Default 30 seconds.
func WithPollInterval(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.pollInterval = d
		}
	}
}
