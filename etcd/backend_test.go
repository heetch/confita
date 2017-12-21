package etcd

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/confita"
	"github.com/stretchr/testify/require"
)

func TestEtcdBackend(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	defer client.Close()

	b := NewBackend(client, "prefix")
	_, err = b.Get(context.Background(), "something that doesn't exist")
	require.Equal(t, confita.ErrNotFound, err)
}
