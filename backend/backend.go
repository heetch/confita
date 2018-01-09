package backend

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned by the backends given key is not found.
	ErrNotFound = errors.New("key not found")
)

// A Backend is used to fetch values from a given key.
type Backend interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// Func creates a Backend from a function.
func Func(fn func(context.Context, string) ([]byte, error)) Backend {
	return &backendFunc{fn: fn}
}

type backendFunc struct {
	fn func(context.Context, string) ([]byte, error)
}

func (b *backendFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return b.fn(ctx, key)
}
