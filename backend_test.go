package confita

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvBackend(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		b := EnvBackend()

		_, err := b.Get(context.Background(), "something that doesn't exist")
		require.Equal(t, ErrNotFound, err)
	})

	t.Run("ExactMatch", func(t *testing.T) {
		b := EnvBackend()

		os.Setenv("TESTCONFIG1", "ok")
		val, err := b.Get(context.Background(), "TESTCONFIG1")
		require.NoError(t, err)
		require.Equal(t, "ok", string(val))
	})

	t.Run("DifferentCase", func(t *testing.T) {
		b := EnvBackend()

		os.Setenv("TEST_CONFIG_2", "ok")
		val, err := b.Get(context.Background(), "test-config-2")
		require.NoError(t, err)
		require.Equal(t, "ok", string(val))
	})
}
