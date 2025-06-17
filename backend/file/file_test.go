package file_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/file"
	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T, name, content string) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "confita")
	require.NoError(t, err)

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = fmt.Fprintf(f, "%s", content)
	require.NoError(t, err)

	require.NoError(t, f.Close())

	return path, func() {
		require.NoError(t, os.RemoveAll(dir))
	}
}

func TestFileBackend(t *testing.T) {
	type config struct {
		Name    string
		Age     int
		Timeout time.Duration
	}

	ekv := config{
		Name:    "some name",
		Age:     10,
		Timeout: 10,
	}

	testLoad := func(t *testing.T, path string, template any, expected any) {
		b := file.NewBackend(path)

		err := b.Unmarshal(context.Background(), template)
		require.NoError(t, err)
		require.EqualValues(t, expected, template)
	}

	t.Run("JSON", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.json", `{
			"name": "some name",
			"age": 10,
			"timeout": 10
		}`)
		defer cleanup()

		testLoad(t, path, &config{}, &ekv)
	})

	t.Run("YAML", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.yml", `
  name: "some name"
  age: 10
  timeout: 10ns
`)
		defer cleanup()

		testLoad(t, path, &config{}, &ekv)
	})
	t.Run("TOML", func(t *testing.T) {

		t.Run("Simple KV", func(t *testing.T) {
			path, cleanup := createTempFile(t, "config.toml",
				`name = "some name"
age = 10
timeout = 10
`)
			defer cleanup()

			testLoad(t, path, &config{}, &ekv)
		})

		t.Run("TOML Table", func(t *testing.T) {
			path, cleanup := createTempFile(t, "config.toml",
				`title = "title!"
[Config]
name = "some name"
age = 10
timeout = 10
`)
			defer cleanup()
			type Config struct {
				Title  string
				Config config
			}
			e := Config{
				Title: "title!",
				Config: config{
					Name:    "some name",
					Age:     10,
					Timeout: 10,
				},
			}

			testLoad(t, path, &Config{}, &e)
		})

		t.Run("TOML Array of Tables", func(t *testing.T) {
			path, cleanup := createTempFile(t, "config.toml",
				`[[Config]]
name = "Alice"
age = 10
timeout = 10
[[Config]]
name = "Bob"
age = 11
timeout = 11
`)
			defer cleanup()
			type Configs struct {
				Config []config
			}
			e := Configs{Config: []config{{Name: "Alice", Age: 10, Timeout: 10}, {Name: "Bob", Age: 11, Timeout: 11}}}

			testLoad(t, path, &Configs{}, &e)
		})

	})

	t.Run("Unsupported extension", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.xml", `{
			"name": "some name"
		}`)
		defer cleanup()

		var c config
		b := file.NewBackend(path)

		err := b.Unmarshal(context.Background(), &c)
		require.Error(t, err)
	})

	t.Run("Required file not found", func(t *testing.T) {
		var c config
		b := file.NewBackend("some path")

		err := b.Unmarshal(context.Background(), &c)
		require.Error(t, err)
	})

	t.Run("Optional file not found", func(t *testing.T) {
		var c config
		b := file.NewOptionalBackend("some path")

		err := b.Unmarshal(context.Background(), &c)
		require.EqualError(t, err, backend.ErrNotFound.Error())
	})
}
