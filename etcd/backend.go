package etcd

import (
	"context"
	"path"

	"github.com/coreos/etcd/clientv3"
	config "github.com/heetch/go-config"
)

// Backend loads keys from etcd.
type Backend struct {
	client *clientv3.Client
	prefix string
}

// NewBackend creates a configuration loader that loads from etcd.
func NewBackend(client *clientv3.Client, prefix string) *Backend {
	return &Backend{
		client: client,
		prefix: prefix,
	}
}

// Get loads the given key from etcd.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	resp, err := b.client.Get(ctx, path.Join(b.prefix, key))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, config.ErrNotFound
	}

	return resp.Kvs[0].Value, nil
}
