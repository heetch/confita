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

type structConfig struct {
	s      interface{}
	fields []*fieldConfig
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

	s := l.parseStruct(&ref)
	s.s = to
	return l.resolve(ctx, s)
}

func (l *Loader) parseStruct(ref *reflect.Value) *structConfig {
	var s structConfig

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
			s.fields = append(s.fields, l.parseStruct(&value).fields...)
			continue
		case typ.Kind() == reflect.Ptr:
			if value.Type().Elem().Kind() == reflect.Struct && !value.IsNil() {
				value = value.Elem()
				s.fields = append(s.fields, l.parseStruct(&value).fields...)
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

		s.fields = append(s.fields, &f)
	}

	return &s
}

func (l *Loader) resolve(ctx context.Context, s *structConfig) error {
	for _, f := range s.fields {
		if f.Backend != "" {
			var found bool
			for _, b := range l.backends {
				if b.Name() == f.Backend {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("the backend: '%s' is not supported", f.Backend)
			}
		}
	}

	for _, b := range l.backends {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if u, ok := b.(backend.Unmarshaler); ok {
			err := u.Unmarshal(ctx, s.s)
			if err != nil {
				return err
			}

			continue
		}

		for _, f := range s.fields {
			if f.Backend != "" && f.Backend != b.Name() {
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
		}
	}

	for _, f := range s.fields {
		if f.Required && isZero(f.Value) {
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

func isZero(v *reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	current := v.Interface()
	return reflect.DeepEqual(current, zero)
}
