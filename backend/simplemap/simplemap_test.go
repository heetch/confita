package simplemap_test

import (
	"context"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/simplemap"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSimplemapBackend(t *testing.T) {
	type config struct {
		Name     string        `config:"Name"`
		Age      int           `config:"Age"`
		Negative int           `config:"Negative"`
		Balance  float64       `config:"Balance"`
		Timeout  time.Duration `config:"Timeout"`
	}

	b := simplemap.NewBackend(map[string]interface{}{
		"Name":     "Bob",
		"Age":      30,
		"Negative": -10,
		"Balance":  1234.56,
		"Timeout":  10 * time.Second,
	})

	var c config
	require.Equal(t, "simplemap", b.Name())
	err := confita.NewLoader(b).Load(context.Background(), &c)

	require.NoError(t, err)
	require.Equal(t, "Bob", c.Name)
	require.Equal(t, 30, c.Age)
	require.Equal(t, -10, c.Negative)
	require.Equal(t, 1234.56, c.Balance)
	require.Equal(t, 10*time.Second, c.Timeout)
}
