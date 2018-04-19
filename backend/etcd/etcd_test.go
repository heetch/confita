package etcd

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

func TestEtcdBackend(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	defer client.Close()

	b := NewBackend(client, WithPrefix("prefix"))
	_, err = b.Get(context.Background(), "something that doesn't exist")
	require.Equal(t, backend.ErrNotFound, err)
}

func TestEtcdBackendWithPrefetch(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	defer client.Close()

	prefix := "confita-tests"

	ctx := context.Background()

	defer client.KV.Delete(ctx, prefix, clientv3.WithPrefix())

	b := NewBackend(client, WithPrefix(prefix), WithPrefetch())

	t.Run("OK", func(t *testing.T) {
		_, err = client.KV.Put(ctx, prefix+"/key1", "value1")
		require.NoError(t, err)

		_, err = client.KV.Put(ctx, prefix+"/key2", "value2")
		require.NoError(t, err)

		_, err = client.KV.Put(ctx, prefix+"/key3", "value3")
		require.NoError(t, err)

		val, err := b.Get(ctx, "key1")
		require.NoError(t, err)

		// deleting the tree
		client.KV.Delete(ctx, prefix, clientv3.WithPrefix())

		// WithPrefetch should have prefetched all the keys
		// they should be available even if the tree has been removed.
		val, err = b.Get(ctx, "key1")
		require.NoError(t, err)
		require.Equal(t, []byte("value1"), val)

		val, err = b.Get(ctx, "key2")
		require.NoError(t, err)
		require.Equal(t, []byte("value2"), val)

		val, err = b.Get(ctx, "key3")
		require.NoError(t, err)
		require.Equal(t, []byte("value3"), val)
	})

}
