package flags

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/heetch/confita"
	"github.com/pkg/errors"
)

// Backend that loads configuration from the command line flags.
type Backend struct {
	flags *flag.FlagSet
}

// NewBackend creates a flags backend.
func NewBackend() *Backend {
	return &Backend{
		flags: flag.CommandLine,
	}
}

// LoadStruct takes a struct config, define flags based on it and parse the command line args.
func (b *Backend) LoadStruct(ctx context.Context, cfg *confita.StructConfig) error {
	for _, field := range cfg.Fields {
		f := field

		if f.Backend != "" && f.Backend != b.Name() {
			continue
		}

		// Display all the flags and their default values but override the field only if the user has explicitely
		// set the flag.
		k := f.Value.Kind()
		switch {
		case f.Value.Type().String() == "time.Duration":
			var val time.Duration
			b.flags.DurationVar(&val, f.Key, time.Duration(f.Default.Int()), f.Description)
			if f.Short != "" {
				b.flags.DurationVar(&val, f.Short, time.Duration(f.Default.Int()), shortDesc(f.Description))
			}
			// this function must be executed after the flag.Parse call.
			defer func() {
				// if the user has set the flag, save the value in the field.
				if b.isFlagSet(f) {
					f.Value.SetInt(int64(val))
				}
			}()
		case k == reflect.Bool:
			var val bool
			b.flags.BoolVar(&val, f.Key, f.Default.Bool(), f.Description)
			if f.Short != "" {
				b.flags.BoolVar(&val, f.Short, f.Default.Bool(), shortDesc(f.Description))
			}
			defer func() {
				if b.isFlagSet(f) {
					f.Value.SetBool(val)
				}
			}()
		case k >= reflect.Int && k <= reflect.Int64:
			var val int
			b.flags.IntVar(&val, f.Key, int(f.Default.Int()), f.Description)
			if f.Short != "" {
				b.flags.IntVar(&val, f.Short, int(f.Default.Int()), shortDesc(f.Description))
			}
			defer func() {
				if b.isFlagSet(f) {
					f.Value.SetInt(int64(val))
				}
			}()
		case k >= reflect.Uint && k <= reflect.Uint64:
			var val uint64
			b.flags.Uint64Var(&val, f.Key, f.Default.Uint(), f.Description)
			if f.Short != "" {
				b.flags.Uint64Var(&val, f.Short, f.Default.Uint(), shortDesc(f.Description))
			}
			defer func() {
				if b.isFlagSet(f) {
					f.Value.SetUint(val)
				}
			}()
		case k >= reflect.Float32 && k <= reflect.Float64:
			var val float64
			b.flags.Float64Var(&val, f.Key, f.Default.Float(), f.Description)
			if f.Short != "" {
				b.flags.Float64Var(&val, f.Short, f.Default.Float(), shortDesc(f.Description))
			}
			defer func() {
				if b.isFlagSet(f) {
					f.Value.SetFloat(val)
				}
			}()
		case k == reflect.String:
			var val string
			b.flags.StringVar(&val, f.Key, f.Default.String(), f.Description)
			if f.Short != "" {
				b.flags.StringVar(&val, f.Short, f.Default.String(), shortDesc(f.Description))
			}
			defer func() {
				if b.isFlagSet(f) {
					f.Value.SetString(val)
				}
			}()
		default:
			b.flags.Var(&flagValue{f}, f.Key, f.Description)
		}
	}

	// Note: in the usual case, when b.flags is flag.CommandLine, this will exit
	// rather than returning an error.
	return b.flags.Parse(os.Args[1:])
}

func (b *Backend) isFlagSet(config *confita.FieldConfig) bool {
	ok := false
	b.flags.Visit(func(f *flag.Flag) {
		if f.Name == config.Key || f.Name == config.Short {
			ok = true
		}
	})
	return ok
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

func (f *flagValue) Get() interface{} {
	return f.Default.Interface()
}

// Get is not implemented.
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Name returns the name of the flags backend.
func (b *Backend) Name() string {
	return "flags"
}

func shortDesc(description string) string {
	return fmt.Sprintf("%s (short)", description)
}
