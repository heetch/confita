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
		key = strings.Replace(key, "-", "_", -1)
		val, ok := os.LookupEnv(opts(key, fns...))
		if ok {
			return []byte(val), nil
		}
		val, ok = os.LookupEnv(opts(strings.ToUpper(key), fns...))
		if ok {
			return []byte(val), nil
		}
		return nil, backend.ErrNotFound
	})
}

// WithPrefix adds preffix for searching env variable
func WithPrefix(preffix string) opt {
	if !strings.HasSuffix(preffix, "_") {
		preffix = preffix + "_"
	}
	return func(key string) string {
		return preffix + key
	}
}

// ToUpper uppercases searching env variable
func ToUpper() opt {
	return func(key string) string {
		return strings.ToUpper(key)
	}
}

type opt func(key string) string

func opts(key string, fns ...opt) string {
	for i := range fns {
		key = fns[i](key)
	}
	return key
}
