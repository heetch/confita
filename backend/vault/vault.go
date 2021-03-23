package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"

	"github.com/heetch/confita/backend"
)

// Backend loads keys from Vault
type Backend struct {
	client *api.Logical
	path   string
	secret *api.Secret
	// KV secrets engine v2
	v2 bool
}

// NewBackend creates a configuration loader that loads from Vault
// all the keys from the given path and holds them in memory.
// Use this when using Vault KV secrets engine v1.
func NewBackend(client *api.Logical, path string) *Backend {
	return &Backend{
		client: client,
		path:   path,
	}
}

// NewBackendV2 creates a configuration loader that loads from Vault
// all the keys from the given path and holds them in memory.
// Use this when using Vault KV secrets engine v2.
func NewBackendV2(client *api.Logical, path string) *Backend {
	path = strings.TrimPrefix(path, "/")
	// The KV secrets engine v2 uses the "secrets/data" prefix in the path,
	// but we want to support regular paths as well, just like the Vault CLI does.
	if strings.HasPrefix(path, "secret/") && !strings.HasPrefix(path, "secret/data/") {
		path = strings.Replace(path, "secret/", "secret/data/", 1)
	}
	return &Backend{
		client: client,
		path:   path,
		v2:     true,
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

	if b.v2 {
		if data, ok := b.secret.Data["data"]; ok {
			data := data.(map[string]interface{})
			if v, ok := data[key]; ok {
				return []byte(v.(string)), nil
			}
		}
	} else {
		if v, ok := b.secret.Data[key]; ok {
			return []byte(v.(string)), nil
		}
	}

	return nil, backend.ErrNotFound
}

// Name returns the name of the backend.
func (b *Backend) Name() string {
	return "vault"
}
