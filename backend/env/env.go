package env

import (
	"context"
	"os"
	"strings"

	"github.com/heetch/confita/backend"
)

// NewBackend creates a configuration loader that loads from the environment.
// If the key is not found, this backend tries again by turning any kebabcase key to snakecase and
// lowercase letters to uppercase.
func NewBackend() backend.Backend {
	return backend.Func("env", func(ctx context.Context, key string) ([]byte, error) {
		val, ok := os.LookupEnv(key)
		if ok {
			return []byte(val), nil
		}

		key = strings.Replace(strings.ToUpper(key), "-", "_", -1)

		val, ok = os.LookupEnv(key)
		if ok {
			return []byte(val), nil
		}

		return nil, backend.ErrNotFound
	})
}
