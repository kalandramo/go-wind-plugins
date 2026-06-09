package zookeeper

import "context"

// Option is zookeeper config option.
type Option func(o *options)

type options struct {
	ctx  context.Context
	path string
}

// WithContext with zookeeper config context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithPath is config znode path.
func WithPath(p string) Option {
	return func(o *options) {
		o.path = p
	}
}
