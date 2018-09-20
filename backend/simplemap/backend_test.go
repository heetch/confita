package simplemap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/heetch/confita/backend"
)

func TestMapBackend(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		b := NewBackend(make(map[string]string))
		_, err := b.Get(context.Background(), "I am not really there")
		require.Equal(t, backend.ErrNotFound, err)
	})
}
