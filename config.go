package confita

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrNotFound is returned by the backends or by the Load function when the given key is not found.
	ErrNotFound = errors.New("key not found")
)

// Loader loads configuration keys from backends and stores them is a struct.
type Loader struct {
	timeout  time.Duration
	backends []Backend
}

// NewLoader creates a Loader. If no backend is specified, the loader uses the environment.
func NewLoader(options ...Option) *Loader {
	var l Loader
	for _, opt := range options {
		opt(&l)
	}

	if len(l.backends) == 0 {
		l.backends = append(l.backends, EnvBackend())
	}

	return &l
}

// Load analyses all the fields of the given struct for a "config" tag and queries each backend
// in order for the corresponding key.
func (l *Loader) Load(to interface{}) error {
	ctx := context.Background()
	if l.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, l.timeout)
		defer cancel()
	}

	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return errors.New("provided target must be a pointer to struct")
	}

	ref = ref.Elem()
	return l.parseStruct(ctx, &ref)
}

func (l *Loader) parseStruct(ctx context.Context, ref *reflect.Value) error {
	t := ref.Type()

	numFields := ref.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		value := ref.Field(i)
		typ := value.Type()

		// skip if field is unexported
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("config")
		if tag == "-" {
			continue
		}

		// if struct or *struct, parse recursively
		switch {
		case typ.Kind() == reflect.Struct:
			err := l.parseStruct(ctx, &value)
			if err != nil {
				return err
			}

			continue
		case typ.Kind() == reflect.Ptr:
			if value.Type().Elem().Kind() == reflect.Struct && !value.IsNil() {
				value = value.Elem()

				err := l.parseStruct(ctx, &value)
				if err != nil {
					return err
				}

				continue
			}
		}

		if tag == "" {
			continue
		}

		key := tag
		var required bool

		if idx := strings.Index(tag, ","); idx != -1 {
			key = tag[:idx]
			if tag[idx+1:] == "required" {
				required = true
			}
		}

		var found bool
		for _, b := range l.backends {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			raw, err := b.Get(ctx, key)
			if err != nil {
				if err == ErrNotFound {
					continue
				}

				return err
			}

			err = convert(string(raw), &value)
			if err != nil {
				return err
			}

			found = true
			break
		}

		if required && !found {
			return fmt.Errorf("required key '%s' for field '%s' not found", key, field.Name)
		}
	}

	return nil
}

func convert(data string, value *reflect.Value) error {
	k := value.Kind()
	t := value.Type()

	switch {
	case t.String() == "time.Duration":
		d, err := time.ParseDuration(data)
		if err != nil {
			return err
		}

		value.SetInt(int64(d))
	case k == reflect.Bool:
		b, err := strconv.ParseBool(data)
		if err != nil {
			return err
		}

		value.SetBool(b)
	case k >= reflect.Int && k <= reflect.Int64:
		i, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}

		value.SetInt(i)
	case k >= reflect.Uint && k <= reflect.Uint64:
		i, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			return err
		}

		value.SetUint(i)
	case k >= reflect.Float32 && k <= reflect.Float64:
		f, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return err
		}

		value.SetFloat(f)
	case k == reflect.String:
		value.SetString(data)
	case k == reflect.Ptr:
		n := reflect.New(value.Type().Elem())
		value.Set(n)
		e := n.Elem()
		return convert(data, &e)
	default:
		return fmt.Errorf("field type '%s' not supported", k)
	}

	return nil
}

// An Option is a function that configures a Loader.
type Option func(*Loader)

// Backends configures the loader to use the given backends.
// If this option is not used, the loader will load from the environment.
func Backends(backends ...Backend) func(*Loader) {
	return func(l *Loader) {
		l.backends = append(l.backends, backends...)
	}
}

// Timeout sets the timeout for the entire configuration load process.
func Timeout(t time.Duration) func(*Loader) {
	return func(l *Loader) {
		l.timeout = t
	}
}
