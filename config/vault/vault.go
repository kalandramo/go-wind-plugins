package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

const defaultPollInterval = 30 * time.Second

var (
	_ baseConfig.Reader       = (*source)(nil)
	_ baseConfig.ValueWatcher = (*source)(nil)
)

type source struct {
	client  *vaultapi.Client
	options *options
}

// New creates a HashiCorp Vault-backed config source.
// The client and path options are required.
func New(client *vaultapi.Client, opts ...Option) (*source, error) {
	if client == nil {
		return nil, errors.New("vault client is nil")
	}

	o := &options{
		ctx:          context.Background(),
		path:         "",
		dataKey:      "content",
		pollInterval: defaultPollInterval,
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
// It reads the secret at the given path from Vault and extracts the
// configured dataKey field (default "content").
func (s *source) Load(ctx context.Context, key string) ([]byte, error) {
	path := s.resolveKey(key)

	secret, err := s.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("vault: read %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	return extractValue(secret.Data, s.options.dataKey), nil
}

// WatchValue implements [baseConfig.ValueWatcher].
// Vault has no native push mechanism, so this uses periodic polling.
// On each poll interval it re-reads the secret and pushes the value
// when it changes. The initial value is pushed immediately.
func (s *source) WatchValue(ctx context.Context, key string) (<-chan []byte, error) {
	path := s.resolveKey(key)

	out := make(chan []byte, 1)

	go func() {
		defer close(out)

		ticker := time.NewTicker(s.options.pollInterval)
		defer ticker.Stop()

		var lastValue []byte

		// Initial load.
		if secret, err := s.client.Logical().ReadWithContext(ctx, path); err == nil && secret != nil && secret.Data != nil {
			val := extractValue(secret.Data, s.options.dataKey)
			lastValue = val
			select {
			case out <- val:
			case <-ctx.Done():
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				secret, err := s.client.Logical().ReadWithContext(ctx, path)
				if err != nil || secret == nil || secret.Data == nil {
					continue
				}
				val := extractValue(secret.Data, s.options.dataKey)
				if !bytesEqual(val, lastValue) {
					lastValue = val
					select {
					case out <- val:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out, nil
}

// extractValue retrieves the raw config bytes from a Vault secret's Data map.
// If dataKey is "content" and the Data contains a "content" key, use it.
// For KV v2 secrets, Data is wrapped as map[string]interface{}{"data": ..., "metadata": ...}.
// For KV v1, Data is the flat key-value map directly.
func extractValue(data map[string]interface{}, dataKey string) []byte {
	// Handle KV v2: unwrap "data" if present.
	if inner, ok := data["data"]; ok {
		if innerMap, ok := inner.(map[string]interface{}); ok {
			data = innerMap
		}
	}

	// Look for the configured dataKey.
	if val, ok := data[dataKey]; ok {
		switch v := val.(type) {
		case string:
			return []byte(v)
		case []byte:
			return v
		default:
			// JSON-encode complex values.
			if b, err := json.Marshal(v); err == nil {
				return b
			}
		}
	}

	// Fallback: if dataKey not found, JSON-encode the whole data map.
	if b, err := json.Marshal(data); err == nil {
		return b
	}
	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
