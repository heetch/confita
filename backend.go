package confita

import (
	"context"
	"os"
	"strings"
)

// A Backend is used to fetch values from a given key.
type Backend interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// BackendFunc creates a Backend from a function.
func BackendFunc(fn func(context.Context, string) ([]byte, error)) Backend {
	return &backendFunc{fn: fn}
}

type backendFunc struct {
	fn func(context.Context, string) ([]byte, error)
}

func (b *backendFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return b.fn(ctx, key)
}

// EnvBackend creates a configuration loader that loads from the environment.
// If the key is not found, this backend tries again by turning any kebabcase key to snakecase and
// lowercase letters to uppercase.
func EnvBackend() Backend {
	return BackendFunc(func(ctx context.Context, key string) ([]byte, error) {
		val, ok := os.LookupEnv(key)
		if ok {
			return []byte(val), nil
		}

		key = strings.Replace(strings.ToUpper(key), "-", "_", -1)

		val, ok = os.LookupEnv(key)
		if ok {
			return []byte(val), nil
		}

		return nil, ErrNotFound
	})
}
