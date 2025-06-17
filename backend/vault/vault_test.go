package vault

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/heetch/confita/backend"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}
func TestVaultBackend(t *testing.T) {
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)

	client.SetToken("root")
	c := client.Logical()

	path := "secret/test"

	defer c.Delete(path)

	t.Run("SecretPathNotFound", func(t *testing.T) {
		b := NewBackend(c, path)
		_, err := b.Get(context.Background(), "foo")
		require.EqualError(t, err, "secret not found at the following path: secret/test")
	})

	t.Run("OK", func(t *testing.T) {
		b := NewBackend(c, path)

		_, err = c.Write(path,
			map[string]any{
				"foo":  "bar",
				"data": "nan",
			})
		require.NoError(t, err)

		val, err := b.Get(context.Background(), "foo")
		require.NoError(t, err)
		assert.Equal(t, "bar", string(val))

		val, err = b.Get(context.Background(), "data")
		require.NoError(t, err)
		assert.Equal(t, "nan", string(val))
	})

	t.Run("NotFound", func(t *testing.T) {
		b := NewBackend(c, path)
		_, err := b.Get(context.Background(), "badKey")
		require.EqualError(t, err, backend.ErrNotFound.Error())
	})
}

func TestVaultBackendV2(t *testing.T) {
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8222")
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)

	client.SetToken("root")
	c := client.Logical()

	randPathSuffix := randStringBytes(5)
	path := "secret/data/" + randPathSuffix

	defer c.Delete(path)

	t.Run("SecretPathNotFound", func(t *testing.T) {
		b := NewBackendV2(c, path)
		_, err := b.Get(context.Background(), "foo")
		require.EqualError(t, err, "secret not found at the following path: "+path)
	})

	okTests := []struct {
		name string
		path string
	}{
		{
			"OK v2 data path",
			path,
		},
		{
			"OK old path",
			"secret/" + randPathSuffix,
		},
	}
	for _, okTest := range okTests {
		t.Run(okTest.name, func(t *testing.T) {
			b := NewBackendV2(c, okTest.path)

			// For writing we use the Consul client directly,
			// so we need to use the full proper path.
			_, err = c.Write(path,
				map[string]any{
					"data": map[string]string{
						"foo":  "bar",
						"data": "nan",
					},
				})
			require.NoError(t, err)

			val, err := b.Get(context.Background(), "foo")
			require.NoError(t, err)
			assert.Equal(t, "bar", string(val))

			val, err = b.Get(context.Background(), "data")
			require.NoError(t, err)
			assert.Equal(t, "nan", string(val))
		})
	}

	t.Run("NotFound", func(t *testing.T) {
		b := NewBackendV2(c, path)
		_, err := b.Get(context.Background(), "badKey")
		require.EqualError(t, err, backend.ErrNotFound.Error())
	})

}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
