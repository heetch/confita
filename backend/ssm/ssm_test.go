package ssm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/mocks"
	"github.com/heetch/confita/ptr"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAWSError(t *testing.T) {
	client := new(mocks.SSMAPI)
	ssmOpts := getSSMOpts("/borked/")
	ctx := context.Background()
	expected := fmt.Errorf("aws down")
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		nil, expected)

	b := NewBackend(client, "/borked/")
	_, actual := b.Get(context.Background(), "some_key")
	require.Equal(t, expected, actual)
}

func TestKeyNotFound(t *testing.T) {
	client := new(mocks.SSMAPI)
	ssmOpts := getSSMOpts("/whatevs/")
	ctx := context.Background()
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		&ssm.GetParametersByPathOutput{}, nil)

	b := NewBackend(client, "/whatevs/")
	_, actual := b.Get(context.Background(), "some_key")
	require.Equal(t, backend.ErrNotFound, actual)
}

func ptrString(str string) *string {
	return &str
}

func TestKeysFound(t *testing.T) {
	client := new(mocks.SSMAPI)
	ctx := context.Background()
	ssmOpts := getSSMOpts("/yo/whatup/")
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name:ptrString("/yo/whatup/a_key"), Value:ptrString("wow")},
				{Name:ptrString("/yo/whatup/some_key"), Value: ptrString("wondrous")},
			},
		}, nil)

	b := NewBackend(client, "/yo/whatup/")
	actual, err := b.Get(context.Background(), "a_key")
	require.Nil(t, err)
	require.Equal(t, "wow", string(actual))
	actual, err = b.Get(context.Background(), "some_key")
	require.Nil(t, err)
	require.Equal(t, "wondrous", string(actual))
}

func getSSMOpts(path string) *ssm.GetParametersByPathInput {
	return &ssm.GetParametersByPathInput{
		Path:             &path,
		Recursive:        ptr.Bool(true),
		WithDecryption:   ptr.Bool(true),
		MaxResults: ptr.Int64(10),
	}
}