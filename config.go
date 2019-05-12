// Package confita allows multiple packages to obtain configuration from an open-ended
// set of possible data sources (backends).
package confita

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("configuration value not found")

// Loader can be used to load configuration information
// from a variety of sources.
type Loader struct {
	backends []Backend
}

// Backend provides a way of getting values from a particular
// configuration source.
type Backend interface {
	// UseFieldNameKey reports whether the backend expects the key passed
	// to Unmarshal to be the field name, not the name specified in the field tag.
	// Most backends should report false for this - it's here
	// for the benefit of supporting legacy behaviour in the file backend.
	UseFieldNameKey() bool

	// Get unmarshals the value associated with the given key
	// into the value pointed to by to, which will
	// be one of the types as documented in Unmarshal.
	//
	// If there is no value for the key, Get should return
	// ErrNotFound.
	//
	// This method may be called concurrently.
	Unmarshal(ctx context.Context, key string, to interface{}) error
}

// Environ implements Backend by fetching values
// from environment variables.
var Environ = environBackend{}

type environBackend struct{}

// Unmarshal implements Backend.Unmarshal.
func (environBackend) Unmarshal(ctx context.Context, key string, to interface{}) error {
	return Unmarshal(key, to)
}

// UseFieldNameKey implements Backend.UseFieldNameKey.
func (environBackend) UseFieldNameKey() bool {
	return false
}

// NewLoader creates a new Loader instance that will use all the given
// backends for configuration information. Later backends
// in the slice take precedence over earlier ones.
//
// If no backends are specified, Environ will be used.
func NewLoader(backends ...Backend) *Loader {
	if len(backends) == 0 {
		backends = []Backend{Environ}
	}
	return &Loader{
		backends: backends,
	}
}

// Load loads configuration information into the given value, which must
// be a pointer to a struct.
//
// It inspects each exported field of the struct. If a field has a "config" tag
// or is itself a struct or pointer to a struct, it will be considered for
// loading.
//
// If the field has a "config" tag, its value will be fetched from all
// backends in sequence, with each backend potentially overriding the
// previous one. The key used to fetch the value will be the lower-cased
// field name, except when using the Environ provider, which will use
// the value of the config tag.
//
// The special tag value "-" can be used to suppress processing of a
// field even if that field is a struct or pointer to struct.
//
// If the field is a struct or a pointer to a struct, the fields of that struct will
// be considered by Load recursively.
//
// The following type kinds are supported by fields with the "config"
// tag:
//
//    bool
//    string
//    time.Duration
//    numeric types except complex numbers
//    a slice of any of the above
func (l *Loader) Load(ctx context.Context, to interface{}) error {
	v := reflect.ValueOf(to)
	t := v.Type()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("provided target must be a pointer to struct")
	}
	return l.loadStruct(ctx, v.Elem())
}

func (l *Loader) loadStruct(ctx context.Context, v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			// TODO support embedding
			return errors.New("anonymous fields not supported")
		}
		if field.PkgPath != "" {
			// Field is not exported; skip it.
			continue
		}
		tag := field.Tag.Get("config")
		if tag == "-" {
			continue
		}
		fieldv := v.Field(i)
		fieldt := field.Type
		// If struct or *struct, parse recursively
		switch fieldt.Kind() {
		case reflect.Struct:
			if err := l.loadStruct(ctx, fieldv); err != nil {
				return err
			}
		case reflect.Ptr:
			if fieldv.IsNil() || fieldt.Elem().Kind() != reflect.Struct {
				break
			}
			if err := l.loadStruct(ctx, fieldv.Elem()); err != nil {
				return err
			}
		default:
			if tag == "" {
				break
			}
			if err := l.loadField(ctx, fieldv, field, tag); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

// loadField loads the value from the given value, which is described as the
// given field, where tag is the confita-specific struct field tag parsed from field.Tag.
// The value is loaded by querying each backend in turn.
func (l *Loader) loadField(ctx context.Context, v reflect.Value, field reflect.StructField, tag string) error {
	key, required := parseTag(tag)
	if key == "" {
		return fmt.Errorf("no config name for %v", field.Name)
	}
	// Go through backends in reverse order, stopping at the first
	// one that returns an answer, which might save a round-trip or two.
	// TODO fetch values concurrently.
	for i := len(l.backends) - 1; i >= 0; i-- {
		b := l.backends[i]
		// Avoid information pollution in backends by giving the
		// backend a fresh zero instance of the result type
		// instead of the struct field directly. This will
		// enable us to use concurrent fetches without being
		// worried about breaking potential cases where backends
		// end up depending on values that other backends set.
		// Also, it means that if Unmarshal returns nil, we'll
		// still zero the result, something that the Unmarshal
		// implementation should do anyway, but the guarantee is
		// nice.
		dest := reflect.New(field.Type)

		backendKey := key
		if b.UseFieldNameKey() {
			backendKey = field.Name
		}
		if err := b.Unmarshal(ctx, backendKey, dest.Interface()); err != nil {
			if errors.Cause(err) != ErrNotFound {
				return errors.WithStack(err)
			}
			continue
		}
		v.Set(dest.Elem())
	}
	if required && isZero(v) {
		return fmt.Errorf("required key %q for field %q not found", key, field.Name)
	}
	return nil
}

func parseTag(tag string) (envVar string, required bool) {
	tagFields := strings.Split(tag, ",")
	envVar = tagFields[0]
	for _, tagf := range tagFields[1:] {
		if tagf == "required" {
			required = true
		}
	}
	return envVar, required
}

var durationType = reflect.TypeOf(time.Duration(0))

// Unmarshal unmarshals the given string value into to,
// which must be a pointer to one of the types documented
// for config fields in Load.
//
// To unmarshal into a slice, the value is split into
// comma-separated fields and then invoking
// Unmarshal on each resulting field.
//
// Duration values are unmarshaled using time.ParseDuration.
func Unmarshal(val string, to interface{}) error {
	v := reflect.ValueOf(to)
	if !v.IsValid() || v.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unmarshal into %T", to)
	}
	return unmarshal(val, v.Elem())
}

// unmarshal is the internal version of unmarshal which
// unmarshals data into a reflect.Value instead of interface{}.
// The value v is expected to be addressable.
func unmarshal(data string, v reflect.Value) error {
	t := v.Type()
	if t == durationType {
		d, err := time.ParseDuration(data)
		if err != nil {
			return err
		}
		v.SetInt(int64(d))
		return nil
	}
	switch t.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(data)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Slice:
		// Unmarshal a slice by splitting the value into comma-separated fields.
		ss := strings.Split(data, ",")
		nv := reflect.MakeSlice(v.Type(), len(ss), len(ss))
		for i, s := range ss {
			if err := unmarshal(s, nv.Index(i)); err != nil {
				return fmt.Errorf("cannot unmarshal %q into %s", s, v.Type().Elem())
			}
		}
		v.Set(nv)
		return nil
	case reflect.String:
		v.SetString(data)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		i, err := strconv.ParseInt(data, 10, t.Bits())
		if err != nil {
			return err
		}

		v.SetInt(i)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		i, err := strconv.ParseUint(data, 10, t.Bits())
		if err != nil {
			return err
		}

		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(data, t.Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("field type '%s' not supported", t.Kind())
	}

	return nil
}

// isZero reports whether v is the zero value for its type.
func isZero(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(v.Interface(), zero)
}
