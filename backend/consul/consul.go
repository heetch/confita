package consul

import (
	"context"
	"path"

	"github.com/hashicorp/consul/api"
	"github.com/heetch/confita/backend"
)

// Backend loads keys from Consul.
type Backend struct {
	client *api.Client
	prefix string
}

// NewBackend creates a configuration loader that loads from Consul.
func NewBackend(client *api.Client, prefix string) *Backend {
	return &Backend{
		client: client,
		prefix: prefix,
	}
}

// Get loads the given key from Consul.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	var opt api.QueryOptions

	kv, _, err := b.client.KV().Get(path.Join(b.prefix, key), opt.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if kv == nil {
		return nil, backend.ErrNotFound
	}

	return kv.Value, nil
}
