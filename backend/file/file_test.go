package file_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/heetch/confita/backend/file"
	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T, name, content string) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "confita")
	require.NoError(t, err)

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	require.NoError(t, err)

	fmt.Fprintf(f, content)

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

	testLoad := func(t *testing.T, path string) {
		var c config
		b := file.NewBackend(path, false)

		err := b.Unmarshal(context.Background(), &c)
		require.NoError(t, err)
		require.Equal(t, "some name", c.Name)
		require.Equal(t, 10, c.Age)
		require.Equal(t, 10*time.Nanosecond, c.Timeout)
	}

	t.Run("JSON", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.json", `{
			"name": "some name",
			"age": 10,
			"timeout": 10
		}`)
		defer cleanup()

		testLoad(t, path)
	})

	t.Run("YAML", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.yml", `
  name: "some name"
  age: 10
  timeout: 10ns
`)
		defer cleanup()

		testLoad(t, path)
	})

	t.Run("TOML", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.toml",
			`name = "some name"
age = 10
timeout = 10
`)
		defer cleanup()

		testLoad(t, path)
	})

	t.Run("Unsupported extension", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.xml", `{
			"name": "some name"
		}`)
		defer cleanup()

		var c config
		b := file.NewBackend(path, false)

		err := b.Unmarshal(context.Background(), &c)
		require.Error(t, err)
	})

	t.Run("Required file not found", func(t *testing.T) {
		var c config
		b := file.NewBackend("some path", false)

		err := b.Unmarshal(context.Background(), &c)
		require.Error(t, err)
	})

	t.Run("Optional file not found", func(t *testing.T) {
		var c config
		b := file.NewBackend("some path", true)

		err := b.Unmarshal(context.Background(), &c)
		require.Error(t, err)
		_, ok := err.(file.ErrOpenOptionalFile)
		require.True(t, ok)
	})
}
