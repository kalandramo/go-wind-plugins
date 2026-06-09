package fs

import "io/fs"

// Option is fs config option.
type Option func(o *options)

type options struct {
	fsys fs.FS
	path string
}

// WithFS sets the [io/fs.FS] to read configuration from.
// Typically you pass an [embed.FS]:
//
//	//go:embed configs/*
//	var configFS embed.FS
//	src, _ := fs.New(fs.WithFS(configFS), fs.WithPath("configs/app.yaml"))
func WithFS(fsys fs.FS) Option {
	return func(o *options) {
		o.fsys = fsys
	}
}

// WithPath sets the default file path within the embedded file system.
// This path is used when Load is called with an empty key.
func WithPath(p string) Option {
	return func(o *options) {
		o.path = p
	}
}
