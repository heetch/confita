package confita_test

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type store map[string]string

func (s store) Get(ctx context.Context, key string) ([]byte, error) {
	data, ok := s[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return []byte(data), nil
}

func (store) Name() string {
	return "store"
}

type longRunningStore time.Duration

func (s longRunningStore) Get(ctx context.Context, key string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(s)):
		return []byte(time.Now().String()), nil
	}
}

func (longRunningStore) Name() string {
	return "longRunningStore"
}

type unmarshaler []byte

func (u unmarshaler) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (u unmarshaler) Unmarshal(ctx context.Context, to interface{}) error {
	return json.Unmarshal([]byte(u), to)
}

func (unmarshaler) Name() string {
	return "unmarshaler"
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
		unexported      int `config:"ignore"`
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

	loader := confita.NewLoader(
		boolStore,
		intStore,
		uintStore,
		floatStore,
		otherStore)
	err := loader.Load(context.Background(), &s)
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

func TestLoadFromUnmarshaler(t *testing.T) {
	s := struct {
		Name    string `config:"name"`
		Age     int    `config:"age"`
		Ignored string `config:"-"`
	}{}

	st := unmarshaler(`{
		"name": "name",
		"age":  10
	}`)

	err := confita.NewLoader(st).Load(context.Background(), &s)
	require.NoError(t, err)
	require.Equal(t, "name", s.Name)
	require.Equal(t, 10, s.Age)
	require.Zero(t, s.Ignored)
}

func TestLoadFromStructLoader(t *testing.T) {
	s := struct {
		Name    string `config:"name"`
		Age     int    `config:"age"`
		Ignored string `config:"-"`
	}{}

	sl := structLoader{store{
		"name": "name",
		"age":  "10",
	}}

	err := confita.NewLoader(&sl).Load(context.Background(), &s)
	require.NoError(t, err)
	require.Equal(t, "name", s.Name)
	require.Equal(t, 10, s.Age)
	require.Zero(t, s.Ignored)
}

type backendMock struct {
	store  map[string]string
	called int
	name   string
}

func (b *backendMock) Get(ctx context.Context, key string) ([]byte, error) {
	b.called++
	data, ok := b.store[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return []byte(data), nil
}

func (b *backendMock) Name() string {
	return b.name
}

type structLoader struct {
	store
}

func (s *structLoader) LoadStruct(ctx context.Context, cfg *confita.StructConfig) error {
	for _, f := range cfg.Fields {
		v, err := s.Get(ctx, f.Key)
		if err != nil {
			return err
		}

		err = f.Set(string(v))
		if err != nil {
			return err
		}
	}

	return nil
}

func TestBackendTag(t *testing.T) {
	type test struct {
		Tikka  string `config:"tikka,backend=store"`
		Cheese string `config:"cheese,required,backend=backendCalled"`
	}

	backendNotCalled := &backendMock{
		store: make(map[string]string),
		name:  "backendNotCalled",
	}
	backendNotCalled.store["cheese"] = "nan"

	myStore := make(store)
	myStore["tikka"] = "masala"

	t.Run("OK", func(t *testing.T) {
		backendCalled := &backendMock{
			store: make(map[string]string),
			name:  "backendCalled",
		}
		backendCalled.store["cheese"] = "nan"

		ldr := confita.NewLoader(myStore, backendCalled, backendNotCalled)

		var cfg test
		err := ldr.Load(context.Background(), &cfg)
		require.NoError(t, err)

		assert.Equal(t, "nan", cfg.Cheese)
		assert.Equal(t, "masala", cfg.Tikka)
		assert.Equal(t, 1, backendCalled.called)
		assert.Equal(t, 0, backendNotCalled.called)
	})

	t.Run("NOK", func(t *testing.T) {
		backendCalled := &backendMock{
			store: make(map[string]string),
			name:  "backendCalled",
		}

		ldr := confita.NewLoader(myStore, backendCalled, backendNotCalled)

		var cfg test
		err := ldr.Load(context.Background(), &cfg)
		require.EqualError(t, err, "required key 'cheese' for field 'Cheese' not found")

		assert.Equal(t, 1, backendCalled.called)
		assert.Equal(t, 0, backendNotCalled.called)
	})
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

	t.Run("BadBackendValue", func(t *testing.T) {
		type test struct {
			Key string `config:"key,backend=stor"`
		}

		myStore := make(store)
		myStore["key"] = "value"

		ldr := confita.NewLoader(myStore)

		var cfg test
		err := ldr.Load(context.Background(), &cfg)
		require.Error(t, err)
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

		assert.Equal(t, "", cfg.Key)
	})
}

func TestCustomTag(t *testing.T) {
	s := struct {
		Name string `custom:"name"`
		Age  int    `custom:"age"`
	}{}

	st := store{
		"name": "name",
		"age":  "10",
	}

	err := confita.NewLoader(st).Load(context.Background(), &s)
	require.NoError(t, err)
	require.Empty(t, &s)

	l := confita.NewLoader(st)
	l.Tag = "custom"
	err = l.Load(context.Background(), &s)
	require.NoError(t, err)
	require.Equal(t, "name", s.Name)
	require.Equal(t, 10, s.Age)
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
				t.Fatalf("unexpected success; into %#v", test.into)
			}
			if ok, err1 := regexp.MatchString("^("+test.expectError+")$", err.Error()); !ok || err1 != nil {
				t.Fatalf("error mismatch; got %q want %q", err, test.expectError)
			}
		})
	}
}
