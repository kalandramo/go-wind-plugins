package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

var _ baseConfig.Reader = (*source)(nil)

type source struct {
	options *options
}

// New creates a config source backed by an [io/fs.FS] (typically an
// [embed.FS]). The fsys option is required.
func New(opts ...Option) (*source, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.fsys == nil {
		return nil, errors.New("fsys invalid: an io/fs.FS must be provided")
	}

	return &source{options: o}, nil
}

// resolveKey returns the key to use for the given caller-provided key.
// If key is empty the configured default path is used.
func (s *source) resolveKey(key string) string {
	if key != "" {
		return key
	}
	return s.options.path
}

// Load implements [baseConfig.Reader].
// It reads the file at key (or the configured default path when key is empty)
// from the embedded file system and returns its raw contents.
func (s *source) Load(_ context.Context, key string) ([]byte, error) {
	path := s.resolveKey(key)
	if path == "" {
		return nil, errors.New("no file path specified")
	}

	data, err := fs.ReadFile(s.options.fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read embedded file %s: %w", path, err)
	}
	return data, nil
}
