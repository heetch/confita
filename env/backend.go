package env

import (
	"syscall"

	config "github.com/heetch/go-config"
)

// Backend loads keys from the environment.
type Backend struct{}

// Get loads the given key from the environment.
func (b *Backend) Get(key string) ([]byte, error) {
	val, ok := syscall.Getenv(key)
	if !ok {
		return nil, config.ErrNotFound
	}

	return []byte(val), nil
}

// FromEnv creates a configuration loader that loads from the environment.
func FromEnv() *Backend {
	return new(Backend)
}
