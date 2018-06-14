package flags

import (
	"context"
	"flag"

	"github.com/heetch/confita"
	"github.com/pkg/errors"
)

// Backend that loads configuration from the command line flags.
type Backend struct{}

// NewBackend creates a flags backend.
func NewBackend() *Backend {
	return new(Backend)
}

// LoadStruct takes a struct config, define flags based on it and parse the command line args,
func (b *Backend) LoadStruct(ctx context.Context, cfg *confita.StructConfig) error {
	for _, f := range cfg.Fields {
		if f.Backend != "" {
			continue
		}

		flag.Var(&flagValue{f}, f.Key, "")
	}

	flag.Parse()

	return nil
}

type flagValue struct {
	*confita.FieldConfig
}

func (f *flagValue) String() string {
	return f.Key
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the type of the file.
func (b *Backend) Name() string {
	return "flags"
}
