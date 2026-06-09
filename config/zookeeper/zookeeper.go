package zookeeper

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-zookeeper/zk"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

var (
	_ baseConfig.Reader       = (*source)(nil)
	_ baseConfig.ValueWatcher = (*source)(nil)
)

type source struct {
	client  *zk.Conn
	options *options
}

// New creates a ZooKeeper-backed config source.
// The client and path options are required.
func New(client *zk.Conn, opts ...Option) (*source, error) {
	if client == nil {
		return nil, errors.New("zookeeper client is nil")
	}

	o := &options{
		ctx:  context.Background(),
		path: "",
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.path == "" {
		return nil, errors.New("path invalid")
	}

	return &source{
		client:  client,
		options: o,
	}, nil
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
// It returns the raw value stored at the given znode path.
func (s *source) Load(_ context.Context, key string) ([]byte, error) {
	path := s.resolveKey(key)

	data, _, err := s.client.Get(path)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return nil, nil
		}
		return nil, fmt.Errorf("zookeeper: get %s: %w", path, err)
	}
	return data, nil
}

// WatchValue implements [baseConfig.ValueWatcher].
// It uses ExistsW / GetW to receive znode change events and pushes
// the latest value on each event. The channel is closed when ctx is
// cancelled or the watch loop exits.
func (s *source) WatchValue(ctx context.Context, key string) (<-chan []byte, error) {
	path := s.resolveKey(key)

	out := make(chan []byte, 1)

	go func() {
		defer close(out)

		for {
			// GetW returns data + a channel that fires on the next change.
			data, _, eventCh, err := s.client.GetW(path)
			if err != nil {
				if errors.Is(err, zk.ErrNoNode) {
					// Node doesn't exist yet — use ExistsW to wait for creation.
					_, _, ev, existsErr := s.client.ExistsW(path)
					if existsErr != nil {
						return
					}
					select {
					case e := <-ev:
						if e.Type == zk.EventNodeCreated || e.Type == zk.EventNodeDataChanged {
							continue // re-loop to GetW the new value
						}
						return
					case <-ctx.Done():
						return
					}
				}
				return
			}

			// Push initial value.
			select {
			case out <- data:
			case <-ctx.Done():
				return
			}

			// Wait for next change event.
			select {
			case e := <-eventCh:
				if e.Type == zk.EventNodeDeleted {
					return
				}
				if e.Type == zk.EventNodeDataChanged || e.Type == zk.EventNodeCreated {
					continue // re-loop to get updated value
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
