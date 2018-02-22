package vault

import (
	"context"

	"github.com/hashicorp/vault/api"
	"github.com/heetch/confita/backend"
)

type apiVault interface {
	Read(p string) (*api.Secret, error)
}

// Backend loads keys from Vault
type Backend struct {
	client apiVault
	path   string
}

// NewBackend creates a configuration loader that loads from Vault
func NewBackend(c *api.Logical, p string) *Backend {
	return &Backend{
		client: c,
		path:   p,
	}
}

// Get loads the given key from Vault
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	secret, err := b.client.Read(b.path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, backend.ErrNotFound
	}

	if v, ok := secret.Data[key]; ok {
		return []byte(v.(string)), nil
	}

	return nil, backend.ErrNotFound
}
