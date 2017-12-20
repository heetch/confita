package etcd

import (
	"context"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	config "github.com/heetch/go-config"
)

// Backend loads keys from etcd.
type Backend struct {
	client  *clientv3.Client
	prefix  string
	timeout time.Duration
}

// Get loads the given key from etcd.
func (b *Backend) Get(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	resp, err := b.client.Get(ctx, path.Join(b.prefix, key))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, config.ErrNotFound
	}

	return resp.Kvs[0].Value, nil
}

// FromEtcd creates a configuration loader that loads from etcd.
func FromEtcd(client *clientv3.Client, prefix string, timeout time.Duration) *Backend {
	return &Backend{
		client:  client,
		prefix:  prefix,
		timeout: timeout,
	}
}
