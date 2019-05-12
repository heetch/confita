package confita_test

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/heetch/confita"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type store map[string]string

func (s store) Unmarshal(ctx context.Context, key string, to interface{}) error {
	data, ok := s[key]
	if !ok {
		return confita.ErrNotFound
	}
	return confita.Unmarshal(data, to)
}

func (store) UseFieldNameKey() bool {
	return false
}

type longRunningStore time.Duration

func (s longRunningStore) Unmarshal(ctx context.Context, key string, to interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(s)):
		return confita.ErrNotFound
	}
}

func (longRunningStore) UseFieldNameKey() bool {
	return false
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
		String          string        `config:"string"`
		Duration        time.Duration `config:"duration"`
		Struct          nested
		StructPtrNil    *nested
		StructPtrNotNil *nested
		Ignored         string
		unexported      int `config:"ignore"`
	}

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

	loader := confita.NewLoader(
		boolStore,
		intStore,
		uintStore,
		floatStore,
		otherStore,
	)

	var s testStruct
	s.StructPtrNotNil = new(nested)
	err := loader.Load(context.Background(), &s)
	require.NoError(t, err)

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

func TestLoadRequired(t *testing.T) {
	s := struct {
		Name string `config:"name,required"`
	}{}

	st := make(store)
	err := confita.NewLoader(st).Load(context.Background(), &s)
	require.Error(t, err)
}

func TestLoadIgnored(t *testing.T) {
	s := struct {
		Name string `config:"-"`
		Age  int    `config:"age"`
	}{}

	st := store{
		"name": "name",
		"age":  "10",
	}

	err := confita.NewLoader(st).Load(context.Background(), &s)
	require.NoError(t, err)
	require.Equal(t, 10, s.Age)
	require.Zero(t, s.Name)
}

