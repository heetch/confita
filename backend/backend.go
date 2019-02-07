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
	Name() string
}

// Func creates a Backend from a function.
func Func(name string, fn func(context.Context, string) ([]byte, error)) Backend {
	return &backendFunc{fn: fn, name: name}
}

// FuncWithDefaults creates a Backend from a function and allows users to add a default map.
func FuncWithDefaults(name string, defaults map[string]string, fn func(context.Context, string) ([]byte, error)) Backend {
	return &backendFunc{fn: fn, name: name, defaults: defaults}
}

type backendFunc struct {
	fn   func(context.Context, string) ([]byte, error)
	name string
	defaults map[string]string
}

func (b *backendFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return b.fn(ctx, key)
}

func (b *backendFunc) Name() string {
	return b.name
}
