package vault

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			map[string]interface{}{
				"foo":    "bar",
				"cheese": "nan",
			})
		require.NoError(t, err)

		val, err := b.Get(context.Background(), "foo")
		require.NoError(t, err)
		assert.Equal(t, "bar", string(val))

		val, err = b.Get(context.Background(), "cheese")
		require.NoError(t, err)
		assert.Equal(t, "nan", string(val))
	})

	t.Run("NotFound", func(t *testing.T) {
		b := NewBackend(c, path)
		_, err := b.Get(context.Background(), "badKey")
		require.EqualError(t, err, backend.ErrNotFound.Error())
	})
}
