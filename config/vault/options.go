package vault

import (
	"context"
	"time"
)

// Option is vault config option.
type Option func(o *options)

type options struct {
	ctx          context.Context
	path         string
	dataKey      string
	pollInterval time.Duration
}

// WithContext with vault config context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithPath is the Vault secret path (e.g. "secret/data/myapp/config").
func WithPath(p string) Option {
	return func(o *options) {
		o.path = p
	}
}

// WithDataKey sets the key inside the Vault secret's Data map that holds
// the raw config value. Default is "content".
// For complex secret structures, set this to the field that contains
// your config payload (e.g. "config", "data", "value").
func WithDataKey(k string) Option {
	return func(o *options) {
		o.dataKey = k
	}
}

// WithPollInterval sets the interval for polling Vault for changes.
// Default is 30 seconds. Vault does not support push-based notifications.
func WithPollInterval(d time.Duration) Option {
	return func(o *options) {
		o.pollInterval = d
	}
}
