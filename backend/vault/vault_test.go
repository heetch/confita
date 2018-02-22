package vault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clientMock struct{}

func (c *clientMock) Read(p string) (*api.Secret, error) {
	if p == "wrong/path" {
		return nil, nil
	}

	return &api.Secret{
		Data: map[string]interface{}{
			"foo": "baz",
		},
	}, nil
}

func TestGet(t *testing.T) {
	b := Backend{
		client: &clientMock{},
		path:   "good/path",
	}

	ctx := context.Background()
	v, err := b.Get(ctx, "foo")
	require.NoError(t, err)

	assert.Equal(t, "baz", string(v))

	_, err = b.Get(ctx, "bar")
	require.Error(t, err)

	assert.EqualError(t, backend.ErrNotFound, err.Error())
}

func TestGetWrongPath(t *testing.T) {
	b := Backend{
		client: &clientMock{},
		path:   "wrong/path",
	}

	ctx := context.Background()
	_, err := b.Get(ctx, "foo")
	require.Error(t, err)

	assert.EqualError(t, backend.ErrNotFound, err.Error())
}
