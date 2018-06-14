package flags

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/heetch/confita"
	"github.com/stretchr/testify/require"
)

func TestFlags(t *testing.T) {
	var a int

	r := reflect.Indirect(reflect.ValueOf(&a))

	s := confita.StructConfig{Fields: []*confita.FieldConfig{
		&confita.FieldConfig{Key: "some-int", Value: &r},
	}}

	os.Args = append(os.Args, "-some-int", "10")

	b := new(Backend)
	err := b.LoadStruct(context.Background(), &s)
	require.NoError(t, err)
	require.Equal(t, 10, a)
}
