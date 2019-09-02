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

		// Display all the flags and their default values but override the field only if the user has explicitely
		// set the flag.
		k := f.Value.Kind()
		switch {
		case f.Value.Type().String() == "time.Duration":
			var val time.Duration
			flag.DurationVar(&val, f.Key, time.Duration(f.Default.Int()), "")
			if f.Short != "" {
				flag.DurationVar(&val, f.Short, time.Duration(f.Default.Int()), "")
			}
			// this function must be executed after the flag.Parse call.
			defer func() {
				// if the user has set the flag, save the value in the field.
				if isFlagSet(f) {
					f.Value.SetInt(int64(val))
				}
			}()
		case k == reflect.Bool:
			var val bool
			flag.BoolVar(&val, f.Key, f.Default.Bool(), "")
			if f.Short != "" {
				flag.BoolVar(&val, f.Short, f.Default.Bool(), "")
			}
			defer func() {
				if isFlagSet(f) {
					f.Value.SetBool(val)
				}
			}()
		case k >= reflect.Int && k <= reflect.Int64:
			var val int
			flag.IntVar(&val, f.Key, int(f.Default.Int()), "")
			if f.Short != "" {
				flag.IntVar(&val, f.Short, int(f.Default.Int()), "")
			}
			defer func() {
				if isFlagSet(f) {
					f.Value.SetInt(int64(val))
				}
			}()
		case k >= reflect.Uint && k <= reflect.Uint64:
			v := flag.Uint(f.Key, uint(f.Default.Uint()), "")
			defer func() {
				if isFlagSet(f) {
					f.Value.SetUint(uint64(*v))
				}
			}()
		case k >= reflect.Float32 && k <= reflect.Float64:
			v := flag.Float64(f.Key, f.Default.Float(), "")
			defer func() {
				if isFlagSet(f) {
					f.Value.SetFloat(*v)
				}
			}()
		case k == reflect.String:
			var val string
			flag.StringVar(&val, f.Key, f.Default.String(), "")
			if f.Short != "" {
				flag.StringVar(&val, f.Short, f.Default.String(), "")
			}
			defer func() {
				if isFlagSet(f) {
					f.Value.SetString(val)
				}
			}()
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

func isFlagSet(config *confita.FieldConfig) bool {
	flagset := make(map[*confita.FieldConfig]bool)
	flag.Visit(func(f *flag.Flag) { flagset[config] = true })

	_, ok := flagset[config]
	return ok
}
