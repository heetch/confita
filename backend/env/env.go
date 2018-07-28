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
func NewBackend(fns ...opt) backend.Backend {
	return backend.Func("env", func(ctx context.Context, key string) ([]byte, error) {
		val, ok := os.LookupEnv(opts(key, fns...))
		if ok {
			return []byte(val), nil
		}
		if len(fns) == 0 {
			key = strings.Replace(strings.ToUpper(key), "-", "_", -1)

			val, ok = os.LookupEnv(key)
			if ok {
				return []byte(val), nil
			}
		}
		return nil, backend.ErrNotFound
	})
}

// WithPreffix adds preffix for searching env variable
func WithPreffix(preffix string) opt {
	return func(key string) string {
		return preffix + key
	}
}

// ToUpper uppercases searching env variable
func ToUpper() opt {
	return func(key string) string {
		return strings.Replace(strings.ToUpper(key), "-", "_", -1)
	}
}

type opt func(key string) string

func opts(key string, fns ...opt) string {
	for i := range fns {
		key = fns[i](key)
	}
	return key
}