func TestLoadContextCancel(t *testing.T) {
	t.Skip("is it really worth checking context in non-blocking code?")
	s := struct {
		Name string `config:"-"`
		Age  int    `config:"age"`
	}{}

	st := store{
		"name": "name",
		"age":  "10",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := confita.NewLoader(st).Load(ctx, &s)
	require.Equal(t, context.Canceled, err)
}

func TestLoadContextTimeout(t *testing.T) {
	t.Skip("is it really worth checking context in non-blocking code?")
	s := struct {
		Name string `config:"-"`
		Age  int    `config:"age"`
	}{}

	st := longRunningStore(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := confita.NewLoader(st).Load(ctx, &s)
	require.Equal(t, context.DeadlineExceeded, err)
}

type backendMock struct {
	useFieldNameKey bool
	store           map[string]string
	called          int
	name            string
}

func (b *backendMock) Unmarshal(ctx context.Context, key string, to interface{}) error {
	b.called++
	data, ok := b.store[key]
	if !ok {
		return confita.ErrNotFound
	}
	return confita.Unmarshal(data, to)
}

func (b *backendMock) UseFieldNameKey() bool {
	return b.useFieldNameKey
}

func TestTags(t *testing.T) {
	t.Run("BadRequired", func(t *testing.T) {
		type test struct {
			Key string `config:"key,rrequiredd,backend=store"`
		}

		myStore := make(store)
		myStore["oups"] = "value"

		ldr := confita.NewLoader(myStore)

		var cfg test
		err := ldr.Load(context.Background(), &cfg)
		require.NoError(t, err)

		assert.Equal(t, "", cfg.Key)
	})

	t.Run("BadTagsOrder", func(t *testing.T) {
		type test struct {
			Key string `config:"backend=store,key"`
		}

		myStore := make(store)
		myStore["key"] = "value"

		ldr := confita.NewLoader(myStore)

		var cfg test
		err := ldr.Load(context.Background(), &cfg)
		require.NoError(t, err)

		t.Skip("why should cfg.Key be set by the above code? Surely the field with the bad tags should be ignored (or yield an error)?")
		assert.Equal(t, "", cfg.Key)
	})
}

func TestSliceField(t *testing.T) {
	t.Run("Slice of string - empty", func(t *testing.T) {
		s := struct {
			Letters []string `config:"letters"`
		}{}

		st := store{
			"letters": "a,b,c",
		}
		e := []string{
			"a",
			"b",
			"c",
		}
		err := confita.NewLoader(st).Load(context.Background(), &s)
		require.NoError(t, err)
		require.EqualValues(t, e, s.Letters)
	})

	t.Run("Slice of string - non-empty - no appending", func(t *testing.T) {

		s := struct {
			Letters []string `config:"letters"`
		}{
			Letters: []string{"a", "b"},
		}

		st := store{
			"letters": "c,d,e",
		}
		e := []string{
			"c",
			"d",
			"e",
		}
		err := confita.NewLoader(st).Load(context.Background(), &s)
		require.NoError(t, err)
		require.EqualValues(t, e, s.Letters)
	})

	t.Run("Slice of int", func(t *testing.T) {
		s := struct {
			Numbers []int `config:"numbers"`
		}{}

		st := store{
			"numbers": "21,21,42",
		}
		e := []int{
			21,
			21,
			42,
		}
		err := confita.NewLoader(st).Load(context.Background(), &s)
		require.NoError(t, err)
		require.EqualValues(t, e, s.Numbers)
	})

	t.Run("Slice of *int", func(t *testing.T) {
		t.Skip("simpler just not to support pointers?")
		s := struct {
			Numbers []*int `config:"numbers"`
		}{}

		st := store{
			"numbers": "21,21,42",
		}

		err := confita.NewLoader(st).Load(context.Background(), &s)
		require.NoError(t, err)
		require.Equal(t, 21, *s.Numbers[0])
		require.Equal(t, 21, *s.Numbers[1])
		require.Equal(t, 42, *s.Numbers[2])
	})
}

var errorTests = []struct {
	testName    string
	store       store
	into        interface{}
	expectError string
}{{
	testName: "bad-duration",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X time.Duration `config:"X"`
	}),
	expectError: `time: invalid duration xxxx`,
}, {
	testName: "bad-bool",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X bool `config:"X"`
	}),
	expectError: `strconv.ParseBool: parsing "xxxx": invalid syntax`,
}, {
	testName: "bad-int",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X int `config:"X"`
	}),
	expectError: `strconv.ParseInt: parsing "xxxx": invalid syntax`,
}, {
	testName: "bad-uint",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X uint `config:"X"`
	}),
	expectError: `strconv.ParseUint: parsing "xxxx": invalid syntax`,
}, {
	testName: "out-of-range-int",
	store: store{
		"X": "128",
	},
	into: new(struct {
		X int8 `config:"X"`
	}),
	expectError: `strconv.ParseInt: parsing "128": value out of range`,
}, {
	testName: "out-of-range-uint",
	store: store{
		"X": "256",
	},
	into: new(struct {
		X uint8 `config:"X"`
	}),
	expectError: `strconv.ParseUint: parsing "256": value out of range`,
}, {
	testName: "bad-float",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X float64 `config:"X"`
	}),
	expectError: `strconv.ParseFloat: parsing "xxxx": invalid syntax`,
}, {
	testName: "unsupported-field-type",
	store: store{
		"X": "xxxx",
	},
	into: new(struct {
		X uintptr `config:"X"`
	}),
	expectError: `field type 'uintptr' not supported`,
}, {
	testName:    "not-struct-pointer",
	into:        struct{}{},
	expectError: `provided target must be a pointer to struct`,
}}

func TestError(t *testing.T) {
	for _, test := range errorTests {
		t.Run(test.testName, func(t *testing.T) {
			err := confita.NewLoader(test.store).Load(context.Background(), test.into)
			if err == nil {
				t.Fatalf("unexpected success; into %#v, want error matching %q", test.into, test.expectError)
			}
			if ok, err1 := regexp.MatchString("^("+test.expectError+")$", err.Error()); !ok || err1 != nil {
				t.Fatalf("error mismatch; got %q want %q", err, test.expectError)
			}
		})
	}
}
