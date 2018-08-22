package simplemap

import (
	"time"
	"fmt"
	"strconv"
	"github.com/heetch/confita/backend"
	"context"
)

// Backend that loads config from map
type Backend struct{
	theMap	map[string]interface{}
}

// NewBackend creates a simplemap backend.
func NewBackend(theMap map[string]interface{}) *Backend {
	return &Backend{
		theMap: theMap,
	}
}

// Get loads the given key from the map
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	val := b.theMap[key]
	if val == nil {
		return nil, backend.ErrNotFound
	}

	if str, ok := val.(string); ok {
		return []byte(str), nil
	}

	if u, ok := val.(uint); ok {
		return []byte(strconv.FormatUint(uint64(u), 10)), nil
	}

	if i, ok := val.(int); ok {
		return []byte(strconv.FormatInt(int64(i), 10)), nil
	}

	if f, ok := val.(float64); ok {
		return []byte(fmt.Sprintf("%g", f)), nil
	}

	if d, ok := val.(time.Duration); ok {
		return []byte(fmt.Sprintf("%s", d)), nil
	}

	return nil, backend.ErrNotFound
}

// Name returns the name of the flags backend.
func (b *Backend) Name() string {
	return "simplemap"
}