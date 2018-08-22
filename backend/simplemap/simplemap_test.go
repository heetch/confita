package simplemap_test

import (
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/heetch/confita/backend/simplemap"
	"time"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestSimplemapBackend(t *testing.T) {
	type config struct {
		Name    string
		Age     int
		Balance float64
		Timeout time.Duration
	}

	b := simplemap.NewBackend(map[string]interface{}{
		"Name": "Bob",
		"Age": uint(30),
		"Negative": -10,
		"Balance": 1234.56,
		"Timeout": 10 * time.Second,
	})
	
	assert.Equal(t, "simplemap", b.Name())

	val, err := b.Get(context.Background(), "Name")
	require.NoError(t, err)
	assert.Equal(t, "Bob", string(val))

	val, err = b.Get(context.Background(), "Age")
	require.NoError(t, err)
	assert.Equal(t, "30", string(val))

	val, err = b.Get(context.Background(), "Negative")
	require.NoError(t, err)
	assert.Equal(t, "-10", string(val))

	val, err = b.Get(context.Background(), "Balance")
	require.NoError(t, err)
	assert.Equal(t, "1234.56", string(val))

	val, err = b.Get(context.Background(), "Timeout")
	require.NoError(t, err)
	assert.Equal(t, "10s", string(val))

	_, err = b.Get(context.Background(), "NotExists")
	require.Error(t, err)
	assert.Equal(t, backend.ErrNotFound, err)
}