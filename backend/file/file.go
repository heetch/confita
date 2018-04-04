package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/go-yaml/yaml"
	"github.com/heetch/confita/backend"
	"github.com/pkg/errors"
)

// Backend that loads a configuration from a file.
// It supports json and yaml formats.
type Backend struct {
	path        string
	unmarshaler backend.ValueUnmarshaler
	name        string
}

// NewBackend creates a configuration loader that loads from a file.
// The content will get decoded based on the file extension and cached in the backend.
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

func (b *Backend) loadFile() error {
	f, err := os.Open(b.path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file at path \"%s\"", b.path)
	}
	defer f.Close()

	switch ext := filepath.Ext(b.path); ext {
	case ".json":
		var j jsonConfig
		err = json.NewDecoder(f).Decode(&j)
		b.unmarshaler = &j
	case ".yml":
		fallthrough
	case ".yaml":
		var y yamlConfig
		err = yaml.NewDecoder(f).Decode(&y)
		b.unmarshaler = &y
	case ".toml":
		var t tomlConfig
		_, err = toml.DecodeReader(f, &t)
		b.unmarshaler = &t
	default:
		err = errors.Errorf("unsupported extension \"%s\"", ext)
	}

	return errors.Wrapf(err, "failed to decode file \"%s\"", b.path)
}

// UnmarshalValue unmarshals the given key directly to the given target.
// It returns an error if the underlying file cannot be loaded.
func (b *Backend) UnmarshalValue(ctx context.Context, key string, to interface{}) error {
	if b.unmarshaler == nil {
		err := b.loadFile()
		if err != nil {
			return err
		}
	}

	err := b.unmarshaler.UnmarshalValue(ctx, key, to)
	if err == backend.ErrNotFound {
		return err
	}

	return errors.Wrapf(err, "failed to unmarshal key \"%s\"")
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the type of the file.
func (b *Backend) Name() string {
	return b.name
}

type jsonConfig map[string]json.RawMessage

func (j jsonConfig) UnmarshalValue(_ context.Context, key string, to interface{}) error {
	v, ok := j[key]
	if !ok {
		return backend.ErrNotFound
	}

	return json.Unmarshal(v, to)
}

type yamlConfig map[string]yamlRawMessage

func (y yamlConfig) UnmarshalValue(_ context.Context, key string, to interface{}) error {
	v, ok := y[key]
	if !ok {
		return backend.ErrNotFound
	}

	return v.unmarshal(to)
}

// used to postpone yaml unmarshaling
type yamlRawMessage struct {
	unmarshal func(interface{}) error
}

func (msg *yamlRawMessage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

type tomlConfig map[string]toml.Primitive

func (t tomlConfig) UnmarshalValue(_ context.Context, key string, to interface{}) error {
	fmt.Println("toml.UnmarshalValue")
	fmt.Println("t ==>", t)
	fmt.Println("key ==>", key)
	v, ok := t[key]
	fmt.Println("v ==>", v)
	fmt.Println("ok ==>", ok)
	if !ok {
		fmt.Println("wesh")
		return backend.ErrNotFound
	}
	fmt.Printf("---\n---\n---\n")
	return toml.PrimitiveDecode(v, to)
}
