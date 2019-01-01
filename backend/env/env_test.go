package env

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

func TestEnvBackend(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		b := NewBackend()

		_, err := b.Get(context.Background(), "something that doesn't exist")
		require.Equal(t, backend.ErrNotFound, err)
	})

	t.Run("ExactMatch", func(t *testing.T) {
		b := NewBackend()

		os.Setenv("TESTCONFIG1", "ok")
		val, err := b.Get(context.Background(), "TESTCONFIG1")
		require.NoError(t, err)
		require.Equal(t, "ok", string(val))
	})

	t.Run("DifferentCase", func(t *testing.T) {
		b := NewBackend()

		os.Setenv("TEST_CONFIG_2", "ok")
		val, err := b.Get(context.Background(), "test-config-2")
		require.NoError(t, err)
		require.Equal(t, "ok", string(val))
	})
}

func TestOpts(t *testing.T) {
	type args struct {
		fns []opt
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test opts positive",
			args: args{
				fns: []opt{
					ToUpper(),
					WithPrefix("TEST_"),
				},
				key: "var",
			},
			want: "TEST_VAR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := opts(tt.args.key, tt.args.fns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBackend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithPrefix(t *testing.T) {
	type args struct {
		preffix string
		key     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test add preffix positive",
			args: args{
				preffix: "TEST",
				key:     "VAR",
			},
			want: "TEST_VAR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPrefix(tt.args.preffix)(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToUpper(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test add preffix positive",
			args: args{
				"test",
			},
			want: "TEST",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpper()(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToUpper() = %v, want %v", got, tt.want)
			}
		})
	}
}
