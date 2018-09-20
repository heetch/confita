package simplemap

import (
	"context"

	"github.com/heetch/confita/backend"
)

func NewBackend(configMap map[string]string) backend.Backend {
	return backend.Func("map", func(ctx context.Context, key string) ([]byte, error) {
		return nil, backend.ErrNotFound
	})
}
