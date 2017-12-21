package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvBackend(t *testing.T) {
	b := FromEnv()

	_, err := b.Get(context.Background(), "something that doesn't exist")
	require.Equal(t, ErrNotFound, err)
}
