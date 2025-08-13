package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/heetch/confita/backend"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Backend that loads a configuration from a file.
// It supports json and yaml formats.
type Backend struct {
	path     string
	name     string
	optional bool
}

// NewBackend creates a configuration loader that loads from a file.
// The content will get decoded based on the file extension.
// If optional parameter is set to true, calling Unmarshal won't return an error if the file doesn't exist.
func NewBackend(path string) *Backend {
	name := filepath.Ext(path)
	if name != "" {
		name = name[1:]
	}

	return &Backend{
		path: path,
		name: name,
	}
}

// NewOptionalBackend implementation is exactly the same as NewBackend except that
// if the file is not found, backend.ErrNotFound will be returned.
func NewOptionalBackend(path string) *Backend {
	name := filepath.Ext(path)
	if name != "" {
		name = name[1:]
	}

	return &Backend{
		path:     path,
		name:     name,
		optional: true,
	}
}

// Unmarshal takes a struct pointer and unmarshals the file into it,
// using either json or yaml based on the file extention.
func (b *Backend) Unmarshal(ctx context.Context, to any) error {
	f, err := os.Open(b.path)
	if err != nil {
		if b.optional {
			return backend.ErrNotFound
		}
		return errors.Wrapf(err, "failed to open file at path \"%s\"", b.path)
	}
	defer f.Close()

	switch ext := filepath.Ext(b.path); ext {
	case ".json":
		err = json.NewDecoder(f).Decode(to)
	case ".yml":
		fallthrough
	case ".yaml":
		err = yaml.NewDecoder(f).Decode(to)
	case ".toml":
		_, err = toml.DecodeReader(f, to)
	default:
		err = errors.Errorf("unsupported extension \"%s\"", ext)
	}

	return errors.Wrapf(err, "failed to decode file \"%s\"", b.path)
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the type of the file.
func (b *Backend) Name() string {
	return b.name
}
