package flags

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
)

func runHelper(t *testing.T, cfg any, args ...string) {
	t.Helper()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = append([]string{"a.out"}, args...)
	err := confita.NewLoader(&Backend{flags}).Load(context.Background(), cfg)
	require.NoError(t, err)
}

func TestFlags(t *testing.T) {
	t.Run("Use defaults", func(t *testing.T) {
		type config struct {
			A string        `config:"a"`
			B bool          `config:"b"`
			C time.Duration `config:"c"`
			D int           `config:"d"`
			E uint          `config:"e"`
			F float32       `config:"f"`
			G time.Time     `config:"g"`
			H *time.Time    `config:"h"`
		}
		var cfg config
		runHelper(t, &cfg, "-a=hello", "-b=true", "-c=10s", "-d=-100", "-e=1", "-f=100.01", "-g=2001-01-01T01:01:01Z", "-h=2001-01-01T01:01:01Z")
		expTime := time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC)
		require.Equal(t, config{
			A: "hello",
			B: true,
			C: 10 * time.Second,
			D: -100,
			E: 1,
			F: 100.01,
			G: expTime,
			H: &expTime,
		}, cfg)
	})

	t.Run("Override defaults", func(t *testing.T) {
		type config struct {
			Adef string        `config:"a-def,short=ad"`
			Bdef bool          `config:"b-def,short=bd"`
			Cdef time.Duration `config:"c-def,short=cd"`
			Ddef int           `config:"d-def,short=dd"`
			Edef uint          `config:"e-def,short=ed"`
			Fdef float32       `config:"f-def,short=fd"`
			Gdef time.Time     `config:"g-def,short=gd"`
			// Hdef *time.Time    `config:"h-def,short=hd"`
		}
		defaultTime := time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)
		cfg := &config{
			Adef: "hello",
			Bdef: true,
			Cdef: 10 * time.Second,
			Ddef: -100,
			Edef: 200,
			Fdef: 1.23,
			Gdef: defaultTime,
			// Hdef: &defaultTime,
		}
		runHelper(t, cfg, "-a-def=bye", "-b-def=false", "-c-def=15s", "-d-def=-200", "-e-def=400", "-f-def=2.33", "-g-def=2001-01-01T01:01:01Z" /*, "-h-def=2001-01-01T01:01:01Z"*/)
		expTime := time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC)
		require.Equal(t, &config{
			Adef: "bye",
			Bdef: false,
			Cdef: 15 * time.Second,
			Ddef: -200,
			Edef: 400,
			Fdef: 2.33,
			Gdef: expTime,
			// Hdef: &expTime,
		}, cfg)
	})
}

func TestFlagsShort(t *testing.T) {
	type config struct {
		Adef string        `config:"a-def,short=ad"`
		Bdef bool          `config:"b-def,short=bd"`
		Cdef time.Duration `config:"c-def,short=cd"`
		Ddef int           `config:"d-def,short=dd"`
		Edef uint          `config:"e-def,short=ed"`
		Fdef float32       `config:"f-def,short=fd"`
		Gdef time.Time     `config:"g-def,short=gd"`
		// Hdef *time.Time    `config:"h-def,short=hd"`
	}
	defaultTime := time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)
	cfg := &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 10 * time.Second,
		Ddef: -100,
		Edef: 200,
		Fdef: 1.23,
		Gdef: defaultTime,
		// Hdef: &defaultTime,
	}
	runHelper(t, cfg, "-ad=hello", "-bd=true", "-cd=20s", "-dd=500", "-ed=700", "-fd=333.33", "-gd=2001-01-01T01:01:01Z" /*"-hd=2001-01-01T01:01:01Z"*/)
	expTime := time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC)
	require.Equal(t, &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 20 * time.Second,
		Ddef: 500,
		Edef: 700,
		Fdef: 333.33,
		Gdef: expTime,
		// Hdef: &expTime,
	}, cfg)
}

func TestFlagsMixed(t *testing.T) {
	type config struct {
		Adef string        `config:"a-def,short=ad"`
		Bdef bool          `config:"b-def,short=bd"`
		Cdef time.Duration `config:"c-def,short=cd"`
		Ddef int           `config:"d-def,short=dd"`
		Edef uint          `config:"e-def,short=ed"`
		Fdef float32       `config:"f-def,short=fd"`
		Gdef time.Time     `config:"g-def,short=gd"`
		// Hdef *time.Time    `config:"h-def,short=hd"`
	}
	defaultTime := time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)
	cfg := &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 10 * time.Second,
		Ddef: -100,
		Edef: 200,
		Fdef: 1.23,
		Gdef: defaultTime,
		// Hdef: &defaultTime,
	}
	runHelper(t, cfg, "-ad=hello", "-b-def=true", "-cd=20s", "-d-def=500", "-ed=600", "-f-def=42.42", "-gd=2001-01-01T01:01:01Z" /*"-hd=2001-01-01T01:01:01Z"*/)
	expTime := time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC)
	require.Equal(t, &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 20 * time.Second,
		Ddef: 500,
		Edef: 600,
		Fdef: 42.42,
		Gdef: expTime,
		// Hdef: &expTime,
	}, cfg)
}

func TestWithAnotherBackend(t *testing.T) {
	type config struct {
		String   string        `config:"string,required"`
		Bool     bool          `config:"bool,required"`
		Int      int           `config:"int,required"`
		Uint     uint          `config:"uint,required"`
		Float    float64       `config:"float,required"`
		Duration time.Duration `config:"duration,required"`
		Time     time.Time     `config:"time,required"`
		TimePtr  *time.Time    `config:"time_ptr,required"`
	}

	var cfg config

	st := store{
		"bool":     "true",
		"uint":     "42",
		"float":    "42.42",
		"duration": "1ns",
		"time":     "2001-01-01T01:01:01Z",
		"time_ptr": "2001-01-01T01:01:01Z",
	}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = append([]string{"a.out"}, "-int=42", "-string=string", "-float=99.5")
	err := confita.NewLoader(st, &Backend{flags}).Load(context.Background(), &cfg)
	require.NoError(t, err)

	require.Equal(t, "string", cfg.String)
	require.Equal(t, true, cfg.Bool)
	require.Equal(t, 42, cfg.Int)
	require.EqualValues(t, 42, cfg.Uint)
	require.Equal(t, 99.5, cfg.Float)
	require.Equal(t, time.Duration(1), cfg.Duration)
	require.Equal(t, time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC), cfg.Time)
	require.Equal(t, time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC), *cfg.TimePtr)
}

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
