package flags

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

type logLevel int

var (
	logLevelDebug logLevel = 1
	logLevelInfo  logLevel = 2
)

func (l *logLevel) Set(val string) error {
	switch val {
	case "debug":
		*l = logLevelDebug
	case "info":
		*l = logLevelInfo
	default:
		return fmt.Errorf("unknown log level: %s", val)
	}
	return nil
}

func (l logLevel) String() string {
	switch l {
	case logLevelDebug:
		return "debug"
	case logLevelInfo:
		return "info"
	default:
		return "<unknown>"
	}
}

func runHelper(t *testing.T, cfg interface{}, args ...string) {
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
			G logLevel      `config:"g"`
		}
		var cfg config
		runHelper(t, &cfg, "-a=hello", "-b=true", "-c=10s", "-d=-100", "-e=1", "-f=100.01", "-g=info")
		require.Equal(t, config{
			A: "hello",
			B: true,
			C: 10 * time.Second,
			D: -100,
			E: 1,
			F: 100.01,
			G: logLevelInfo,
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
			Gdef logLevel      `config:"g-def,short=gd"`
		}
		cfg := &config{
			Adef: "hello",
			Bdef: true,
			Cdef: 10 * time.Second,
			Ddef: -100,
			Gdef: logLevelInfo,
		}
		runHelper(t, cfg, "-a-def=bye", "-b-def=false", "-c-def=15s", "-d-def=-200", "-e-def=400", "-f-def=2.33", "-g-def=debug")

		require.Equal(t, &config{
			Adef: "bye",
			Bdef: false,
			Cdef: 15 * time.Second,
			Ddef: -200,
			Edef: 400,
			Fdef: 2.33,
			Gdef: logLevelDebug,
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
	}
	cfg := &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 10 * time.Second,
		Ddef: -100,
	}
	runHelper(t, cfg, "-ad=hello", "-bd=true", "-cd=20s", "-dd=500", "-ed=700", "-fd=333.33")
	require.Equal(t, &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 20 * time.Second,
		Ddef: 500,
		Edef: 700,
		Fdef: 333.33,
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
	}
	cfg := &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 10 * time.Second,
		Ddef: -100,
	}
	runHelper(t, cfg, "-ad=hello", "-b-def=true", "-cd=20s", "-d-def=500", "-ed=600", "-f-def=42.42")
	require.Equal(t, &config{
		Adef: "hello",
		Bdef: true,
		Cdef: 20 * time.Second,
		Ddef: 500,
		Edef: 600,
		Fdef: 42.42,
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
	}

	var cfg config

	st := store{
		"bool":     "true",
		"uint":     "42",
		"float":    "42.42",
		"duration": "1ns",
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
