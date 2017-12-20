package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	// ErrNotFound is returned by the backends or by the Load function when the given key is not found.
	ErrNotFound = errors.New("key not found")
)

// Load configuration from any backend.
func Load(to interface{}, backends ...Backend) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return errors.New("provided target must be a pointer to struct")
	}

	ref = ref.Elem()
	t := ref.Type()

	numFields := ref.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		value := ref.Field(i)

		if field.PkgPath != "" {
			continue
		}

		key := field.Tag.Get("config")
		if key == "" {
			continue
		}

		var found bool
		for _, b := range backends {
			raw, err := b.Get(key)
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

		if !found {
			return ErrNotFound
		}
	}

	return nil
}

// A Backend is used to fetch values from a given key.
type Backend interface {
	Get(key string) ([]byte, error)
}

func convert(data string, value *reflect.Value) error {
	k := value.Kind()

	switch {
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
