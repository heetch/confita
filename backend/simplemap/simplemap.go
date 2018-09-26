package simplemap

import (
	"github.com/heetch/confita"
	"github.com/pkg/errors"
	"reflect"

	"context"
)

// Backend that loads config from map
type Backend struct {
	theMap map[string]interface{}
}

// NewBackend creates a simplemap backend.
func NewBackend(theMap map[string]interface{}) *Backend {
	return &Backend{
		theMap: theMap,
	}
}

// LoadStruct takes a struct config and loads the map into it
func (b *Backend) LoadStruct(ctx context.Context, cfg *confita.StructConfig) error {
	for _, f := range cfg.Fields {
		mapVal := b.theMap[f.Key]
		if mapVal == nil {
			continue
		}
		mapRef := reflect.ValueOf(mapVal)
		f.Value.Set(mapRef)
	}
	return nil
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the name of the flags backend.
func (b *Backend) Name() string {
	return "simplemap"
}
