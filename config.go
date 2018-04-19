package confita

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
)

// Loader loads configuration keys from backends and stores them is a struct.
type Loader struct {
	backends []backend.Backend

	// Tag specifies the tag name used to parse
	// configuration keys and options.
	// If empty, "config" is used.
	Tag string
}

// NewLoader creates a Loader. If no backend is specified, the loader uses the environment.
func NewLoader(backends ...backend.Backend) *Loader {
	l := Loader{
		backends: backends,
	}

	if len(l.backends) == 0 {
		l.backends = append(l.backends, env.NewBackend())
	}

	return &l
}

type fieldConfig struct {
	Name     string
	Key      string
	Value    *reflect.Value
	Required bool
	Backend  string
}

// Load analyses all the fields of the given struct for a "config" tag and queries each backend
// in order for the corresponding key. The given context can be used for timeout and cancelation.
func (l *Loader) Load(ctx context.Context, to interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return errors.New("provided target must be a pointer to struct")
	}

	ref = ref.Elem()

	fields := l.parseStruct(&ref)

	return l.resolve(ctx, fields)
}

func (l *Loader) parseStruct(ref *reflect.Value) []*fieldConfig {
	var list []*fieldConfig

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

		tagKey := l.Tag
		if tagKey == "" {
			tagKey = "config"
		}

		tag := field.Tag.Get(tagKey)
		if tag == "-" {
			continue
		}

		// if struct or *struct, parse recursively
		switch {
		case typ.Kind() == reflect.Struct:
			list = append(list, l.parseStruct(&value)...)
			continue
		case typ.Kind() == reflect.Ptr:
			if value.Type().Elem().Kind() == reflect.Struct && !value.IsNil() {
				value = value.Elem()
				list = append(list, l.parseStruct(&value)...)
				continue
			}
		}

		// empty tag or no tag, skip the field
		if tag == "" {
			continue
		}

		f := fieldConfig{
			Name:  field.Name,
			Key:   tag,
			Value: &value,
		}

		if idx := strings.Index(tag, ","); idx != -1 {
			f.Key = tag[:idx]
			opts := strings.Split(tag[idx+1:], ",")

			for _, opt := range opts {
				if opt == "required" {
					f.Required = true
				}

				if strings.HasPrefix(opt, "backend=") {
					f.Backend = opt[len("backend="):]
				}
			}
		}

		list = append(list, &f)
	}

	return list
}

func (l *Loader) resolve(ctx context.Context, fields []*fieldConfig) error {
	for _, f := range fields {
		var found bool
		var backendFound bool

		for _, b := range l.backends {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if f.Backend != "" && f.Backend != b.Name() {
				continue
			}

			backendFound = true

			if u, ok := b.(backend.ValueUnmarshaler); ok {
				err := u.UnmarshalValue(ctx, f.Key, f.Value.Addr().Interface())
				if err != nil && err != backend.ErrNotFound {
					return err
				}

				continue
			}

			raw, err := b.Get(ctx, f.Key)
			if err != nil {
				if err == backend.ErrNotFound {
					continue
				}

				return err
			}

			err = convert(string(raw), f.Value)
			if err != nil {
				return err
			}

			found = true
			break
		}

		if f.Backend != "" && !backendFound {
			return fmt.Errorf("the backend: '%s' is not supported", f.Backend)
		}

		if f.Required && !found {
			return fmt.Errorf("required key '%s' for field '%s' not found", f.Key, f.Name)
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
