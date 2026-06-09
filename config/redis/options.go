package redis

import "context"

// Option is redis config option.
type Option func(o *options)

type options struct {
	ctx  context.Context
	path string
}

// WithContext with redis config context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithPath is config key path in Redis.
func WithPath(p string) Option {
	return func(o *options) {
		o.path = p
	}
}
