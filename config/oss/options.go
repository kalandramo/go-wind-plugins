package oss

import (
	"context"
	"time"
)

// Option is oss config option.
type Option func(o *options)

type options struct {
	ctx          context.Context
	bucket       string
	key          string
	pollInterval time.Duration
}

// WithContext with oss config context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithBucket sets the S3 bucket name (required).
func WithBucket(b string) Option {
	return func(o *options) {
		o.bucket = b
	}
}

// WithKey sets the default object key within the bucket.
// This key is used when Load/WatchValue is called with an empty key.
func WithKey(k string) Option {
	return func(o *options) {
		o.key = k
	}
}

// WithPollInterval sets the interval for polling-based change detection in
// WatchValue. S3-compatible storage does not support push notifications,
// so WatchValue polls HeadObject and compares ETags.
// Default 30 seconds.
func WithPollInterval(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.pollInterval = d
		}
	}
}
