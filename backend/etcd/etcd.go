package etcd

import (
	"context"
	"path"
	"strings"

	"github.com/heetch/confita/backend"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Backend loads keys from etcd.
type Backend struct {
	client   *clientv3.Client
	prefix   string
	prefetch bool
	cache    map[string][]byte
}

// NewBackend creates a configuration loader that loads from etcd.
func NewBackend(client *clientv3.Client, opts ...Option) *Backend {
	b := Backend{
		client: client,
	}

	for _, opt := range opts {
		opt(&b)
	}

	return &b
}

// Get loads the given key from etcd.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	if b.cache == nil && b.prefetch {
		err := b.fetchTree(ctx)
		if err != nil {
			return nil, err
		}
	}

	if b.cache != nil {
		return b.fromCache(ctx, key)
	}

	resp, err := b.client.Get(ctx, path.Join(b.prefix, key))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, backend.ErrNotFound
	}

	return resp.Kvs[0].Value, nil
}

func (b *Backend) fetchTree(ctx context.Context) error {
	resp, err := b.client.KV.Get(ctx, b.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	b.cache = make(map[string][]byte)

	for _, kv := range resp.Kvs {
		b.cache[strings.TrimLeft(string(kv.Key), b.prefix+"/")] = kv.Value
	}

	return nil
}

func (b *Backend) fromCache(_ context.Context, key string) ([]byte, error) {
	v, ok := b.cache[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return v, nil
}

// Name returns the name of the backend.
func (b *Backend) Name() string {
	return "etcd"
}

// Option is used to configure the Consul backend.
type Option func(*Backend)

// WithPrefix is used to specify the prefix to prepend on every keys.
func WithPrefix(prefix string) Option {
	return func(b *Backend) {
		b.prefix = prefix
	}
}

// WithPrefetch is used to prefetch the entire tree and cache it to
// avoid further round trips. If the WithPrefix option is used, will fetch
// the tree under the specified prefix.
func WithPrefetch() Option {
	return func(b *Backend) {
		b.prefetch = true
	}
}
