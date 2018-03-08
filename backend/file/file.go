package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/heetch/confita/backend"
	"github.com/pkg/errors"
)

// Backend that loads a configuration from a file.
// It supports json and yaml formats.
type Backend struct {
	path        string
	unmarshaler backend.KeyUnmarshaler
}

// NewBackend creates a configuration loader that loads from a file.
// The content will get decoded based on the file extension and cached in the backend.
func NewBackend(path string) *Backend {
	return &Backend{
		path: path,
	}
}

func (b *Backend) loadFile() error {
	f, err := os.Open(b.path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file at path \"%s\"", b.path)
	}
	defer f.Close()

	switch ext := filepath.Ext(b.path); ext {
	case ".json":
		var j jsonConfig
		err = json.NewDecoder(f).Decode(&j.content)
		b.unmarshaler = &j
	default:
		err = errors.Errorf("unsupported extension \"%s\"", ext)
	}

	return errors.Wrapf(err, "failed to decode file \"%s\"", b.path)
}

// UnmarshalKey unmarshals the given key directly to the given target.
// It returns an error if the underlying file cannot be loaded.
func (b *Backend) UnmarshalKey(ctx context.Context, key string, to interface{}) error {
	if b.unmarshaler == nil {
		err := b.loadFile()
		if err != nil {
			return err
		}
	}

	err := b.unmarshaler.UnmarshalKey(ctx, key, to)
	if err == backend.ErrNotFound {
		return err
	}

	return errors.Wrapf(err, "failed to unmarshal key \"%s\"")
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

type jsonConfig struct {
	content map[string]json.RawMessage
}

func (j *jsonConfig) UnmarshalKey(_ context.Context, key string, to interface{}) error {
	v, ok := j.content[key]
	if !ok {
		return backend.ErrNotFound
	}

	return json.Unmarshal(v, to)
}