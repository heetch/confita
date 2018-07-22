package strukt

import (
	"context"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStruktBackend(t *testing.T) {
	s := struct {
		myString      string
		myInt         int
		myBool        bool
		myStringSlice []string
	}{
		myString:      "Greg",
		myInt:         8,
		myBool:        true,
		myStringSlice: []string{"one", "two", "three", "four"},
	}

	b := NewBackend(s)

	t.Run("NotFound", func(t *testing.T) {
		_, err := b.Get(context.Background(), "something that doesn't exist")
		require.Equal(t, backend.ErrNotFound, err)
	})

	t.Run("MatchString", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myString")
		require.NoError(t, err)
		require.Equal(t, "Greg", string(val))
	})

	t.Run("MatchInt", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myInt")
		require.NoError(t, err)
		require.Equal(t, "8", string(val))
	})

	t.Run("MatchBool", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myBool")
		require.NoError(t, err)
		require.Equal(t, "true", string(val))
	})

	t.Run("MatchSlice", func(t *testing.T) {
		val, err := b.Get(context.Background(), "myStringSlice")
		require.NoError(t, err)
		require.Equal(t, "one,two,three,four", string(val))
	})
}
