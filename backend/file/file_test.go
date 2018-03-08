package file_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/heetch/confita/backend"
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
		require.NoError(t, os.RemoveAll(path))
	}
}

func TestFileBackend(t *testing.T) {
	testLoad := func(t *testing.T, path string) {
		tests := []struct {
			key      string
			expected interface{}
			target   interface{}
		}{
			{"name", "some name", new(string)},
			{"age", 10, new(int)},
			{"timeout", 10 * time.Nanosecond, new(time.Duration)},
		}

		b := file.NewBackend(path)

		for _, test := range tests {
			err := b.UnmarshalValue(context.Background(), test.key, test.target)
			require.NoError(t, err)
			// dirty trick to fetch the actual value instead of its pointer
			actual := reflect.Indirect(reflect.ValueOf(test.target)).Interface()
			require.Equal(t, test.expected, actual)
		}
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

	t.Run("Unsupported extension", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.xml", `{
			"name": "some name"
		}`)
		defer cleanup()

		b := file.NewBackend(path)
		var name string
		err := b.UnmarshalValue(context.Background(), "name", &name)
		require.Error(t, err)
	})

	t.Run("File not found", func(t *testing.T) {
		b := file.NewBackend("some path")
		var name string
		err := b.UnmarshalValue(context.Background(), "name", &name)
		require.Error(t, err)
	})

	t.Run("JSON Key not found", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.json", `{
			"age": 10,
			"timeout": 10
		}`)
		defer cleanup()

		b := file.NewBackend(path)

		var name string
		err := b.UnmarshalValue(context.Background(), "name", &name)
		require.Equal(t, backend.ErrNotFound, err)
	})

	t.Run("YAMl Key not found", func(t *testing.T) {
		path, cleanup := createTempFile(t, "config.yml", `
  age: 10
  timeout: 10ns
`)
		defer cleanup()

		b := file.NewBackend(path)

		var name string
		err := b.UnmarshalValue(context.Background(), "name", &name)
		require.Equal(t, backend.ErrNotFound, err)
	})
}
