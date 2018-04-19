package consul

import (
	"context"
	"path"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/heetch/confita/backend"
)

// Backend loads keys from Consul.
type Backend struct {
	client   *api.Client
	prefix   string
	prefetch bool
	cache    map[string][]byte
}

// NewBackend creates a configuration loader that loads from Consul.
func NewBackend(client *api.Client, opts ...Option) *Backend {
	b := Backend{
		client: client,
	}

	for _, opt := range opts {
		opt(&b)
	}

	return &b
}

// Get loads the given key from Consul.
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

func (b *Backend) fetchTree(ctx context.Context) error {
	var opt api.QueryOptions

	list, _, err := b.client.KV().List(b.prefix, opt.WithContext(ctx))
	if err != nil {
		return err
	}

	b.cache = make(map[string][]byte)

	for _, kv := range list {
		b.cache[strings.TrimLeft(kv.Key, b.prefix+"/")] = kv.Value
	}

	return nil
}

func (b *Backend) fromCache(ctx context.Context, key string) ([]byte, error) {
	v, ok := b.cache[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return v, nil
}

// Name returns the name of the backend.
func (b *Backend) Name() string {
	return "consul"
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
