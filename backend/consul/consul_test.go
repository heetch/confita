package consul

import (
	"context"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

func TestConsulBackend(t *testing.T) {
	prefix := "confita-tests"

	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	defer client.KV().DeleteTree(prefix, nil)

	b := NewBackend(client, prefix)

	t.Run("NotFound", func(t *testing.T) {
		_, err = b.Get(context.Background(), "something that doesn't exist")
		require.Equal(t, backend.ErrNotFound, err)
	})

	t.Run("OK", func(t *testing.T) {
		_, err = client.KV().Put(&api.KVPair{Key: prefix + "/key1", Value: []byte("value")}, nil)
		require.NoError(t, err)

		val, err := b.Get(context.Background(), "key1")
		require.NoError(t, err)
		require.Equal(t, []byte("value"), val)
	})

	t.Run("Canceled", func(t *testing.T) {
		_, err = client.KV().Put(&api.KVPair{Key: prefix + "/key2", Value: []byte("value")}, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = b.Get(ctx, "key2")
		require.Error(t, err)
	})
}
