package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/heetch/confita/backend"
)

// Backend loads keys from Vault
type Backend struct {
	client *api.Logical
	path   string
	secret *api.Secret
}

// NewBackend creates a configuration loader that loads from Vault
// all the keys from the given path and holds them in memory.
func NewBackend(client *api.Logical, path string) *Backend {
	return &Backend{
		client: client,
		path:   path,
	}
}

// Get loads the given key from Vault.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	var err error

	if b.secret == nil {
		b.secret, err = b.client.Read(b.path)
		if err != nil {
			return nil, err
		}

		if b.secret == nil {
			return nil, fmt.Errorf("secret not found at the following path: %s", b.path)
		}
	}

	if v, ok := b.secret.Data[key]; ok {
		return []byte(v.(string)), nil
	}

	return nil, backend.ErrNotFound
}
