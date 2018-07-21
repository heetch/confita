package strukt

import (
	"context"
	"testing"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

func TestStruktBackend(t *testing.T) {
	s := struct {
		myName string
		myAge  int
	}{
		myName: "Greg",
		myAge:  8,
	}

	b := NewBackend(s)

	t.Run("NotFound", func(t *testing.T) {
		_, err := b.Get(context.Background(), "something that doesn't exist")
		require.Equal(t, backend.ErrNotFound, err)
	})

	t.Run("MatchString", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myName")
		require.NoError(t, err)
		require.Equal(t, "Greg", string(val))
	})

	t.Run("MatchInt", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myAge")
		require.NoError(t, err)
		require.Equal(t, "8", string(val))
	})
}
