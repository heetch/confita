package env

import (
	"testing"

	config "github.com/heetch/go-config"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	var b Backend

	_, err := b.Get("something that doesn't exist")
	require.Equal(t, config.ErrNotFound, err)
}
