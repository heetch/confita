package flags

import (
	"context"
	"flag"
	"reflect"
	"time"

	"github.com/heetch/confita"
	"github.com/pkg/errors"
)

// Backend that loads configuration from the command line flags.
type Backend struct{}

// NewBackend creates a flags backend.
func NewBackend() *Backend {
	return new(Backend)
}

// LoadStruct takes a struct config, define flags based on it and parse the command line args.
func (b *Backend) LoadStruct(ctx context.Context, cfg *confita.StructConfig) error {
	for _, field := range cfg.Fields {
		f := field

		if f.Backend != "" && f.Backend != b.Name() {
			continue
		}

		k := f.Value.Kind()
		switch {
		case f.Value.Type().String() == "time.Duration":
			flag.DurationVar(f.Value.Addr().Interface().(*time.Duration), f.Key, time.Duration(f.Value.Int()), "")
		case k == reflect.Bool:
			flag.BoolVar(f.Value.Addr().Interface().(*bool), f.Key, f.Value.Bool(), "")
		case k >= reflect.Int && k <= reflect.Int64:
			v := flag.Int(f.Key, int(f.Value.Int()), "")
			defer func() {
				f.Value.SetInt(int64(*v))
			}()
		case k >= reflect.Uint && k <= reflect.Uint64:
			v := flag.Uint(f.Key, uint(f.Value.Uint()), "")
			defer func() {
				f.Value.SetUint(uint64(*v))
			}()
		case k >= reflect.Float32 && k <= reflect.Float64:
			v := flag.Float64(f.Key, f.Value.Float(), "")
			defer func() {
				f.Value.SetFloat(*v)
			}()
		case k == reflect.String:
			flag.StringVar(f.Value.Addr().Interface().(*string), f.Key, f.Value.String(), "")
		default:
			flag.Var(&flagValue{f}, f.Key, "")
		}
	}

	flag.Parse()

	return nil
}

type flagValue struct {
	*confita.FieldConfig
}

func (f *flagValue) String() string {
	if f.FieldConfig == nil {
		return ""
	}

	return f.Key
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the type of the file.
func (b *Backend) Name() string {
	return "flags"
}
