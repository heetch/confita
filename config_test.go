package confita_test

import (
	"context"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/heetch/confita"
	"github.com/stretchr/testify/require"
)

type store map[string]string

func (s store) Get(ctx context.Context, key string) ([]byte, error) {
	data, ok := s[key]
	if !ok {
		return nil, confita.ErrNotFound
	}

	return []byte(data), nil
}

func TestLoad(t *testing.T) {
	type nested struct {
		Int    int    `config:"int"`
		String string `config:"string"`
	}

	type testStruct struct {
		Bool            bool          `config:"bool"`
		Int             int           `config:"int"`
		Int8            int8          `config:"int8"`
		Int16           int16         `config:"int16"`
		Int32           int32         `config:"int32"`
		Int64           int64         `config:"int64"`
		Uint            uint          `config:"uint"`
		Uint8           uint8         `config:"uint8"`
		Uint16          uint16        `config:"uint16"`
		Uint32          uint32        `config:"uint32"`
		Uint64          uint64        `config:"uint64"`
		Float32         float32       `config:"float32"`
		Float64         float64       `config:"float64"`
		Ptr             *string       `config:"ptr"`
		String          string        `config:"string"`
		Duration        time.Duration `config:"duration"`
		Struct          nested
		StructPtrNil    *nested
		StructPtrNotNil *nested
		Ignored         string
	}

	var s testStruct
	s.StructPtrNotNil = new(nested)

	boolStore := store{
		"bool": "true",
	}

	intStore := store{
		"int":   strconv.FormatInt(math.MaxInt64, 10),
		"int8":  strconv.FormatInt(math.MaxInt8, 10),
		"int16": strconv.FormatInt(math.MaxInt16, 10),
		"int32": strconv.FormatInt(math.MaxInt32, 10),
		"int64": strconv.FormatInt(math.MaxInt64, 10),
	}

	uintStore := store{
		"uint":   strconv.FormatUint(math.MaxUint64, 10),
		"uint8":  strconv.FormatUint(math.MaxUint8, 10),
		"uint16": strconv.FormatUint(math.MaxUint16, 10),
		"uint32": strconv.FormatUint(math.MaxUint32, 10),
		"uint64": strconv.FormatUint(math.MaxUint64, 10),
	}

	floatStore := store{
		"float32": strconv.FormatFloat(math.MaxFloat32, 'f', 6, 32),
		"float64": strconv.FormatFloat(math.MaxFloat64, 'f', 6, 64),
	}

	otherStore := store{
		"ptr":      "ptr",
		"string":   "string",
		"duration": "10s",
	}

	loader := confita.NewLoader(confita.Backends(
		boolStore,
		intStore,
		uintStore,
		floatStore,
		otherStore))
	err := loader.Load(&s)
	require.NoError(t, err)

	ptr := "ptr"
	require.EqualValues(t, testStruct{
		Bool:     true,
		Int:      math.MaxInt64,
		Int8:     math.MaxInt8,
		Int16:    math.MaxInt16,
		Int32:    math.MaxInt32,
		Int64:    math.MaxInt64,
		Uint:     math.MaxUint64,
		Uint8:    math.MaxUint8,
		Uint16:   math.MaxUint16,
		Uint32:   math.MaxUint32,
		Uint64:   math.MaxUint64,
		Float32:  math.MaxFloat32,
		Float64:  math.MaxFloat64,
		Ptr:      &ptr,
		String:   "string",
		Duration: 10 * time.Second,
		Struct: nested{
			Int:    math.MaxInt64,
			String: "string",
		},
		StructPtrNotNil: &nested{
			Int:    math.MaxInt64,
			String: "string",
		},
	}, s)
}
