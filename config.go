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

// Unmarshaler can be implemented by backends to receive the struct directly and load values into it.
type Unmarshaler interface {
	Unmarshal(ctx context.Context, to interface{}) error
}

// StructLoader can be implemented by backends to receive the parsed struct informations and load values into it.
type StructLoader interface {
	LoadStruct(ctx context.Context, cfg *StructConfig) error
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

// Load analyses all the Fields of the given struct for a "config" tag and queries each backend
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

	s := l.parseStruct(ref)
	s.S = to
	return l.resolve(ctx, s)
}

func (l *Loader) parseStruct(ref reflect.Value) *StructConfig {
	var s StructConfig

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
		switch typ.Kind() {
		case reflect.Struct:
			s.Fields = append(s.Fields, l.parseStruct(value).Fields...)
			continue
		case reflect.Ptr:
			if typ.Elem().Kind() == reflect.Struct && !value.IsNil() {
				s.Fields = append(s.Fields, l.parseStruct(value.Elem()).Fields...)
				continue
			}
		}

		// empty tag or no tag, skip the field
		if tag == "" {
			continue
		}

		f := FieldConfig{
			Name:  field.Name,
			Key:   tag,
			Value: value,
		}

		// copying field content to a new value
		clone := reflect.Indirect(reflect.New(f.Value.Type()))
		clone.Set(f.Value)
		f.Default = clone

		if idx := strings.Index(tag, ","); idx != -1 {
			f.Key = tag[:idx]
			opts := strings.Split(tag[idx+1:], ",")

			for _, opt := range opts {
				if opt == "required" {
					f.Required = true
					continue
				}

				if strings.HasPrefix(opt, "short=") {
					f.Short = opt[len("short="):]
					continue
				}

				if strings.HasPrefix(opt, "description=") {
					f.Description = opt[len("description="):]
					continue
				}

				if strings.HasPrefix(opt, "backend=") {
					f.Backend = opt[len("backend="):]
				}
			}
		}

		s.Fields = append(s.Fields, &f)
	}

	return &s
}

func (l *Loader) resolve(ctx context.Context, s *StructConfig) error {
	for _, f := range s.Fields {
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

	foundFields := make(map[*FieldConfig]bool)

	for _, b := range l.backends {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if len(foundFields) == len(s.Fields) {
			break
		}

		if u, ok := b.(Unmarshaler); ok {
			err := u.Unmarshal(ctx, s.S)
			if err != nil {
				if err == backend.ErrNotFound {
					continue
				}
				return err
			}

			continue
		}

		if u, ok := b.(StructLoader); ok {
			err := u.LoadStruct(ctx, s)
			if err != nil {
				return err
			}

			continue
		}

		for _, f := range s.Fields {
			if _, ok := foundFields[f]; ok {
				continue
			}

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
			foundFields[f] = true
		}
	}

	for _, f := range s.Fields {
		if f.Required && isZero(f.Value) {
			return fmt.Errorf("required key '%s' for field '%s' not found", f.Key, f.Name)
		}
	}

	return nil
}

// StructConfig holds informations about each field of a struct S.
type StructConfig struct {
	S      interface{}
	Fields []*FieldConfig
}

// FieldConfig holds informations about a struct field.
type FieldConfig struct {
	Name        string
	Short       string
	Key         string
	Description string
	Value       reflect.Value
	Default     reflect.Value
	Required    bool
	Backend     string
}

// Set converts data into f.Value.
func (f *FieldConfig) Set(data string) error {
	return convert(data, f.Value)
}

var durationType = reflect.TypeOf(time.Duration(0))

func convert(data string, value reflect.Value) error {
	t := value.Type()
	if t == durationType {
		d, err := time.ParseDuration(data)
		if err != nil {
			return err
		}
		value.SetInt(int64(d))
		return nil
	}
	switch t.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(data)
		if err != nil {
			return err
		}
		value.SetBool(b)
	case reflect.Slice:
		var err error
		// create a new temporary slice to override the actual Value if it's not empty
		nv := reflect.MakeSlice(value.Type(), 0, 0)
		ss := strings.Split(data, ",")
		for _, s := range ss {
			// create a new Value v based on the type of the slice
			v := reflect.Indirect(reflect.New(t.Elem()))
			// call convert to set the current value of the slice to v
			err = convert(s, v)
			// append v to the temporary slice
			nv = reflect.Append(nv, v)
		}
		// Set the newly created temporary slice to the target Value
		value.Set(nv)
		return err
	case reflect.String:
		value.SetString(data)
	case reflect.Ptr:
		n := reflect.New(value.Type().Elem())
		value.Set(n)
		return convert(data, n.Elem())
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		i, err := strconv.ParseInt(data, 10, t.Bits())
		if err != nil {
			return err
		}

		value.SetInt(i)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		i, err := strconv.ParseUint(data, 10, t.Bits())
		if err != nil {
			return err
		}

		value.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(data, t.Bits())
		if err != nil {
			return err
		}
		value.SetFloat(f)
	default:
		return fmt.Errorf("field type '%s' not supported", t.Kind())
	}

	return nil
}

func isZero(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	current := v.Interface()
	return reflect.DeepEqual(current, zero)
}
