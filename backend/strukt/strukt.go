package strukt

import (
	"context"
	"fmt"
	"github.com/heetch/confita/backend"
	"reflect"
	"strconv"
)

type Backend struct {
	v   reflect.Value
	err error
}

// NewBackend creates a configuration loader that loads from a struct.
func NewBackend(s interface{}) *Backend {
	b := Backend{
		v: reflect.ValueOf(s),
	}

	if b.v.Kind() != reflect.Struct {
		b.err = fmt.Errorf("strukt backend expected a struct but a %s was given", b.v.Type())
	}

	return &b
}

// Get retrieves the field from the previously given struct with matching key
func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}

	val := b.v.FieldByName(key)
	for val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	var str string
	switch val.Kind() {
	case reflect.Invalid:
		return nil, backend.ErrNotFound
	case reflect.String:
		str = val.String()
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:

		str = strconv.FormatInt(val.Int(), 10)
	case reflect.Bool:
		str = strconv.FormatBool(val.Bool())
	case reflect.Array, reflect.Slice:
		//TODO
	}

	return []byte(str), nil
}

func (b *Backend) Name() string {
	return "strukt"
}
